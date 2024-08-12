package managers

import (
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func Test_ReadOnly_Mode(t *testing.T) {
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	assert.Equal(t, man.CanEdit(), true, "It should be able to edit")

	cfg.ReadOnly = true
	utils.SetCfg(*cfg)
	assert.Equal(t, man.CanEdit(), false, "We should not be able to edit now")
}

// A basic DB search (ilike matching)
func TestDbManagerSearch(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	assert.Equal(t, man.CanEdit(), true, "It should be a DB manager")

	cnt1, content1 := test_common.GetContentByDirName("dir1")
	cnt2, content2 := test_common.GetContentByDirName("dir2")
	c1_err := man.CreateContainer(cnt1)
	assert.NoError(t, c1_err)
	c2_err := man.CreateContainer(cnt2)
	assert.NoError(t, c2_err)

	cnts, count, s_err := man.SearchContainers(ContainerQuery{Search: "dir1", Page: 1, PerPage: 2})
	assert.NoError(t, s_err, "Searching for dir1 caused an error")
	assert.Equal(t, 1, len(*cnts), "We should only get one container back")
	assert.Equal(t, 1, int(count), "It should get the count right")

	for _, mc := range content1 {
		man.CreateContent(&mc)
	}
	for _, mc := range content2 {
		man.CreateContent(&mc)
		if mc.Src == test_common.VIDEO_FILENAME {
			man.CreateScreen(&models.Screen{ContentID: mc.ID, Src: "screen1"})
			man.CreateScreen(&models.Screen{ContentID: mc.ID, Src: "screen2"})
		}
	}

	sr := ContentQuery{Search: "Large", Page: 1, PerPage: 20}
	mcs, _, err := man.SearchContent(sr)
	assert.NoError(t, err, "It should be able to search")
	assert.NotNil(t, mcs, "It should be")
	assert.Equal(t, 3, len(*mcs), fmt.Sprintf("We should have 3 large images with an ilike %s", mcs))

	sr = ContentQuery{Search: "donut", Page: 1, PerPage: 10}
	mcs_d, vsTotal, vErr := man.SearchContent(sr)
	assert.NoError(t, vErr, "Video error by name search failed")
	assert.Equal(t, 1, int(vsTotal), "We should be able to find donut.mp4 with an ilike")
	mc_donut := (*mcs_d)[0]
	assert.Equal(t, 2, len(mc_donut.Screens), fmt.Sprintf("It should load two screens %s", mc_donut.Screens))

	sr = ContentQuery{Page: 1, PerPage: 40, ContentType: "video"}
	vids, vidTotal, dbErr := man.SearchContent(sr)
	assert.NoError(t, dbErr, "Should search content type")
	assert.Equal(t, int64(1), vidTotal, "The total count for videos is 1")
	assert.Equal(t, 1, len(*vids), "We should have one result")

	sr = ContentQuery{Page: 1, PerPage: 10, ContainerID: strconv.FormatInt(cnt1.ID, 10)}
	all_mcs, total, err := man.SearchContent(sr)
	assert.NoError(t, err, "It should be able to empty search")
	assert.Equal(t, int64(12), total, "The total count for this dir is 12")
	assert.Equal(t, 10, len(*all_mcs), "But we limited the pagination")
}

func TestDbManagerMultiSearch(t *testing.T) {
	// Test that a search restricting containerID works
	// Test that search restricting container and text works
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)

	man := GetManagerTestSuite(cfg)
	assert.Equal(t, man.CanEdit(), true)

	cnt1, content1, err1 := test_common.CreateContentByDirName("dir1")
	assert.NoError(t, err1)
	assert.Greater(t, len(content1), 1, "Dir1 should have content")

	cnt2, content2, err2 := test_common.CreateContentByDirName("dir2")
	assert.NoError(t, err2)
	assert.Greater(t, len(content2), 1, "Dir2 should also have content")

	sr := ContentQuery{Search: content1[1].Src, PerPage: 10, ContainerID: strconv.FormatInt(cnt1.ID, 10)}
	found, count, err := man.SearchContent(sr)
	assert.Equal(t, len(*found), 1, "We should have found our item")
	assert.Equal(t, int64(1), count)
	assert.NoError(t, err)

	sr = ContentQuery{Search: "blah", ContainerID: strconv.FormatInt(cnt1.ID, 10)}
	_, nCount, n_err := man.SearchContent(sr)
	assert.Equal(t, int64(0), nCount, "It should not find this the content name is invalid")
	assert.NoError(t, n_err)
	sr = ContentQuery{Search: content1[1].Src, ContainerID: strconv.FormatInt(cnt2.ID, 10)}

	_, notInCntCount, not_err := man.SearchContent(sr)
	assert.Equal(t, int64(0), notInCntCount, "It should not find this valid content as it is not in container")
	assert.NoError(t, not_err)
}

func TestManagerDBBasics(t *testing.T) {
	db := models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)

	cnt, content := test_common.GetContentByDirName("dir1")
	assert.Equal(t, "dir1", cnt.Name, "It should be the right dir")
	assert.Equal(t, 12, cnt.Total, "The container total should be this for dir1")
	assert.Equal(t, 12, len(content))

	cntRes := db.Create(cnt)
	assert.NoError(t, cntRes.Error, "It should create the container")
	for _, mc := range content {
		cRes := db.Create(&mc)
		assert.NoError(t, cRes.Error, "It should create contents")
	}

	man := GetManagerTestSuite(cfg)
	qContent, count, err := man.ListContent(ContentQuery{PerPage: 14})
	assert.NoError(t, err, "We should be able to list")
	assert.NotNil(t, qContent, "The content should be defined")
	assert.Equal(t, len(*qContent), 12, "there should be 12 results")
	assert.Equal(t, int64(12), count, "Count should be the same")

	lim_content, count, _ := man.ListContent(ContentQuery{PerPage: 3})
	assert.Equal(t, 3, len(*lim_content), "The DB should be setup with 10 items")
	assert.Equal(t, int64(12), count, "The count does not care about the page")
}

func TestManagerDBTags(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)

	assert.NoError(t, man.CreateTag(&models.Tag{ID: "A"}), "couldn't create tag A")
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "B"}), "couldn't create tag B")
	tags, total, err := man.ListAllTags(TagQuery{PerPage: 3})
	assert.NoError(t, err, "It should be able to list tags")
	assert.Greater(t, total, int64(0), "It should have a total")
	assert.Equal(t, len(*tags), 2, "We should have two tags")
}

func TestManagerDBTagsAssignment(t *testing.T) {
	db := models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)

	// TODO: Break up between DB and memory (check that it can ignore tags maybe)
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "aws"}), "couldn't create tag aws")
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "donut"}), "couldn't create tag donut")

	cnt, content := test_common.GetContentByDirName("dir2")
	assert.Greater(t, len(content), 0, "There should be content")
	cntRes := db.Create(cnt)
	assert.NoError(t, cntRes.Error, "It should create the container")
	for _, mc := range content {
		cRes := db.Create(&mc)
		assert.NoError(t, cRes.Error, "It should create content")
	}

	tags, total, err := man.ListAllTags(TagQuery{})
	assert.NoError(t, err)
	assert.Equal(t, total, 2, fmt.Sprintf("And only two tags %s", tags))
	assert.Greater(t, len(*tags), 0, "We should have tags")
	assignmentErr := AssignTagsAndUpdate(man, *tags)
	assert.NoError(t, assignmentErr, fmt.Sprintf("Error %s", assignmentErr))

	cq := ContentQuery{Tags: []string{"aws", "donut"}}
	contentsMatching, total, err := man.SearchContent(cq)
	assert.NoError(t, err)
	assert.Equal(t, len(*contentsMatching), 2, fmt.Sprintf("Found content but not the right amount %s", contentsMatching))
	assert.Equal(t, total, 2)
}

func Test_ManagerTagsDB_CRUD(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	tag := models.Tag{ID: "A"}
	assert.NoError(t, man.CreateTag(&tag), "couldn't create tag A")

	tags, total, err := man.ListAllTags(TagQuery{PerPage: 3})
	assert.NoError(t, err)
	assert.Greater(t, total, 0, "A tag should exist")
	assert.Equal(t, len(*tags), 1, "We should have one tag")
	man.DestroyTag(tag.ID)
	tags_gone, total_gone, err := man.ListAllTags(TagQuery{PerPage: 3})
	assert.NoError(t, err)
	assert.Equal(t, len(*tags_gone), 0, "No tags should be in the DB")
	assert.Equal(t, total_gone, 0, "It should have no tags")
}

func Test_DbManager_AssociateTags(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)

	// The Eager code just doesn't work in Buffalo?
	t1 := models.Tag{ID: "A"}
	t2 := models.Tag{ID: "B"}
	man.CreateTag(&t1)
	man.CreateTag(&t2)
	mc := models.Content{Src: "A", Preview: "p", ContentType: "video"}
	mc.Tags = models.Tags{t1, t2}
	man.CreateContent(&mc)

	s := models.Screen{Src: "screen1", ContentID: mc.ID}
	man.CreateScreen(&s)
	mc.Screens = models.Screens{s}
	man.UpdateContent(&mc)

	tags, total, t_err := man.ListAllTags(TagQuery{PerPage: 10})
	assert.NoError(t, t_err, "We should be able to list tags.")
	assert.Equal(t, 2, len(*tags), fmt.Sprintf("There should be two tags %s", mc))
	assert.Greater(t, total, 0, "It should have a total")

	screens, count, s_err := man.ListScreens(ScreensQuery{ContentID: strconv.FormatInt(mc.ID, 10)})
	assert.NoError(t, s_err, "Screens should list")
	assert.Equal(t, 1, len(*screens), "We should have a screen associated")
	assert.Equal(t, 1, count, "We should have a proper screen count")

	tCheck, _ := man.GetContent(mc.ID)
	assert.Equal(t, 2, len(tCheck.Tags), fmt.Sprintf("It should eager load tags %s", tCheck))

	t3 := models.Tag{ID: "C"}
	man.CreateTag(&t3)
	err := man.AssociateTagByID(t3.ID, mc.ID)
	assert.NoError(t, err, fmt.Sprintf("We shouldn't have an issue associating this %s \n", err))
	mcCheck, mc_err := man.GetContent(mc.ID)
	assert.NoError(t, mc_err, fmt.Sprintf("We should be able to load back the content %s", err))
	assert.Equal(t, 3, len(mcCheck.Tags), fmt.Sprintf("There should be a new tag %s", mcCheck))
}

// A Lot more of these could be a test in manager that passes in the manager
// TODO: Remove copy pasta and make it almost identical.
func Test_DbManager_TagSearch(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)

	man := GetManagerTestSuite(cfg)
	ManagersTagSearchValidation(t, man)
}

func Test_ManagerDBPreviews(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)

	mc1 := models.Content{Src: "A", Preview: "p", ContentType: "video"}
	mc2 := models.Content{Src: "B", Preview: "p", ContentType: "video"}
	mc3 := models.Content{Src: "C", Preview: "p", ContentType: "video"}
	man.CreateContent(&mc1)
	man.CreateContent(&mc2)
	man.CreateContent(&mc3)
	assert.Greater(t, 0, mc1.ID)

	p1 := models.Screen{Src: "fake1", Idx: 0, ContentID: mc1.ID}
	p2 := models.Screen{Src: "fake2.png", Idx: 1, ContentID: mc1.ID}
	p3 := models.Screen{Src: "fake3.png", Idx: 1, ContentID: mc2.ID}

	man.CreateScreen(&p1)
	man.CreateScreen(&p2)
	man.CreateScreen(&p3)

	previewList, count1, err := man.ListScreens(ScreensQuery{ContentID: strconv.FormatInt(mc1.ID, 10)})
	assert.NoError(t, err)
	assert.Equal(t, len(*previewList), 2, "We should have two previews")
	assert.Equal(t, count1, 2, "We should have two previews")

	previewOne, count2, p_err := man.ListScreens(ScreensQuery{ContentID: strconv.FormatInt(mc2.ID, 10)})
	assert.NoError(t, p_err)
	assert.Equal(t, len(*previewOne), 1, "Now there should be 1")
	assert.Equal(t, count2, 1, "Now there should be 1")

	p4 := models.Screen{Src: "fake4.png", Idx: 1, ContentID: mc2.ID}
	c_err := man.CreateScreen(&p4)
	assert.NoError(t, c_err)

	p4_check, p4_err := man.GetScreen(p4.ID)
	assert.NoError(t, p4_err, fmt.Sprintf("Failed to pull back the screen by ID %d", p4.ID))
	assert.Equal(t, p4_check.Src, p4.Src)
}

func Test_ManagerDBSearchScreens(t *testing.T) {
	db := models.ResetDB(models.InitGorm(false))
	cfg := test_common.InitFakeApp(true)

	man := ContentManagerDB{cfg: cfg}
	man.GetConnection = func() *gorm.DB {
		return db
	}

	// Hmm, might want to make a wrapper for the create
	mc1 := models.Content{Src: "1", Preview: "one", ContentType: "video/mp4"}
	mc2 := models.Content{Src: "2", Preview: "none", ContentType: "video/mp4"}
	mc3 := models.Content{Src: "3", Preview: "none", ContentType: "video/mp4"}
	mc4 := models.Content{Src: "4", Preview: "none", ContentType: "image/png"}
	mc5 := models.Content{Src: "No Previews", Preview: "none", ContentType: "video/mp4"}
	man.CreateContent(&mc1)
	man.CreateContent(&mc2)
	man.CreateContent(&mc3)
	man.CreateContent(&mc4)
	man.CreateContent(&mc5)

	p1 := models.Screen{Src: "fake1.screen", Idx: 1, ContentID: mc1.ID}
	p2 := models.Screen{Src: "fake2.screen", Idx: 1, ContentID: mc2.ID}
	p3 := models.Screen{Src: "fake3.screen1", Idx: 1, ContentID: mc3.ID}
	p4 := models.Screen{Src: "fake3.screen2", Idx: 1, ContentID: mc3.ID}
	p5 := models.Screen{Src: "ShouldNotLoadContentIsImage", Idx: 1, ContentID: mc4.ID}
	man.CreateScreen(&p1)
	man.CreateScreen(&p2)
	man.CreateScreen(&p3)
	man.CreateScreen(&p4)
	man.CreateScreen(&p5)

	// Intentionally exclude mc2 to ensure we get some screens, include one with no screens
	content := models.Contents{mc1, mc3, mc4, mc5}
	screens, s_err := man.LoadRelatedScreens(&content)
	assert.NoError(t, s_err, "It shouldn't error out")
	assert.NotNil(t, screens, "No screens were returned")
	assert.Equal(t, 2, len(screens), "It should load all the screens but only two of these have screens")

	assert.Equal(t, 1, len(screens[mc1.ID]), "MC1 has 1 screen")
	assert.Equal(t, 2, len(screens[mc3.ID]), "MC3 has 2 screens")

	// Test that an image will not load previews
	content_2 := models.Contents{mc2, mc4}
	screens_2, s2_err := man.LoadRelatedScreens(&content_2)
	assert.NoError(t, s2_err, "It shouldn't error out")
	assert.Equal(t, 1, len(screens_2), "It should load all the screens for mc2 but EXCLUDE mc4")
}

func Test_DBManager_IllegalContainers(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.ResetConfig()
	test_common.InitFakeApp(true)
	ctx := test_common.GetContext()
	man := GetManager(ctx)

	notUnderDir := models.Container{Name: "ssl", Path: "/etc"}
	assert.Error(t, man.CreateContainer(&notUnderDir), "Not under the configured directory, rejected")

	upAccess := models.Container{Name: "../../.ssh/", Path: cfg.Dir}
	assert.Error(t, man.CreateContainer(&upAccess), "No up access allowed in names")

	multiLevelDownOk := models.Container{Name: "screens/screens_sub_dir", Path: cfg.Dir}
	assert.NoError(t, man.CreateContainer(&multiLevelDownOk), "This should exist in the mock data")

	knownDirOk := models.Container{Name: "dir2", Path: cfg.Dir}
	assert.NoError(t, man.CreateContainer(&knownDirOk), "This directory should be ok")
	knownDirOk.Name = "NowInvalid"
	_, err := man.UpdateContainer(&knownDirOk)
	assert.Error(t, err, "It should not allow this invalid directory")

	test_common.CreateContainerPath(&knownDirOk)
	defer test_common.CleanupContainer(&knownDirOk)
	up, upErr := man.UpdateContainer(&knownDirOk)
	assert.NoError(t, upErr, "Now it should be ok as the directory exists")
	assert.Equal(t, up.Name, knownDirOk.Name, "And it returns a fresh load to prove it is updated.")
}

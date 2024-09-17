package managers

import (
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryManagerContainers(t *testing.T) {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext()
	man := GetManager(ctx)
	containers, count, err := man.ListContainersContext()
	assert.NoError(t, err)
	assert.Greater(t, count, int64(1), "There should be containers")

	for _, c := range *containers {
		c_mem, err := man.GetContainer(c.ID)
		if err != nil {
			assert.Fail(t, "It should not have an issue finding valid containers")
		}
		assert.Equal(t, c_mem.ID, c.ID)
	}
}

func TestMemoryManagerContent(t *testing.T) {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext()
	man := GetManager(ctx)
	contents, count, err := man.ListContent(ContentQuery{})
	assert.NoError(t, err)
	assert.Greater(t, count, int64(0), "It should have contents")

	for _, content := range *contents {
		cm, err := man.GetContent(content.ID)
		if err != nil {
			assert.Fail(t, "It should not have an issue finding valid content")
		}
		assert.Equal(t, cm.ID, content.ID)
	}
}

func TestMemoryManagerAssignManager(t *testing.T) {
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = false
	utils.InitConfig(cfg.Dir, cfg)

	mem := ContentManagerMemory{}
	mem.validate = "Memory"
	mem.SetCfg(cfg)
	mem.Initialize()

	memCfg := mem.GetCfg()
	assert.NotNil(t, memCfg, "It should be defined")
	mcs, _, err := mem.ListContent(ContentQuery{})
	assert.NoError(t, err)
	assert.Greater(t, len(*mcs), 0, "It should have valid files in the manager")

	cfg.UseDatabase = false
	ctx := test_common.GetContext()
	man := GetManager(ctx) // New Reference but should have the same count of content
	mcs_2, _, _ := man.ListContent(ContentQuery{})
	assert.Equal(t, len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}

func TestMemoryManagerPaginate(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	cfg.UseDatabase = false
	cfg.ReadOnly = true

	ctx := test_common.GetContextParams("/containers", "1", "2")
	man := GetManager(ctx)
	assert.Equal(t, man.CanEdit(), false, "Memory manager should not allow editing")

	containers, count, err := man.ListContainers(ContainerQuery{Page: 1, PerPage: 1})
	assert.NoError(t, err, "It should list with pagination")
	assert.Equal(t, len(*containers), 1, "It should respect paging")
	assert.Equal(t, int64(test_common.TOTAL_CONTAINERS_WITH_CONTENT), count, "Paging check that count is still correct")

	cnt := (*containers)[0]
	assert.NotNil(t, cnt, "There should be a container with 12 entries")
	assert.Equal(t, cnt.Total, 12, "There should be 12 test images in the first ORDERED containers")
	assert.NoError(t, err)
	assert.NotEqual(t, "", cnt.PreviewUrl, "The previewUrl should be set")
	content_page_1, count, _ := man.ListContent(ContentQuery{ContainerID: strconv.FormatInt(cnt.ID, 10), PerPage: 4})
	assert.Equal(t, len(*content_page_1), 4, "It should respect page size")
	assert.Equal(t, count, int64(12), "It should respect page size but get the total count")

	content_page_3, count, _ := man.ListContent(ContentQuery{ContainerID: strconv.FormatInt(cnt.ID, 10), Page: 3, PerPage: 4})
	assert.Equal(t, len(*content_page_3), 4, "It should respect page size and get the last page")
	assert.NotEqual(t, (*content_page_3)[3].ID, (*content_page_1)[3].ID, "Ensure it actually paged")
	assert.Greater(t, count, int64(0), "We should still have a count")

	// Last container pagination check
	l_cnts, count, _ := man.ListContainers(ContainerQuery{Page: 4, PerPage: 1})
	assert.Equal(t, 1, len(*l_cnts), "It should still return only as we are on the last page")
	assert.Equal(t, int64(test_common.TOTAL_CONTAINERS_WITH_CONTENT), count, "The count should be consistent")
	l_cnt := (*l_cnts)[0]
	assert.Equal(t, test_common.EXPECT_CNT_COUNT[l_cnt.Name], l_cnt.Total, "There are 3 entries in the ordered test data last container")
}

func TestMemoryManagerInitialize(t *testing.T) {
	test_common.InitFakeApp(false)

	ctx := test_common.GetContext()
	man := GetManager(ctx)
	assert.NotNil(t, man, "It should have a manager defined after init")

	containers, _, err := man.ListContainersContext()
	assert.NoError(t, err, "It should list all containers")
	assert.NotNil(t, containers, "It should have containers")
	assert.Equal(t, len(*containers), test_common.TOTAL_CONTAINERS_WITH_CONTENT, "Unexpected container count")

	// Memory test working
	for _, c := range *containers {
		// fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
		content, _, err := man.ListContent(ContentQuery{ContainerID: strconv.FormatInt(c.ID, 10)})
		assert.NoError(t, err)
		assert.NotNil(t, content)

		content_len := len(*content)
		// fmt.Printf("Content length was %d\n", content_len)
		assert.Greater(t, content_len, 0, "There should be a number of content")
		assert.Equal(t, test_common.EXPECT_CNT_COUNT[c.Name], content_len, "It should have this many instances: "+c.Name)
		assert.Greater(t, c.Total, 0, "All of them should have a total assigned")
	}
}

func TestMemoryManagerSearch(t *testing.T) {
	test_common.InitFakeApp(false)

	ctx := test_common.GetContext()
	man := GetManager(ctx)
	assert.NotNil(t, man, "It should have a manager defined after init")

	containers, _, err := man.ListContainersContext()
	assert.NoError(t, err, "It should list all containers")
	assert.NotNil(t, containers, "It should have containers")
	assert.Equal(t, test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(*containers), "Wrong number of containers found")

	s_cnts, count, s_err := man.SearchContainers(ContainerQuery{Search: "dir2", Page: 1, PerPage: 2})
	assert.NoError(t, s_err, "Error searching memory containers")
	assert.Equal(t, 1, len(*s_cnts), "It should only filter to one directory")
	assert.Equal(t, int64(1), count, "There should be one count")

	sr := ContentQuery{Search: "Donut", PerPage: 20}
	mcs, total, err := man.SearchContent(sr)
	assert.NoError(t, err, "Can we search in the memory manager")
	assert.Equal(t, 1, len(*mcs), "One donut should be found")
	assert.Equal(t, total, int64(len(*mcs)), "It should get the total right")

	sr = ContentQuery{Search: "Large", PerPage: 6}
	mcs_1, _, err_1 := man.SearchContent(sr)
	assert.NoError(t, err_1, "Can we search in the memory manager")
	assert.Equal(t, 5, len(*mcs_1), "There are 5 images with 'large' in them ignoring case")

	sr = ContentQuery{PerPage: 9001}
	all_mc, _, err_all := man.SearchContent(sr)
	assert.NoError(t, err_all, "Can in search everything")
	assert.Equal(t, len(*all_mc), test_common.TOTAL_MEDIA, "The Kitchen sink")
}

func TestMemoryManagerSearchMulti(t *testing.T) {
	// Test that a search restricting containerID works
	// Test that search restricting container and text works
	cfg, _ := test_common.InitFakeApp(false)
	cfg.ReadOnly = false
	ctx := test_common.GetContext()
	man := GetManager(ctx)

	// Ensure we initialized with a known search
	assert.Equal(t, man.CanEdit(), true)
	sr := ContentQuery{Search: "donut"}
	mcs, total, err := man.SearchContent(sr)
	assert.NoError(t, err, "Can we search in the memory manager")
	assert.Equal(t, len(*mcs), 1, "One donut should be found")
	assert.Equal(t, total, int64(len(*mcs)), "It should get the total right")

	cnts, _, eep := man.ListContainers(ContainerQuery{Page: 1, PerPage: 10})
	assert.NoError(t, eep, fmt.Sprintf("It should have 4 containers %s", eep))
	assert.Greater(t, len(*cnts), 1, "We should have containers")

	allContent, count, errAll := man.ListContent(ContentQuery{PerPage: 50})
	assert.Greater(t, len(*allContent), 0, "We should have content")
	assert.Greater(t, count, int64(0), "We should have content")
	assert.NoError(t, errAll)

	sr = ContentQuery{Text: "", PerPage: 40}
	all_content, wild_total, _ := man.SearchContent(sr)
	assert.Greater(t, wild_total, int64(0), "It should work with a large query")
	assert.Equal(t, int64(len(*all_content)), wild_total)

	sr = ContentQuery{ContentType: "video", Order: "src", Direction: "asc"}
	video_content, vid_total, _ := man.SearchContent(sr)
	assert.Equal(t, int64(test_common.TOTAL_VIDEO), vid_total)
	assert.Equal(t, int64(len(*video_content)), vid_total)
	vs := *video_content
	assert.Equal(t, test_common.VIDEO_FILENAME, vs[0].Src)

	for _, cnt := range *cnts {
		if cnt.Name == "dir1" {
			sr = ContentQuery{Search: "donut", ContainerID: strconv.FormatInt(cnt.ID, 10)}
			_, noTotal, n_err := man.SearchContent(sr)
			assert.NoError(t, n_err)
			assert.Equal(t, int64(0), noTotal, "It should not be in this directory")
		}
		if cnt.Name == "dir2" {
			sr = ContentQuery{Search: "donut", ContainerID: strconv.FormatInt(cnt.ID, 10)}
			yes_match, yTotal, r_err := man.SearchContent(sr)
			assert.NoError(t, r_err)
			assert.Equal(t, int64(1), yTotal, "We did not find the expected content")

			movie := (*yes_match)[0]
			assert.Equal(t, movie.Src, test_common.VIDEO_FILENAME)

			sr = ContentQuery{ContainerID: strconv.FormatInt(cnt.ID, 10), ContentType: "image"}
			_, imgCount, _ := man.SearchContent(sr)
			assert.Equal(t, int64(2), imgCount, "It should filter out the donut this time")
		}
		if cnt.Name == "dir3" {
			sr = ContentQuery{ContainerID: strconv.FormatInt(cnt.ID, 10), PerPage: 1}
			has_content, _, err := man.SearchContent(sr)
			assert.NoError(t, err, "We should have content")
			assert.Greater(t, len(*has_content), 0)
		}
	}
}

func TestMemoryPreviewInitialization(t *testing.T) {
	cfg := test_common.ResetConfig()
	utils.SetupContentMatchers(cfg, "", "video", "DS_Store", "")
	utils.SetCfg(*cfg)

	// Create a fake file that would sub in by name for a preview
	testDir := models.GetEnvString("DIR", "")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := utils.GetPreviewDst(srcDir)
	testFile := test_common.VIDEO_FILENAME

	// Create a fake preview
	utils.ResetPreviewDir(dstDir)

	fqPath := utils.GetPreviewPathDestination(testFile, dstDir, "video/mp4")
	f, err := os.Create(fqPath)
	assert.NoError(t, err, fmt.Sprintf("Could not create the file at %s", fqPath))
	_, wErr := f.WriteString("Now something exists in the file")
	assert.NoError(t, wErr, fmt.Sprintf("Could not write to the file at %s", fqPath))
	assert.Contains(t, fqPath, fmt.Sprintf("%s.png", test_common.VIDEO_FILENAME))
	f.Sync()

	// Checks that if a preview exists
	cnts, content, _, _ := utils.PopulateMemoryView(cfg.Dir)
	assert.Equal(t, 2, len(cnts), "We should only pull in containers that have video content")
	assert.Equal(t, test_common.TOTAL_VIDEO, len(content), fmt.Sprintf("There are %d videos", test_common.TOTAL_VIDEO))
	foundDonut := false
	for _, mc := range content {
		if mc.Src == test_common.VIDEO_FILENAME {
			expect := fmt.Sprintf("/container_previews/%s.png", test_common.VIDEO_FILENAME)
			assert.Equal(t, expect, mc.Preview)
			foundDonut = true
		}
	}
	assert.Equal(t, true, foundDonut, "We should have a video preview for the donuts")

	cfg.ExcludeEmptyContainers = false
	all_cnts, one_content, _, _ := utils.PopulateMemoryView(cfg.Dir)
	assert.Equal(t, test_common.TOTAL_VIDEO, len(one_content), "Expect some videos")
	assert.Equal(t, test_common.TOTAL_CONTAINERS, len(all_cnts), "Allow it to pull in all containers")
}

func TestManagerMemoryTags(t *testing.T) {
	cfg := test_common.InitMemoryFakeAppEmpty()
	man := GetManagerTestSuite(cfg)
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "A"}), "couldn't create tag A")
	assert.NoError(t, man.CreateTag(&models.Tag{ID: "B"}), "couldn't create tag B")
	tags, total, err := man.ListAllTags(TagQuery{PerPage: 3})
	assert.NoError(t, err, "It should be able to list tags")
	assert.Equal(t, 2, len(*tags), "We should have two tags")
	assert.Equal(t, int64(2), total, "It should have a tag count")
}

// A Lot more of these could be a test in manager that passes in the manager
// TODO: Remove copy pasta and make it almost identical.
func TestMemoryManagerTagSearch(t *testing.T) {
	cfg := test_common.InitMemoryFakeAppEmpty()
	man := GetManagerTestSuite(cfg)
	ManagersTagSearchValidation(t, man)
}

func TestMemoryMangerTagsMemoryCRUD(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	man := GetManagerTestSuite(cfg)

	tag := models.Tag{ID: "A"}
	assert.NoError(t, man.CreateTag(&tag), "couldn't create tag A")
	assert.NoError(t, man.UpdateTag(&tag), "It should udpate")

	tags, total, err := man.ListAllTags(TagQuery{PerPage: 3})
	assert.NoError(t, err, "It should be able to list tags")
	assert.Equal(t, int64(1), total, "there should be one tag")
	assert.Equal(t, 1, len(*tags), "We should have one tag")
	man.DestroyTag(tag.ID)

	tags_gone, total_gone, _ := man.ListAllTags(TagQuery{PerPage: 3})
	assert.Equal(t, len(*tags_gone), 0, "Now there should be no tags")
	assert.Equal(t, int64(0), total_gone, "it should be empty")
}

func TestManagerMemoryScreens(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)

	man := GetManagerTestSuite(cfg)
	content, count, err := man.ListContent(ContentQuery{PerPage: 100})
	assert.NoError(t, err)
	assert.Greater(t, len(*content), 0, "It should have content setup")
	assert.Greater(t, count, int64(0), "It should have content counted")

	contentArr := *content
	mc := contentArr[0]
	id1 := utils.AssignNumerical(0, "screens")
	id2 := utils.AssignNumerical(0, "screens")

	s1 := models.Screen{ID: id1, Path: "A", Src: "a.txt", ContentID: mc.ID}
	s2 := models.Screen{ID: id2, Path: "B", Src: "b.txt", ContentID: mc.ID}
	mc.Screens = models.Screens{s1, s2}

	// Ensure we actually set the right object in the backing Map
	mem := utils.GetMemStorage()
	mem.ValidContent[mc.ID] = mc
	mem.ValidScreens[s1.ID] = s1
	mem.ValidScreens[s2.ID] = s2

	screens, count, err := man.ListScreens(ScreensQuery{ContentID: strconv.FormatInt(mc.ID, 10)})
	assert.NoError(t, err)
	assert.NotNil(t, screens)
	assert.Equal(t, 2, len(*screens), "We should have two screens")
	assert.Equal(t, int64(2), count, "And the count should be right")
	// Check that our single lookup hash is also populated
	for _, screen := range *screens {
		obj, mia := man.GetScreen(screen.ID)
		assert.NoError(t, mia)
		assert.Equal(t, obj.ID, screen.ID)
	}

	allScreens, allCount, allErr := man.ListScreens(ScreensQuery{})
	assert.NoError(t, allErr, "It should work out ok")
	assert.Equal(t, 2, len(*allScreens), "We should have 2 screens")
	assert.Equal(t, int64(2), allCount, "We should have 2 screens")
}

func TestManagerMemoryCRU(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	ctx := test_common.GetContext()
	man := GetManager(ctx)

	// Only valid paths now allow for creation even in tests.
	c := models.Container{Path: cfg.Dir, Name: "A"}
	c2 := models.Container{Path: cfg.Dir, Name: "B"}
	test_common.CreateContainerPath(&c)
	test_common.CreateContainerPath(&c2)
	defer test_common.CleanupContainer(&c)
	defer test_common.CleanupContainer(&c2)

	assert.NoError(t, man.CreateContainer(&c), "Did not create container")
	assert.NoError(t, man.CreateContainer(&c2), "Did not create container")
	c_check, c_err := man.GetContainer(c.ID)
	assert.NoError(t, c_err, "We should be able to get back the container")
	assert.Equal(t, c_check.Path, c.Path, "Ensure we are not stomping unset ID data")

	content1 := models.Content{Src: "content1", ContainerID: &c.ID, NoFile: true}
	content2 := models.Content{Src: "content2", ContainerID: &c.ID, NoFile: true}
	assert.NoError(t, man.CreateContent(&content1), "Did not create content correctly")
	assert.NoError(t, man.CreateContent(&content2), "Did not create content correctly")
	mcUp := models.Content{Src: "updated", ID: content1.ID, ContainerID: &c.ID, NoFile: true}
	man.UpdateContent(&mcUp)
	mc_check, m_err := man.GetContent(content1.ID)
	assert.NoError(t, m_err, "It should find this content")
	assert.Equal(t, mc_check.Src, "updated")

	s1 := models.Screen{Path: "A", Src: "a.txt", ContentID: content1.ID}
	s2 := models.Screen{Path: "B", Src: "b.txt", ContentID: content2.ID}
	assert.NoError(t, man.CreateScreen(&s1), "Did not associate screen correctly")
	assert.NoError(t, man.CreateScreen(&s2), "It should not allow a screen that does not have")

	sCheck, count, sErr := man.ListScreens(ScreensQuery{ContentID: strconv.FormatInt(content1.ID, 10)})
	assert.NoError(t, sErr, "Failed to list screens")
	assert.Equal(t, len(*sCheck), 1, "It should properly filter screens.")
	assert.Equal(t, int64(1), count, "Count should be correct")

	s1Update := models.Screen{ID: s1.ID, Path: "C", ContentID: content1.ID}
	assert.NoError(t, man.UpdateScreen(&s1Update))
	s1Check, scErr := man.GetScreen(s1.ID)
	assert.NoError(t, scErr, "Failed to get the screen back")
	assert.Equal(t, s1Check.Path, "C")
}

func TestMemoryManagerTags(t *testing.T) {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext()
	man := GetManager(ctx)

	aTag := models.Tag{ID: "A"}
	bTag := models.Tag{ID: "B"}
	assert.NoError(t, man.CreateTag(&aTag))
	assert.NoError(t, man.CreateTag(&bTag))

	content := models.Content{Src: "SomethingSomethingDarkside", NoFile: true}
	assert.NoError(t, man.CreateContent(&content))

	assert.NoError(t, man.AssociateTagByID(aTag.ID, content.ID))
	assert.NoError(t, man.AssociateTagByID(bTag.ID, content.ID))
	checkContent, err := man.GetContent(content.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, checkContent.Tags, "We should have tags")
	assert.Equal(t, len(checkContent.Tags), 2, "There should be two tags")

	// Not in the DB so should not associate
	notExistsTag := models.Tag{ID: "NOPE"}
	assert.Error(t, man.AssociateTagByID(notExistsTag.ID, content.ID))
}

func TestMemoryManagerIllegalContainers(t *testing.T) {
	cfg := test_common.ResetConfig()
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext()
	man := GetManager(ctx)

	notUnderDir := models.Container{Name: "ssl", Path: "/etc"}
	assert.Error(t, man.CreateContainer(&notUnderDir), "Not under the configured directory, rejected")

	upAccess := models.Container{Name: "../../.ssh/", Path: cfg.Dir}
	assert.Error(t, man.CreateContainer(&upAccess), "No up access allowed in names")

	// Ensure that a container can create, but an invalid update is prevented.
	knownDirOk := models.Container{Name: "dir2", Path: cfg.Dir}
	assert.NoError(t, man.CreateContainer(&knownDirOk), "This directory should be ok")
	knownDirOk.Name = "INVALID"
	_, err := man.UpdateContainer(&knownDirOk)
	assert.Error(t, err, "The Path is illegal")

	multiLevelDownOk := models.Container{Name: "screens/screens_sub_dir", Path: cfg.Dir}
	assert.NoError(t, man.CreateContainer(&multiLevelDownOk), "This should exist in the mock data")
}

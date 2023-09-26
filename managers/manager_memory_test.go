package managers

import (
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_ManagerContainers() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)
	containers, err := man.ListContainersContext()
	as.NoError(err)

	for _, c := range *containers {
		c_mem, err := man.GetContainer(c.ID)
		if err != nil {
			as.Fail("It should not have an issue finding valid containers")
		}
		as.Equal(c_mem.ID, c.ID)
	}
}

func (as *ActionSuite) Test_ManagerContent() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)
	mcs, err := man.ListAllContent(1, 9001)
	as.NoError(err)

	for _, mc := range *mcs {
		cm, err := man.GetContent(mc.ID)
		if err != nil {
			as.Fail("It should not have an issue finding valid content")
		}
		as.Equal(cm.ID, mc.ID)
	}
}

func (as *ActionSuite) Test_AssignManager() {
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = false
	utils.InitConfig(cfg.Dir, cfg)

	mem := ContentManagerMemory{}
	mem.validate = "Memory"
	mem.SetCfg(cfg)
	mem.Initialize()

	memCfg := mem.GetCfg()
	as.NotNil(memCfg, "It should be defined")
	mcs, err := mem.ListAllContent(1, 9001)
	as.NoError(err)
	as.Greater(len(*mcs), 0, "It should have valid files in the manager")

	cfg.UseDatabase = false
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx) // New Reference but should have the same count of content
	mcs_2, _ := man.ListAllContent(1, 9000)

	as.Equal(len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}

func (as *ActionSuite) Test_MemoryManagerPaginate() {
	cfg := test_common.InitFakeApp(false)
	cfg.UseDatabase = false

	ctx := test_common.GetContextParams(as.App, "/containers", "1", "2")
	man := GetManager(&ctx)
	as.Equal(man.CanEdit(), false, "Memory manager should not allow editing")

	containers, err := man.ListContainers(1, 1)
	as.NoError(err, "It should list with pagination")
	as.Equal(1, len(*containers), "It should respect paging")

	cnt := (*containers)[0]
	as.NotNil(cnt, "There should be a container with 12 entries")
	as.Equal(cnt.Total, 12, "There should be 12 test images in the first ORDERED containers")
	as.NoError(err)
	as.NotEqual("", cnt.PreviewUrl, "The previewUrl should be set")
	content_page_1, _ := man.ListContent(cnt.ID, 1, 4)
	as.Equal(len(*content_page_1), 4, "It should respect page size")

	content_page_3, _ := man.ListContent(cnt.ID, 3, 4)
	as.Equal(len(*content_page_3), 4, "It should respect page size and get the last page")
	as.NotEqual((*content_page_3)[3].ID, (*content_page_1)[3].ID, "Ensure it actually paged")

	// Last container pagination check
	l_cnts, _ := man.ListContainers(4, 1)
	as.Equal(1, len(*l_cnts), "It should still return only as we are on the last page")
	l_cnt := (*l_cnts)[0]
	as.Equal(test_common.EXPECT_CNT_COUNT[l_cnt.Name], l_cnt.Total, "There are 3 entries in the ordered test data last container")
}

func (as *ActionSuite) Test_ManagerInitialize() {
	test_common.InitFakeApp(false)

	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)
	as.NotNil(man, "It should have a manager defined after init")

	containers, err := man.ListContainersContext()
	as.NoError(err, "It should list all containers")
	as.NotNil(containers, "It should have containers")
	as.Equal(len(*containers), test_common.TOTAL_CONTAINERS, "Unexpected container count")

	// Memory test working
	for _, c := range *containers {
		// fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
		content, err := man.ListContentContext(c.ID)
		as.NoError(err)
		as.NotNil(content)

		content_len := len(*content)
		// fmt.Printf("Content length was %d\n", content_len)
		as.Greater(content_len, 0, "There should be a number of content")
		as.Equal(test_common.EXPECT_CNT_COUNT[c.Name], content_len, "It should have this many instances: "+c.Name)
		as.Greater(c.Total, 0, "All of them should have a total assigned")
	}
}

func (as *ActionSuite) Test_MemoryManagerSearch() {
	test_common.InitFakeApp(false)

	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)
	as.NotNil(man, "It should have a manager defined after init")

	containers, err := man.ListContainersContext()
	as.NoError(err, "It should list all containers")
	as.NotNil(containers, "It should have containers")
	as.Equal(len(*containers), test_common.TOTAL_CONTAINERS, "Wrong number of containers found")

	s_cnts, s_err := man.SearchContainers("dir2", 1, 2, false)
	as.NoError(s_err, "Error searching memory containers")
	as.Equal(1, len(*s_cnts), "It should only filter to one directory")

	mcs, total, err := man.SearchContent("Donut", 1, 20, "", "", false)
	as.NoError(err, "Can we search in the memory manager")
	as.Equal(len(*mcs), 1, "One donut should be found")
	as.Equal(total, len(*mcs), "It should get the total right")

	mcs_1, _, err_1 := man.SearchContent("Large", 1, 6, "", "", false)
	as.NoError(err_1, "Can we search in the memory manager")
	as.Equal(5, len(*mcs_1), "There are 5 images with 'large' in them ignoring case")

	all_mc, _, err_all := man.SearchContent("", 0, 9000, "", "", false)
	as.NoError(err_all, "Can in search everything")
	as.Equal(len(*all_mc), test_common.TOTAL_MEDIA, "The Kitchen sink")
}

func (as *ActionSuite) Test_MemoryManagerSearchMulti() {
	// Test that a search restricting containerID works
	// Test that search restricting container and text works
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)

	// Ensure we initialized with a known search
	as.Equal(man.CanEdit(), false)
	mcs, total, err := man.SearchContent("donut", 1, 20, "", "", false)
	as.NoError(err, "Can we search in the memory manager")
	as.Equal(len(*mcs), 1, "One donut should be found")
	as.Equal(total, len(*mcs), "It should get the total right")

	cnts, eep := man.ListContainers(0, 10)
	as.NoError(eep, "It should have 4 containers")
	as.Greater(len(*cnts), 1, "We should have containers")

	allContent, errAll := man.ListAllContent(0, 50)
	as.Greater(len(*allContent), 0, "We should have content")
	as.NoError(errAll)

	all_content, wild_total, _ := man.SearchContent("", 0, 40, "", "", false)
	as.Greater(wild_total, 0)
	as.Equal(len(*all_content), wild_total)

	video_content, vid_total, _ := man.SearchContent("", 0, 40, "", "video", false)
	as.Equal(vid_total, 1)
	as.Equal(len(*video_content), vid_total)
	vs := *video_content
	as.Equal(vs[0].Src, test_common.VIDEO_FILENAME)

	for _, cnt := range *cnts {
		if cnt.Name == "dir1" {
			_, no_total, n_err := man.SearchContent("donut", 1, 20, cnt.ID.String(), "", false)
			as.NoError(n_err)
			as.Equal(no_total, 0, "It should not be in this directory")
		}
		if cnt.Name == "dir2" {
			yes_match, y_total, r_err := man.SearchContent("donut", 1, 20, cnt.ID.String(), "", false)
			as.NoError(r_err)
			as.Equal(y_total, 1, "We did not find the expected content")

			movie := (*yes_match)[0]
			as.Equal(movie.Src, test_common.VIDEO_FILENAME)

			_, imgCount, _ := man.SearchContent("", 0, 20, cnt.ID.String(), "image", false)
			as.Equal(imgCount, 2, "It should filter out the donut this time")
		}
		if cnt.Name == "dir3" {
			has_content, _, err := man.SearchContent("", 0, 1, cnt.ID.String(), "", false)
			as.NoError(err, "We should have content")
			as.Greater(len(*has_content), 0)
		}
	}
}

func (as *ActionSuite) Test_MemoryPreviewInitialization() {
	cfg := test_common.ResetConfig()
	cfg.MaxSearchDepth = 1
	utils.SetupContentMatchers(cfg, "", "video", "DS_Store", "")
	utils.SetCfg(*cfg)

	// Create a fake file that would sub in by name for a preview
	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := utils.GetPreviewDst(srcDir)
	testFile := test_common.VIDEO_FILENAME

	// Create a fake preview
	utils.ResetPreviewDir(dstDir)

	fqPath := utils.GetPreviewPathDestination(testFile, dstDir, "video/mp4")
	f, err := os.Create(fqPath)
	if err != nil {
		as.T().Errorf("Could not create the file at %s", fqPath)
	}
	_, wErr := f.WriteString("Now something exists in the file")
	if wErr != nil {
		as.T().Errorf("Could not write to the file at %s", fqPath)
	}
	as.Contains(fqPath, fmt.Sprintf("%s.png", test_common.VIDEO_FILENAME))
	f.Sync()

	// Checks that if a preview exists
	cnts, content, _, _ := utils.PopulateMemoryView(cfg.Dir)
	as.Equal(1, len(cnts), "We should only pull in containers that have content")
	as.Equal(len(content), 1, "But there is only one video by mime type")
	for _, mc := range content {
		expect := fmt.Sprintf("/container_previews/%s.png", test_common.VIDEO_FILENAME)
		as.Equal(expect, mc.Preview)
	}

	cfg.ExcludeEmptyContainers = false
	all_cnts, one_content, _, _ := utils.PopulateMemoryView(cfg.Dir)
	as.Equal(1, len(one_content), "But there is only one video by mime type")

	as.Equal(test_common.TOTAL_CONTAINERS, len(all_cnts), "Allow it to pull in all containers")
}

func (as *ActionSuite) Test_ManagerTagsMemory() {
	cfg := test_common.InitFakeApp(false)
	man := GetManagerActionSuite(cfg, as)
	as.NoError(man.CreateTag(&models.Tag{ID: "A"}), "couldn't create tag A")
	as.NoError(man.CreateTag(&models.Tag{ID: "B"}), "couldn't create tag B")
	tags, err := man.ListAllTags(0, 3)
	as.NoError(err, "It should be able to list tags")
	as.Equal(len(*tags), 2, "We should have two tags")
}

func (as *ActionSuite) Test_MangerTagsMemoryCRUD() {
	cfg := test_common.InitFakeApp(false)
	man := GetManagerActionSuite(cfg, as)

	t := models.Tag{ID: "A"}
	as.NoError(man.CreateTag(&t), "couldn't create tag A")
	as.NoError(man.UpdateTag(&t), "It should udpate")

	tags, err := man.ListAllTags(0, 3)
	as.NoError(err)
	as.Equal(len(*tags), 1, "We should have one tag")
	man.DestroyTag(t.ID)
	tags_gone, _ := man.ListAllTags(0, 3)
	as.Equal(len(*tags_gone), 0, "Now there should be no tags")
}

func (as *ActionSuite) Test_ManagerMemoryScreens() {
	cfg := test_common.InitFakeApp(false)

	man := GetManagerActionSuite(cfg, as)
	content, err := man.ListAllContent(1, 100)
	as.NoError(err)
	as.Greater(len(*content), 0, "It should have content setup")

	contentArr := *content
	mc := contentArr[0]
	id1, _ := uuid.NewV4()
	id2, _ := uuid.NewV4()

	s1 := models.Screen{ID: id1, Path: "A", Src: "a.txt", ContentID: mc.ID}
	s2 := models.Screen{ID: id2, Path: "B", Src: "b.txt", ContentID: mc.ID}
	mc.Screens = models.Screens{s1, s2}

	// Ensure we actually set the right object in the backing Map
	mem := utils.GetMemStorage()
	mem.ValidContent[mc.ID] = mc
	mem.ValidScreens[s1.ID] = s1
	mem.ValidScreens[s2.ID] = s2

	screens, err := man.ListScreens(mc.ID, 1, 10)
	as.NoError(err)
	as.NotNil(screens)
	as.Equal(2, len(*screens))
	// Check that our single lookup hash is also populated
	for _, screen := range *screens {
		obj, mia := man.GetScreen(screen.ID)
		as.NoError(mia)
		as.Equal(obj.ID, screen.ID)
	}

	allScreens, all_err := man.ListAllScreens(0, 10)
	as.NoError(all_err, "It should work out ok")
	as.Equal(2, len(*allScreens), "We should have 2 screens")
}

func (as *ActionSuite) Test_ManagerMemoryCRU() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)

	// TODO: It should probably validate path exists and access
	c := models.Container{Path: "/a/b"}
	as.NoError(man.CreateContainer(&c), "Did not create container")
	c2 := models.Container{Path: "/a/c"}
	as.NoError(man.CreateContainer(&c2), "Did not create container")
	c_check, c_err := man.GetContainer(c.ID)
	as.NoError(c_err, "We should be able to get back the container")
	as.Equal(c_check.Path, c.Path, "Ensure we are not stomping unset ID data")

	mc := models.Content{Src: "content", ContainerID: nulls.NewUUID(c.ID), NoFile: true}
	as.NoError(man.CreateContent(&mc), "Did not create content correctly")
	mcUp := models.Content{Src: "updated", ID: mc.ID, ContainerID: nulls.NewUUID(c.ID), NoFile: true}
	man.UpdateContent(&mcUp)
	mc_check, m_err := man.GetContent(mc.ID)
	as.NoError(m_err, "It should find this content")
	as.Equal(mc_check.Src, "updated")

	id, _ := uuid.NewV4()
	s1 := models.Screen{Path: "A", Src: "a.txt", ContentID: mc.ID}
	s2 := models.Screen{Path: "B", Src: "b.txt", ContentID: id}
	as.NoError(man.CreateScreen(&s1), "Did not associate screen correctly")
	as.NoError(man.CreateScreen(&s2), "Did not associate screen correctly")

	sCheck, sErr := man.ListScreens(mc.ID, 1, 10)
	as.NoError(sErr, "Failed to list screens")
	as.Equal(len(*sCheck), 1, "It should properly filter screens.")

	s1Update := models.Screen{ID: s1.ID, Path: "C", ContentID: mc.ID}
	as.NoError(man.UpdateScreen(&s1Update))
	s1Check, scErr := man.GetScreen(s1.ID)
	as.NoError(scErr, "Failed to get the screen back")
	as.Equal(s1Check.Path, "C")
}

func (as *ActionSuite) Test_MemoryManagerTags() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)

	aTag := models.Tag{ID: "A"}
	bTag := models.Tag{ID: "B"}
	as.NoError(man.CreateTag(&aTag))
	as.NoError(man.CreateTag(&bTag))

	content := models.Content{Src: "SomethingSomethingDarkside", NoFile: true}
	as.NoError(man.CreateContent(&content))

	as.NoError(man.AssociateTagByID(aTag.ID, content.ID))
	as.NoError(man.AssociateTagByID(bTag.ID, content.ID))
	checkContent, err := man.GetContent(content.ID)
	as.NoError(err)
	as.NotEmpty(checkContent.Tags, "We should have tags")
	as.Equal(len(checkContent.Tags), 2, "There should be two tags")

	// Not in the DB so should not associate
	notExistsTag := models.Tag{ID: "NOPE"}
	as.Error(man.AssociateTagByID(notExistsTag.ID, content.ID))
}

func (as *ActionSuite) Test_MemoryManager_IllegalContainers() {
	cfg := test_common.ResetConfig()
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := GetManager(&ctx)

	notUnderDir := models.Container{Name: "ssl", Path: "/etc"}
	as.Error(man.CreateContainer(&notUnderDir), "Not under the configured directory, rejected")

	upAccess := models.Container{Name: "../../.ssh/", Path: cfg.Dir}
	as.Error(man.CreateContainer(&upAccess), "No up access allowed in names")

	knownDirOk := models.Container{Name: "dir2", Path: cfg.Dir}
	as.NoError(man.CreateContainer(&knownDirOk), "This directory should be ok")

	multiLevelDownOk := models.Container{Name: "screens/screens_sub_dir", Path: cfg.Dir}
	as.NoError(man.CreateContainer(&multiLevelDownOk), "This should exist in the mock data")
}

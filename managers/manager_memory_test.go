package managers

import (
    "contented/internals"
    "contented/models"
    "contented/utils"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
    "os"
    "path/filepath"
)

func (as *ActionSuite) Test_ManagerContainers() {
    internals.InitFakeApp(false)
    ctx := internals.GetContext(as.App)
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

func (as *ActionSuite) Test_ManagerMediaContainer() {
    internals.InitFakeApp(false)
    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx)
    mcs, err := man.ListAllMedia(1, 9001)
    as.NoError(err)

    for _, mc := range *mcs {
        cm, err := man.FindFileRef(mc.ID)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(cm.ID, mc.ID)
    }
}

func (as *ActionSuite) Test_AssignManager() {
    cfg := internals.ResetConfig()
    cfg.UseDatabase = false
    utils.InitConfig(cfg.Dir, cfg)

    mem := ContentManagerMemory{}
    mem.validate = "Memory"
    mem.SetCfg(cfg)
    mem.Initialize()

    memCfg := mem.GetCfg()
    as.NotNil(memCfg, "It should be defined")
    mcs, err := mem.ListAllMedia(1, 9001)
    as.NoError(err)
    as.Greater(len(*mcs), 0, "It should have valid files in the manager")

    cfg.UseDatabase = false
    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx) // New Reference but should have the same count of media
    mcs_2, _ := man.ListAllMedia(1, 9000)

    as.Equal(len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}

func (as *ActionSuite) Test_MemoryManagerPaginate() {
    cfg := internals.InitFakeApp(false)
    cfg.UseDatabase = false

    ctx := internals.GetContextParams(as.App, "/containers", "1", "2")
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
    media_page_1, _ := man.ListMedia(cnt.ID, 1, 4)
    as.Equal(len(*media_page_1), 4, "It should respect page size")

    media_page_3, _ := man.ListMedia(cnt.ID, 3, 4)
    as.Equal(len(*media_page_3), 4, "It should respect page size and get the last page")
    as.NotEqual((*media_page_3)[3].ID, (*media_page_1)[3].ID, "Ensure it actually paged")

    // Last container pagination check
    l_cnts, _ := man.ListContainers(4, 1)
    as.Equal(1, len(*l_cnts), "It should still return only as we are on the last page")
    l_cnt := (*l_cnts)[0]
    as.Equal(expect_len[l_cnt.Name], l_cnt.Total, "There are 3 entries in the ordered test data last container")
}

func (as *ActionSuite) Test_ManagerInitialize() {
    internals.InitFakeApp(false)

    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx)
    as.NotNil(man, "It should have a manager defined after init")

    containers, err := man.ListContainersContext()
    as.NoError(err, "It should list all containers")
    as.NotNil(containers, "It should have containers")
    as.Equal(len(*containers), internals.TOTAL_CONTAINERS, "Unexpected container count")

    // Memory test working
    for _, c := range *containers {
        // fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
        media, err := man.ListMediaContext(c.ID)
        as.NoError(err)
        as.NotNil(media)

        media_len := len(*media)
        // fmt.Printf("Media length was %d\n", media_len)
        as.Greater(media_len, 0, "There should be a number of media")
        as.Equal(expect_len[c.Name], media_len, "It should have this many instances: "+c.Name)
        as.Greater(c.Total, 0, "All of them should have a total assigned")
    }
}

func (as *ActionSuite) Test_MemoryManagerSearch() {
    internals.InitFakeApp(false)

    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx)
    as.NotNil(man, "It should have a manager defined after init")

    containers, err := man.ListContainersContext()
    as.NoError(err, "It should list all containers")
    as.NotNil(containers, "It should have containers")
    as.Equal(len(*containers), internals.TOTAL_CONTAINERS, "Wrong number of containers found")

    mcs, total, err := man.SearchMedia("donut", 1, 20, "", "")
    as.NoError(err, "Can we search in the memory manager")
    as.Equal(len(*mcs), 1, "One donut should be found")
    as.Equal(total, len(*mcs), "It should get the total right")

    mcs_1, _, err_1 := man.SearchMedia("Large", 1, 6, "", "")
    as.NoError(err_1, "Can we search in the memory manager")
    as.Equal(3, len(*mcs_1), "One donut should be found")

    all_mc, _, err_all := man.SearchMedia("", 0, 9000, "", "")
    as.NoError(err_all, "Can in search everything")
    as.Equal(len(*all_mc), internals.TOTAL_MEDIA, "The Kitchen sink")
}

func (as *ActionSuite) Test_MemoryManagerSearchMulti() {
    // Test that a search restricting containerID works
    // Test that search restricting container and text works
    internals.InitFakeApp(false)
    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx)

    // Ensure we initialized with a known search
    as.Equal(man.CanEdit(), false)
    mcs, total, err := man.SearchMedia("donut", 1, 20, "", "")
    as.NoError(err, "Can we search in the memory manager")
    as.Equal(len(*mcs), 1, "One donut should be found")
    as.Equal(total, len(*mcs), "It should get the total right")

    cnts, eep := man.ListContainers(0, 10)
    as.NoError(eep, "It should have 4 containers")
    as.Greater(len(*cnts), 1, "We should have containers")

    allMedia, errAll := man.ListAllMedia(0, 50)
    as.Greater(len(*allMedia), 0, "We should have media")
    as.NoError(errAll)

    all_media, wild_total, _ := man.SearchMedia("", 0, 40, "", "")
    as.Greater(wild_total, 0)
    as.Equal(len(*all_media), wild_total)

    video_media, vid_total, _ := man.SearchMedia("", 0, 40, "", "video")
    as.Equal(vid_total, 1)
    as.Equal(len(*video_media), vid_total)
    vs := *video_media
    as.Equal(vs[0].Src, "donut.mp4")

    for _, cnt := range *cnts {
        if cnt.Name == "dir1" {
            _, no_total, n_err := man.SearchMedia("donut", 1, 20, cnt.ID.String(), "")
            as.NoError(n_err)
            as.Equal(no_total, 0, "It should not be in this directory")
        }
        if cnt.Name == "dir2" {
            yes_match, y_total, r_err := man.SearchMedia("donut", 1, 20, cnt.ID.String(), "")
            as.NoError(r_err)
            as.Equal(y_total, 1, "We did not find the expected media")

            movie := (*yes_match)[0]
            as.Equal(movie.Src, "donut.mp4")

            _, imgCount, _ := man.SearchMedia("", 0, 20, cnt.ID.String(), "image")
            as.Equal(imgCount, 2, "It should filter out the donut this time")
        }
        if cnt.Name == "dir3" {
            has_media, _, err := man.SearchMedia("", 0, 1, cnt.ID.String(), "")
            as.NoError(err, "We should have media")
            as.Greater(len(*has_media), 0)
        }
    }
}


func (as *ActionSuite) Test_MemoryPreviewInitialization() {
    cfg := internals.ResetConfig()
    cfg.MaxSearchDepth = 1
    utils.SetupMediaMatchers(cfg, "", "video", "DS_Store", "")
    utils.SetCfg(*cfg)

    // Create a fake file that would sub in by name for a preview
    var testDir, _ = envy.MustGet("DIR")
    srcDir := filepath.Join(testDir, "dir2")
    dstDir := utils.GetPreviewDst(srcDir)
    testFile := "donut.mp4"

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
    as.Contains(fqPath, "donut.mp4.png")
    f.Sync()

    // Checks that if a preview exists
    cnts, media, _ := utils.PopulateMemoryView(cfg.Dir)
    as.Equal(1, len(cnts), "We should only pull in containers that have media")
    as.Equal(len(media), 1, "But there is only one video by mime type")
    for _, mc := range media {
        as.Equal("/container_previews/donut.mp4.png", mc.Preview)
    }

    cfg.ExcludeEmptyContainers = false
    all_cnts, one_media, _ := utils.PopulateMemoryView(cfg.Dir)
    as.Equal(1, len(one_media), "But there is only one video by mime type")

    as.Equal(internals.TOTAL_CONTAINERS, len(all_cnts), "Allow it to pull in all containers")
}


func (as *ActionSuite) Test_ManagerTagsMemory() {
    cfg := internals.InitFakeApp(false)
    man := GetManagerActionSuite(cfg, as)
    as.NoError(man.CreateTag(&models.Tag{Name: "A",}), "couldn't create tag A")
    as.NoError(man.CreateTag(&models.Tag{Name: "B",}), "couldn't create tag B")
    tags, err := man.ListAllTags(0, 3)
    as.NoError(err, "It should be able to list tags")
    as.Equal(len(*tags), 2, "We should have two tags")
}

func (as *ActionSuite) Test_MangerTagsMemoryCRUD() {
    cfg := internals.InitFakeApp(false)
    man := GetManagerActionSuite(cfg, as)

    t := models.Tag{Name: "A",}
    as.NoError(man.CreateTag(&t), "couldn't create tag A")
    t.Name = "Changed"
    as.NoError(man.UpdateTag(&t), "It should udpate")

    tags, err := man.ListAllTags(0, 3)
    as.NoError(err)
    as.Equal(len(*tags), 1, "We should have one tag")
    as.Equal((*tags)[0].Name, "Changed", "It should update")
    man.DeleteTag(&t)
    tags_gone, _ := man.ListAllTags(0, 3)
    as.Equal(len(*tags_gone), 0, "Now there should be no tags")
}


func (as *ActionSuite) Test_ManagerMemoryScreens() {
    cfg := internals.InitFakeApp(false)

    man := GetManagerActionSuite(cfg, as)
    media, err := man.ListAllMedia(1, 100)
    as.NoError(err)
    as.Greater(len(*media), 0, "It should have media setup")

    mediaArr := *media
    mc := mediaArr[0]
    id1, _ := uuid.NewV4()
    id2, _ := uuid.NewV4()

    s1 := models.Screen{ID: id1, Path: "A", Src: "a.txt", MediaID: mc.ID}
    s2 := models.Screen{ID: id2, Path: "B", Src: "b.txt", MediaID: mc.ID}
    mc.Screens = models.Screens{s1, s2}

    // Ensure we actually set the right object in the backing Map
    mem := utils.GetMemStorage()
    mem.ValidMedia[mc.ID] = mc
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
    internals.InitFakeApp(false)
    ctx := internals.GetContext(as.App)
    man := GetManager(&ctx)

    // TODO: It should probably validate path exists and access
    c := models.Container{Path: "/a/b"}
    as.NoError(man.CreateContainer(&c), "Did not create container")
    c2 := models.Container{Path: "/a/c"}
    as.NoError(man.CreateContainer(&c2), "Did not create container")
    c_check, c_err := man.GetContainer(c.ID)
    as.NoError(c_err, "We should be able to get back the container")
    as.Equal(c_check.Path, c.Path, "Ensure we are not stomping unset ID data")

    mc := models.MediaContainer{Src: "media", ContainerID: nulls.NewUUID(c.ID)}
    as.NoError(man.CreateMedia(&mc), "Did not create media correctly")
    mcUp := models.MediaContainer{Src: "updated", ID: mc.ID}
    man.UpdateMedia(&mcUp)
    mc_check, m_err := man.GetMedia(mc.ID)
    as.NoError(m_err, "It should find this media")
    as.Equal(mc_check.Src, "updated")

    id, _ := uuid.NewV4()
    s1 := models.Screen{Path: "A", Src: "a.txt", MediaID: mc.ID}
    s2 := models.Screen{Path: "B", Src: "b.txt", MediaID: id}
    as.NoError(man.CreateScreen(&s1), "Did not associate screen correctly")
    as.NoError(man.CreateScreen(&s2), "Did not associate screen correctly")

    sCheck, sErr := man.ListScreens(mc.ID, 1, 10)
    as.NoError(sErr, "Failed to list screens")
    as.Equal(len(*sCheck), 1, "It should properly filter screens.")

    s1Update := models.Screen{ID: s1.ID, Path: "C", MediaID: mc.ID}
    as.NoError(man.UpdateScreen(&s1Update))
    s1Check, scErr := man.GetScreen(s1.ID)
    as.NoError(scErr, "Failed to get the screen back")
    as.Equal(s1Check.Path, "C")
}

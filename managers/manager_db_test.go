package managers

import (
    "contented/internals"
    "contented/models"
    "fmt"
    "github.com/gobuffalo/pop/v6"
)

// A basic DB search (ilike matching)
func (as *ActionSuite) Test_DbManagerSearch() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)
    man := GetManagerActionSuite(cfg, as)
    as.Equal(man.CanEdit(), true, "It should be a DB manager")

    cnt1, content1 := internals.GetContentByDirName("dir1")
    cnt2, content2 := internals.GetContentByDirName("dir2")
    c1_err := man.CreateContainer(cnt1)
    as.NoError(c1_err)
    c2_err := man.CreateContainer(cnt2)
    as.NoError(c2_err)
    for _, mc := range content1 {
        man.CreateContent(&mc)
    }
    for _, mc := range content2 {
        man.CreateContent(&mc)
        if mc.Src == internals.VIDEO_FILENAME {
            man.CreateScreen(&models.Screen{ContentID: mc.ID, Src: "screen1"})
            man.CreateScreen(&models.Screen{ContentID: mc.ID, Src: "screen2"})
        }
    }
    mcs, _, err := man.SearchContent("Large", 1, 20, "", "")
    as.NoError(err, "It should be able to search")
    as.NotNil(mcs, "It should be")
    as.Equal(3, len(*mcs), fmt.Sprintf("We should have 3 large images with an ilike %s", mcs))

    mcs_d, vsTotal, vErr := man.SearchContent("donut", 1, 10, "", "")
    as.NoError(vErr, "Video error by name search failed")
    as.Equal(1, vsTotal, "We should be able to find donut.mp4 with an ilike")
    mc_donut := (*mcs_d)[0]
    as.Equal(2, len(mc_donut.Screens), fmt.Sprintf("It should load two screens %s", mc_donut.Screens))

    vids, vidTotal, dbErr := man.SearchContent("", 1, 40, "", "video")
    as.NoError(dbErr, "Should search content type")
    as.Equal(1, vidTotal, "The total count for videos is 1")
    as.Equal(1, len(*vids), "We should have one result")

    all_mcs, total, err := man.SearchContent("", 1, 10, cnt1.ID.String(), "")
    as.NoError(err, "It should be able to empty search")
    as.Equal(12, total, "The total count for this dir is 12")
    as.Equal(10, len(*all_mcs), "But we limited the pagination")
}

func (as *ActionSuite) Test_DbManagerMultiSearch() {
    // Test that a search restricting containerID works
    // Test that search restricting container and text works
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)

    man := GetManagerActionSuite(cfg, as)
    as.Equal(man.CanEdit(), true)

    cnt1, content1, err1 := internals.CreateContentByDirName("dir1")
    as.NoError(err1)
    as.Greater(len(content1), 1)

    cnt2, content2, err2 := internals.CreateContentByDirName("dir2")
    as.NoError(err2)
    as.Greater(len(content2), 1)

    found, count, err := man.SearchContent(content1[1].Src, 0, 10, cnt1.ID.String(), "")
    as.Equal(len(*found), 1, "We should have found our item")
    as.Equal(count, 1)
    as.NoError(err)

    _, n_count, n_err := man.SearchContent("blah", 0, 10, cnt1.ID.String(), "")
    as.Equal(n_count, 0, "It should not find this the content name is invalid")
    as.NoError(n_err)

    _, not_in_cnt_count, not_err := man.SearchContent(content1[1].Src, 0, 10, cnt2.ID.String(), "")
    as.Equal(not_in_cnt_count, 0, "It should not find this valid content as it is not in the container")
    as.NoError(not_err)
}


func (as *ActionSuite) Test_ManagerDB() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)

    cnt, content := internals.GetContentByDirName("dir1")
    as.Equal("dir1", cnt.Name, "It should be the right dir")
    as.Equal(12, cnt.Total, "The container total should be this for dir1")
    as.Equal(12, len(content))

    c_err := models.DB.Create(cnt)
    as.NoError(c_err)
    for _, mc := range content {
        models.DB.Create(&mc)
    }

    man := GetManagerActionSuite(cfg, as)
    q_content, err := man.ListAllContent(0, 14)
    as.NoError(err, "We should be able to list")
    as.Equal(len(*q_content), 12, "there should be 12 results")

    lim_content, _ := man.ListAllContent(0, 3)
    as.Equal(3, len(*lim_content), "The DB should be setup with 10 items")
}


func (as *ActionSuite) Test_ManagerTagsDB() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)
    man := GetManagerActionSuite(cfg, as)

    as.NoError(man.CreateTag(&models.Tag{Name: "A",}), "couldn't create tag A")
    as.NoError(man.CreateTag(&models.Tag{Name: "B",}), "couldn't create tag B")
    tags, err := man.ListAllTags(0, 3)
    as.NoError(err, "It should be able to list tags")
    as.Equal(len(*tags), 2, "We should have two tags")
}

func (as *ActionSuite) Test_ManagerTagsDBCRUD() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)
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
    as.Equal(len(*tags_gone), 0, "No tags should be in the DB")
}


    /*
func (as *ActionSuite) Test_ManagerAssociateTagsDB() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)
    man := GetManagerActionSuite(cfg, as)

    // hate
    t1 := models.Tag{Name: "A",}
    t2 := models.Tag{Name: "B",}
    man.CreateTag(&t1)
    man.CreateTag(&t2)
    mc := models.Content{Src: "A", Preview: "p", ContentType: "video"}
    mc.Tags = models.Tags{t1, t2,}
    man.CreateContent(&mc)

    s := models.Screen{Src: "screen1", ContentID: mc.ID}
    man.CreateScreen(&s)
    mc.Screens = models.Screens{s,}
    man.UpdateContent(&mc)
    // as.Equal(len(mc.Tags), 0, "There should be no tags at this point")

    tags, t_err := man.ListAllTags(0, 10)
    as.NoError(t_err, "We should be able to list tags.")
    as.Equal(2, len(*tags), fmt.Sprintf("There should be two tags %s", mc))

    screens, s_err := man.ListScreens(mc.ID, 0, 10)
    as.NoError(s_err, "Screens should list")
    as.Equal(1, len(*screens), "We should have a screen associated")

    // TODO: list screens
    //screens, s_err := man.(0, 10)

    // TODO: ok so the damn eager loading is just not working?
    tCheck, _ := man.GetContent(mc.ID)
    as.Equal(2, len(tCheck.Tags), fmt.Sprintf("Wat %s", tCheck))
    err := man.AssociateTagByID(t1.ID, mc.ID)
    as.NoError(err, "We shouldn't have an issue associating this")
    mcCheck, mc_err := man.GetContent(mc.ID)
    as.NoError(mc_err, "We should be able to load back the content")
    as.Equal(1, len(mcCheck.Tags), fmt.Sprintf("There should be a new tag %s", mcCheck))
}
    */


func (as *ActionSuite) Test_ManagerDBPreviews() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)
    man := GetManagerActionSuite(cfg, as)

    mc1 := models.Content{Src: "A", Preview: "p", ContentType: "video"}
    mc2 := models.Content{Src: "B", Preview: "p", ContentType: "video"}
    mc3 := models.Content{Src: "C", Preview: "p", ContentType: "video"}
    man.CreateContent(&mc1)
    man.CreateContent(&mc2)
    man.CreateContent(&mc3)
    as.NotZero(mc1.ID)

    p1 := models.Screen{Src: "fake1", Idx: 0, ContentID: mc1.ID}
    p2 := models.Screen{Src: "fake2.png", Idx: 1, ContentID: mc1.ID}
    p3 := models.Screen{Src: "fake3.png", Idx: 1, ContentID: mc2.ID}

    man.CreateScreen(&p1)
    man.CreateScreen(&p2)
    man.CreateScreen(&p3)

    previewList, err := man.ListScreens(mc1.ID, 1, 10)
    as.NoError(err)
    as.Equal(len(*previewList), 2, "We should have two previews")

    previewOne, p_err := man.ListScreens(mc2.ID, 1, 10)
    as.NoError(p_err)
    as.Equal(len(*previewOne), 1, "Now there should be 1")

    p4 := models.Screen{Src: "fake4.png", Idx: 1, ContentID: mc2.ID}
    c_err := man.CreateScreen(&p4)
    as.NoError(c_err)

    p4_check, p4_err := man.GetScreen(p4.ID)
    as.NoError(p4_err, "Failed to pull back the screen by ID"+p4.ID.String())
    as.Equal(p4_check.Src, p4.Src)
}


func (as *ActionSuite) Test_ManagerDBSearchScreens() {
    models.DB.TruncateAll()
    cfg := internals.InitFakeApp(true)

    man := ContentManagerDB{cfg: cfg}
    man.GetConnection = func() *pop.Connection {
        return models.DB
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
    as.NoError(s_err, "It shouldn't error out")
    as.NotNil(screens, "No screens were returned")
    as.Equal(2, len(screens), "It should load all the screens but only two of these have screens")

    as.Equal(1, len(screens[mc1.ID]), "MC1 has 1 screen")
    as.Equal(2, len(screens[mc3.ID]), "MC3 has 2 screens")

    // Test that an image will not load previews
    content_2 := models.Contents{mc2, mc4}
    screens_2, s2_err := man.LoadRelatedScreens(&content_2)
    as.NoError(s2_err, "It shouldn't error out")
    as.Equal(1, len(screens_2), "It should load all the screens for mc2 but EXCLUDE mc4")
}

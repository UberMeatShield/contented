package actions

import (
    //"path/filepath"
    "os"
    "io/ioutil"
    "log"
	"contented/models"
	"contented/utils"
    "time"
	//"os"
	//"testing"
	//"github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/envy"
	// "github.com/gobuffalo/suite"
)

func GetScreens() (*models.Container, models.MediaContainers){
	dir, _ := envy.MustGet("DIR")
    appCfg.Dir = dir
    cnts := utils.FindContainers(dir)

    var screenCnt *models.Container = nil
    for _, c := range cnts {
        if c.Name == "screens" {
            screenCnt = &c
        }
    }
    if screenCnt == nil {
        log.Panic("Could not find the screens directory")
    }
    media := utils.FindMedia(*screenCnt, 4, 0)
    screenCnt.Total = len(media)
    return screenCnt, media
}

func (as *ActionSuite) Test_InitialCreation() {
	dir, _ := envy.MustGet("DIR")
    as.NotEmpty(dir, "The test must specify a directory to run on")

    err := CreateInitialStructure(dir)
    as.NoError(err, "It should successfully create the full DB setup")

    cnts := models.Containers{}
    as.DB.All(&cnts)

    media := models.MediaContainers{}
    as.DB.All(&media)
    as.Equal(len(media), 24, "The mocks have a specific expected number of items")
}

func (as *ActionSuite) Test_ImgPreview() {
	dir, _ := envy.MustGet("DIR")
    appCfg.Dir = dir

    id, _ := uuid.NewV4()
    cnt := models.Container{
        Path: dir,
        Name: "screens",
        ID: id,
        Active: true,
    }
    media := utils.FindMedia(cnt, 4, 0)
    as.Equal(len(media), 4)

    // Basic sanity check that things exist and can preview
    for _, m := range media {
        fq_path, err := GetFilePathInContainer(m.Src, cnt.Name)
        as.NoError(err, "It should not fail getting a full file path" + m.Src)

        f, open_err := os.Open(fq_path)
        as.NoError(open_err, "And all the paths should exist")
        as.Equal(utils.ShouldCreatePreview(f, 0), true, "Create all previews with no min size")
    }
}


func (as *ActionSuite) Test_CreatePreview() {
    // Create one that is a fail
    c_pt, media := GetScreens()
    err := ClearContainerPreviews(c_pt)
    as.NoError(err, "It should nuke out the preview directory")
    as.Equal(len(media), 4, "There should be 4 of these in the screens dir")

    dstPath := GetContainerPreviewDst(c_pt)
    dir_err := utils.MakePreviewPath(dstPath)
    as.NoError(dir_err, "Did we createa preview path")

    // Create one that does create a preview
    for _, mc := range media {
        preview_path, err := CreateMediaPreview(c_pt, &mc, 0)
        as.NoError(err, "It should be ble to create previews")
        as.NotEqual(preview_path, "", "The path should be defined")
    }

    previews, read_err := ioutil.ReadDir(dstPath)
    as.Equal(len(previews), 4, "It should create 4 previews")
    as.NoError(read_err, "It should be able to read the directory")
}

func (as *ActionSuite) Test_CreateContainerPreviews() {
    // Get a local not in DB setup for the container and media
    c_pt, media := GetScreens()

    err := models.DB.TruncateAll()
    as.NoError(err, "It should dump the DB")
    clear_err := ClearContainerPreviews(c_pt)
    as.NoError(clear_err, "And we should clear the preview dir")

    // Ensure that we do have a valid destination path
    dstPath := GetContainerPreviewDst(c_pt)
    dir_err := utils.MakePreviewPath(dstPath)
    as.NoError(dir_err, "It should have created the preview path")
    empty, read_err := ioutil.ReadDir(dstPath)
    as.Equal(len(empty), 0, "It should start completely empty")

    // Now add the data into the database
    c_err := models.DB.Create(c_pt)
    as.NoError(c_err)
    for _, mc := range media {
        mc.ContainerID = nulls.NewUUID(c_pt.ID)
        mc_err := models.DB.Create(&mc)
        as.NoError(mc_err)
        as.Equal(mc.Preview, "", "There should be no preview at this point")
    }
    // Create a bunch of previews
    p_err := CreateContainerPreviews(c_pt, 0)
    as.NoError(p_err, "An error happened creating the previews")

    previews, read_err := ioutil.ReadDir(dstPath)
    as.Equal(len(previews), 4, "It should create 4 previews")
    as.NoError(read_err, "It should be able to read the directory")

    media_check := models.MediaContainers{}
    models.DB.Where("container_id = ?", c_pt.ID).All(&media_check)
    as.Equal(len(media_check), 4, "We should just have 4 things to check")

    for _, mc_check := range media_check {
        as.NotEqual(mc_check.Preview, "", "It should now have a preview")
    }
}

// hate
func (as *ActionSuite) TestAsyncPreviews() {
    c_pt, media := GetScreens()
    err := models.DB.TruncateAll()
    as.NoError(err, "It should dump the DB")
    clear_err := ClearContainerPreviews(c_pt)
    as.NoError(clear_err, "And we should clear the preview dir")

    dstPath := GetContainerPreviewDst(c_pt)
    dir_err := utils.MakePreviewPath(dstPath)
    as.NoError(dir_err, "Did we createa preview path")

    as.Greater(len(media), 3, "There should be some things to iterate over")
    prevs := make(chan utils.ProcessingResult, c_pt.Total)
    as.Equal(c_pt.Total, 4, "It should have four entries")
    as.Equal(c_pt.Total, len(media), "Definitely it should equal our media")

    as.Equal(c_pt.Name, "screens", "It should have a screens directory")
    for _, mc := range media {
        as.NotEqual(mc.Src, "", "It should have a source path")
        go AsyncMediaPreview(c_pt, &mc, 0, prevs)
    }
    sleep := 1200
    sleepTime := time.Duration(sleep) * time.Millisecond
    time.Sleep(sleepTime)

    as.Equal(len(prevs), 4, "It should have four elements at this point")
    wat := <-prevs
    as.NoError(wat.Err, "It should not have an error")
    as.NotEqual(wat.Preview, "It should have a preview location")
}


func (as *ActionSuite) Test_PreviewAllData() {
    err := models.DB.TruncateAll()
    as.NoError(err, "Couldn't clean the DB")

	dir, _ := envy.MustGet("DIR")
    as.NotEmpty(dir, "The test must specify a directory to run on")

    appCfg.Dir = dir
    c_err := CreateInitialStructure(dir)
    as.NoError(c_err, "Failed to build out the initial database")

    all_created_err := CreateAllPreviews(1000 * 1024)
    as.NoError(all_created_err, "Failed to create all previews")
}

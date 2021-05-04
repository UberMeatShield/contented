package actions

import (
    //"path/filepath"
    "os"
    "io/ioutil"
    "log"
	"contented/models"
	"contented/utils"
	//"os"
	//"testing"
	//"github.com/gobuffalo/buffalo"
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

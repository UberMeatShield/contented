package managers

import (
    "os"
    "io/ioutil"
	"contented/models"
	"contented/utils"
	"contented/internals"
    "encoding/json"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/nulls"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/suite"
)

var TOTAL_IN_SCREENS = 4

type ActionSuite struct {
    *suite.Action
}

func GetScreens() (*models.Container, models.MediaContainers) {
    return internals.GetMediaByDirName("screens")
}

func SetupScreensPreview(as *ActionSuite) (*models.Container, models.MediaContainers) {
    c_pt, media := GetScreens()
    err := models.DB.TruncateAll()
    as.NoError(err, "It should dump the DB")
    clear_err := ClearContainerPreviews(c_pt)
    as.NoError(clear_err, "And we should clear the preview dir")

    dstPath := GetContainerPreviewDst(c_pt)
    dir_err := utils.MakePreviewPath(dstPath)
    as.NoError(dir_err, "Did we createa preview path")

    empty, read_err := ioutil.ReadDir(dstPath)
    as.Equal(len(empty), 0, "It should start completely empty")
    as.NoError(read_err, "It should be able to read the dst directory")
    return c_pt, media
}

func (as *ActionSuite) Test_InitialCreation() {
	dir, _ := envy.MustGet("DIR")
    as.NotEmpty(dir, "The test must specify a directory to run on")

    cfg := utils.GetCfg()
    cfg.Dir = dir

    err := CreateInitialStructure(cfg)
    as.NoError(err, "It should successfully create the full DB setup")

    cnts := models.Containers{}
    as.DB.All(&cnts)

    media := models.MediaContainers{}
    as.DB.All(&media)
    as.Equal(27, len(media), "The mocks have a specific expected number of items")
}

func (as *ActionSuite) Test_CfgIncExcFiles() {
    models.DB.TruncateAll()

    // Exclude all images
	dir, _ := envy.MustGet("DIR")
    cfg := utils.GetCfg()
    cfg.Dir = dir
    cfg.ExcFiles = utils.CreateMatcher("DS_STORE", "image", "OR")
    err := CreateInitialStructure(cfg)
    as.NoError(err)

    media := models.MediaContainers{}
    as.DB.All(&media)
    dbg, _ := json.Marshal(media)
    as.Equal(1, len(media), "There should be one match: " + string(dbg))
    as.Equal(media[0].ContentType, "video/mp4", "It should be the video")

    clear_err := models.DB.TruncateAll()
    as.NoError(clear_err)
    cfg.ExcFiles = utils.ExcludeNoFiles
    cfg.IncFiles = utils.CreateMatcher("", "jpeg", "AND")

    err_png := CreateInitialStructure(cfg)
    as.NoError(err_png)

    jpeg_media := models.MediaContainers{}
    as.DB.All(&jpeg_media)
    as.Equal(1, len(media), "There is one jpeg")
}

func (as *ActionSuite) Test_ImgShouldCreatePreview() {
	dir, _ := envy.MustGet("DIR")

    cfg := utils.GetCfg()
    cfg.Dir = dir

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
        fq_path, err := utils.GetFilePathInContainer(m.Src, cnt.Name)
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
    as.Equal(TOTAL_IN_SCREENS, len(media), "There should be 4 of these in the screens dir")

    dstPath := GetContainerPreviewDst(c_pt)
    dir_err := utils.MakePreviewPath(dstPath)
    as.NoError(dir_err, "Did we createa preview path")

    // Create one that does create a preview
    cfg := utils.GetCfg()
    cfg.PreviewOverSize = 0
    for _, mc := range media {
        preview_path, err := CreateMediaPreview(c_pt, &mc)
        as.NoError(err, "It should be ble to create previews")
        as.NotEqual(preview_path, "", "The path should be defined")
    }

    previews, read_err := ioutil.ReadDir(dstPath)
    as.Equal(TOTAL_IN_SCREENS, len(previews), "It should create 4 previews")
    as.NoError(read_err, "It should be able to read the directory")
}

func (as *ActionSuite) Test_CreateContainerPreviews() {
    // Get a local not in DB setup for the container and media
    // Create a bunch of previews
    cfg := internals.ResetConfig()
    cfg.UseDatabase = true
    cfg.PreviewOverSize = 0

    // Now add the data into the database
    c_pt, media := SetupScreensPreview(as)
    c_err := models.DB.Create(c_pt)
    as.NoError(c_err)
    
    nulls.NewUUID(c_pt.ID)
    as.Greater(len(media), 0, "We should have media")

    // Check that we have a container preview at this point
    expect_c_preview := ""
    for _, mc := range media {
        mc.ContainerID = nulls.NewUUID(c_pt.ID)
        mc_err := models.DB.Create(&mc)
        as.NoError(mc_err)
        as.Equal(mc.Preview, "", "There should be no preview at this point")
        if expect_c_preview == "" {
            expect_c_preview = "/preview/" + mc.ID.String()
        }
    }
    man := GetManagerActionSuite(cfg, as)
    cnts, c_err := man.ListContainers(0, 2)
    as.Equal(len(*cnts), 1, "It should have containers")
    as.NoError(c_err)

    p_err := CreateContainerPreviews(c_pt, man)
    as.Equal(expect_c_preview, c_pt.PreviewUrl, "It should assign a mc preview to the container")
    as.NoError(p_err, "An error happened creating the previews")
    dstPath := GetContainerPreviewDst(c_pt)
    previews, read_err := ioutil.ReadDir(dstPath)
    as.Equal(TOTAL_IN_SCREENS, len(previews), "It should create 6 previews")
    as.NoError(read_err, "It should be able to read the directory")

    // Validate that the media was updated in the DB
    media_check := models.MediaContainers{}
    models.DB.Where("container_id = ?", c_pt.ID).All(&media_check)
    as.Equal(TOTAL_IN_SCREENS, len(media_check), "We should just have 6 things to check")
    for _, mc_check := range media_check {
        as.NotEqual(mc_check.Preview, "", "It should now have a preview")
    }
}

func (as *ActionSuite) Test_AsyncContainerPreviews() {
    c_pt, media := SetupScreensPreview(as)

    // On the DB side these would then need to be updated in the DB for linkage
    cfg := utils.GetCfg()
    cfg.PreviewOverSize = 0
    previews, err := CreateMediaPreviews(c_pt, media)
    as.NoError(err, "It should be able to create all previews successfully")

    as.Equal(len(previews), len(media), "With size zero we should have 4 previews")
    for _, p_mc := range previews {
        as.NotEqual(p_mc.Preview, "", "All results should have a preview")
    }
    // TODO: Validate the previews are created on disk
    // Map the results back to the media containers
    // Maybe just return them vs update the DB
}

func (as *ActionSuite) Test_PreviewAllData() {
    err := models.DB.TruncateAll()
    as.NoError(err, "Couldn't clean the DB")

	dir, _ := envy.MustGet("DIR")
    as.NotEmpty(dir, "The test must specify a directory to run on")

    cfg := utils.GetCfg()
    cfg.UseDatabase = true
    cfg.Dir = dir

    c_err := CreateInitialStructure(cfg)
    man := GetManagerActionSuite(cfg, as)

    cnts, c_err := man.ListContainers(0, 3)
    as.Equal(len(*cnts), 3, "It should have containers")
    as.NoError(c_err)

    as.NoError(c_err, "Failed to build out the initial database")
    as.Equal(true, man.CanEdit(), "It should be able to edit")

    cfg.PreviewOverSize = 0
    all_created_err := CreateAllPreviews(man)
    as.NoError(all_created_err, "Failed to create all previews")
}

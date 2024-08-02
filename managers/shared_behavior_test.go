package managers

import (
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/suite/v4"
	"github.com/gofrs/uuid"
)

var TOTAL_IN_SCREENS = 4

type ActionSuite struct {
	*suite.Action
}

// TODO: This naming is now bad with preview screens
func GetScreens() (*models.Container, models.Contents) {
	return test_common.GetContentByDirName("screens")
}

func SetupScreensPreview(as *ActionSuite) (*models.Container, models.Contents) {
	c_pt, content := GetScreens()
	err := models.DB.TruncateAll()
	as.NoError(err, "It should dump the DB")
	clear_err := utils.ClearContainerPreviews(c_pt)
	as.NoError(clear_err, "And we should clear the preview dir")

	dstPath := utils.GetContainerPreviewDst(c_pt)
	dir_err := utils.MakePreviewPath(dstPath)
	as.NoError(dir_err, "Did we createa preview path")

	empty, read_err := ioutil.ReadDir(dstPath)
	as.Equal(len(empty), 0, "It should start completely empty")
	as.NoError(read_err, "It should be able to read the dst directory")
	return c_pt, content
}

func (as *ActionSuite) Test_InitialCreation() {
	dir, _ := envy.MustGet("DIR")
	as.NotEmpty(dir, "The test must specify a directory to run on")

	cfg := test_common.ResetConfig()
	as.True(cfg.ExcContent(".DS_Store", "application/octet-stream"), "This should not be allowed")

	err := CreateInitialStructure(cfg)
	as.NoError(err, "It should successfully create the full DB setup")

	cnts := models.Containers{}
	as.DB.All(&cnts)
	as.Equal(test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(cnts), "The mocks have a specific expected number of items")

	content := models.Contents{}
	as.DB.All(&content)
	as.Equal(test_common.TOTAL_MEDIA, len(content), "The mocks have a specific expected number of items")
}

func (as *ActionSuite) Test_CfgIncExcContent() {
	models.DB.TruncateAll()

	// Exclude all images
	dir, _ := envy.MustGet("DIR")
	cfg := test_common.ResetConfig()
	cfg.Dir = dir
	nope := utils.CreateContentMatcher("DS_Store", "image|text", "OR")

	as.True(nope(".DS_Store", "image/png"))
	cfg.ExcContent = nope
	utils.SetCfg(*cfg)
	err := CreateInitialStructure(cfg)
	as.NoError(err)

	content := models.Contents{}
	as.DB.All(&content)
	dbg, _ := json.Marshal(content)
	as.Equal(test_common.TOTAL_VIDEO, len(content), fmt.Sprintf("There should be three: %s", dbg))
	as.Equal(content[0].ContentType, "video/mp4", "It should be the video")

	clear_err := models.DB.TruncateAll()
	as.NoError(clear_err)
	cfg.ExcContent = utils.ExcludeNoFiles
	cfg.IncContent = utils.CreateContentMatcher("", "jpeg", "AND")

	err_png := CreateInitialStructure(cfg)
	as.NoError(err_png)

	jpeg_content := models.Contents{}
	as.DB.All(&jpeg_content)
	as.Equal(2, len(jpeg_content), fmt.Sprintf("There are 2 jpeg %s", jpeg_content))
}

func (as *ActionSuite) Test_ImgShouldCreatePreview() {
	dir, _ := envy.MustGet("DIR")

	cfg := utils.GetCfg()
	cfg.Dir = dir

	id, _ := uuid.NewV4()
	cnt := models.Container{
		Path:   dir,
		Name:   "screens",
		ID:     id,
		Active: true,
	}
	content := utils.FindContent(cnt, 4, 0)
	as.Equal(len(content), 4)

	// Basic sanity check that things exist and can preview
	for _, m := range content {
		fq_path, err := utils.GetFilePathInContainer(m.Src, cnt.GetFqPath())
		as.NoError(err, "It should not fail getting a full file path"+m.Src)

		f, open_err := os.Open(fq_path)
		as.NoError(open_err, "And all the paths should exist")
		as.Equal(utils.ShouldCreatePreview(f, 0), true, "Create all previews with no min size")
	}
}

// TODO: This should eventually move to utils/previews_test.go
func (as *ActionSuite) Test_CreatePreview() {
	// Create one that is a fail
	c_pt, content := GetScreens()
	err := utils.ClearContainerPreviews(c_pt)
	as.NoError(err, "It should nuke out the preview directory")
	as.Equal(TOTAL_IN_SCREENS, len(content), "There should be 4 of these in the screens dir")

	dstPath := utils.GetContainerPreviewDst(c_pt)
	dir_err := utils.MakePreviewPath(dstPath)
	as.NoError(dir_err, "Did we createa preview path")

	// Create one that does create a preview
	cfg := utils.GetCfg()
	cfg.PreviewOverSize = 0
	for _, mc := range content {
		preview_path, err := utils.CreateContentPreview(c_pt, &mc)
		as.NoError(err, "It should be ble to create previews")
		as.NotEqual(preview_path, "", "The path should be defined")
	}
	previews, read_err := ioutil.ReadDir(dstPath)
	as.Equal(TOTAL_IN_SCREENS, len(previews), "It should create 4 previews")
	as.NoError(read_err, "It should be able to read the directory")
}

func (as *ActionSuite) Test_CreateBaseTags() {
	err := models.DB.TruncateAll()
	as.NoError(err)
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = true
	cfg.TagFile = filepath.Join(cfg.Dir, "/dir2/tags.txt")
	utils.SetCfg(*cfg)

	man := GetManagerActionSuite(cfg, as)
	as.Equal(man.CanEdit(), true)

	tags, createTagsErr := CreateTagsFromFile(man)
	as.NoError(createTagsErr)
	as.NotNil(tags)
	as.Equal(len(*tags), test_common.TOTAL_TAGS, "It should keep track of what was created")

	tagsCheck, total, lErr := man.ListAllTags(TagQuery{PerPage: 50})
	as.NoError(lErr, "It should not error out listing tags")
	as.Greater(total, 0, "We should have a tag count")
	as.NotNil(tags, "There should be tags")
	as.Equal(test_common.TOTAL_TAGS, len(*tagsCheck), "The tags should be created and queried")

	tagsAgain, retryError := CreateTagsFromFile(man)
	as.NoError(retryError, "We should not have any DB error")
	as.Equal(test_common.TOTAL_TAGS, len(*tagsAgain), "The Tag count should be the same.")
	// Add test to ensure we do not try and create the same tag over and over
	// TODO: Need to do a test that actually checks memory
}

func (as *ActionSuite) Test_CreateContainerPreviews() {
	// Get a local not in DB setup for the container and content
	// Create a bunch of previews
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = true
	cfg.PreviewOverSize = 0

	// Now add the data into the database
	c_pt, content := SetupScreensPreview(as)
	c_err := models.DB.Create(c_pt)
	as.NoError(c_err)

	nulls.NewUUID(c_pt.ID)
	as.Greater(len(content), 0, "We should have content")

	// Check that we have a container preview at this point
	expect_c_preview := ""
	for idx, mc := range content {
		mc.Idx = idx
		mc.ContainerID = nulls.NewUUID(c_pt.ID)
		mc_err := models.DB.Create(&mc)
		as.NoError(mc_err)
		as.Equal(mc.Preview, "", "There should be no preview at this point")
		if expect_c_preview == "" {
			expect_c_preview = "/api/preview/" + mc.ID.String()
		}
	}

	man := GetManagerActionSuite(cfg, as)
	cnts, count, c_err := man.ListContainers(ContainerQuery{Page: 1, PerPage: 2})
	as.Greater(count, 0, "There should be a container")
	as.Equal(len(*cnts), 1, "It should have containers")
	as.NoError(c_err)

	p_err := CreateContainerPreviews(c_pt, man)
	as.Equal(expect_c_preview, c_pt.PreviewUrl, "It should assign a mc preview to the container")

	as.NoError(p_err, "An error happened creating the previews")
	dstPath := utils.GetContainerPreviewDst(c_pt)
	previews, read_err := ioutil.ReadDir(dstPath)
	as.Equal(TOTAL_IN_SCREENS, len(previews), "It should create 6 previews")
	as.NoError(read_err, "It should be able to read the directory")

	// Validate that the content was updated in the DB
	content_check := models.Contents{}
	models.DB.Where("container_id = ?", c_pt.ID).Order("created_at desc").All(&content_check)
	as.Equal(TOTAL_IN_SCREENS, len(content_check), "We should just have 6 things to check")
	for _, mc_check := range content_check {
		as.NotEqual(mc_check.Preview, "", "It should now have a preview")
	}
}

func (as *ActionSuite) Test_AsyncContainerPreviews() {
	c_pt, content := SetupScreensPreview(as)

	// On the DB side these would then need to be updated in the DB for linkage
	cfg := utils.GetCfg()
	cfg.PreviewOverSize = 0
	previews, err := CreateContentPreviews(c_pt, content)
	as.NoError(err, "It should be able to create all previews successfully")

	as.Equal(len(previews), len(content), "With size zero we should have 4 previews")
	for _, p_mc := range previews {
		as.NotEqual(p_mc.Preview, "", "All results should have a preview")
	}
	// TODO: Validate the previews are created on disk
	// Map the results back to the content containers
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
	cfg.ExcContent = utils.CreateContentMatcher("corrupted", "", cfg.ExcludeOperator)

	c_err := CreateInitialStructure(cfg)
	man := GetManagerActionSuite(cfg, as)

	cquery := ContainerQuery{Page: 1, PerPage: 3}
	cnts, count, c_err := man.ListContainers(cquery)
	as.Greater(count, 1, "There should be a positive count")
	as.Equal(cquery.PerPage, len(*cnts), "It should have containers")
	as.NoError(c_err)

	as.NoError(c_err, "Failed to build out the initial database")
	as.Equal(true, man.CanEdit(), "It should be able to edit")

	// Exclude the corrupted files
	cfg.PreviewOverSize = 0
	all_created_err := CreateAllPreviews(man)
	as.NoError(all_created_err, "Failed to create all previews")
}

func (as *ActionSuite) Test_PreviewsWithCorrupted() {
	// Create all previews including the corrupted, and then do a search
	// for the two corrupted files in the DB
	err := models.DB.TruncateAll()
	as.NoError(err, "Couldn't clean the DB")

	dir, _ := envy.MustGet("DIR")
	as.NotEmpty(dir, "The test must specify a directory to run on")

	// Note for this test we DO allow corrupted files
	cfg := utils.GetCfg()
	cfg.UseDatabase = true
	cfg.Dir = dir

	// Match only our corrupted files
	cfg.IncContent = utils.CreateContentMatcher(".*corrupted.*", "", cfg.IncludeOperator)
	cfg.ExcContent = utils.ExcludeNoFiles

	c_err := CreateInitialStructure(cfg)
	man := GetManagerActionSuite(cfg, as)
	as.NoError(c_err, "Failed to build out the initial database")
	as.Equal(true, man.CanEdit(), "It should be able to edit")

	content, count, m_err := man.ListContent(ContentQuery{PerPage: 100})
	as.NoError(m_err)
	as.Equal(2, len(*content), "It should all be loaded in the db")
	as.Equal(2, count, "Count should be correct")
	for _, mc := range *content {
		if mc.Src != "nature-corrupted-free-use.jpg" && mc.Src != "snow-corrupted-free-use.png" {
			as.Equal(mc.Corrupt, false, fmt.Sprintf("And at this point nothing is corrupt %s", mc.Src))
		}
	}

	// Exclude the corrupted files
	cfg.PreviewOverSize = 0
	all_created_err := CreateAllPreviews(man)
	as.Error(all_created_err, "It should ACTUALLY have an error now.")

	content_check, count, m_err := man.ListContent(ContentQuery{PerPage: 100})
	as.NoError(m_err)
	as.Equal(2, len(*content_check), "It should have two corrupt content items")
	as.Equal(2, count, "Count should be correct")

	for _, mc_check := range *content_check {
		as.Equal(mc_check.Corrupt, true, "These images should actually be corrupt")
	}
}

func (as *ActionSuite) Test_FindDuplicateVideos() {
	err := models.DB.TruncateAll()
	as.NoError(err, "Couldn't clean the DB")

	dir, _ := envy.MustGet("DIR")
	as.NotEmpty(dir, "The test must specify a directory to run on")

	cfg := utils.GetCfg()
	cfg.UseDatabase = false
	cfg.Dir = dir

	c_err := CreateInitialStructure(cfg)
	as.NoError(c_err)
	man := GetManagerActionSuite(cfg, as)

	cquery := ContentQuery{ContentType: "video"}
	contents, total, cErr := man.ListContent(cquery)
	as.NoError(cErr)
	as.Equal(test_common.TOTAL_VIDEO, total, "There should be this many videos")
	as.Equal(len(*contents), total, "It should find all video content")

	dupeSample := "SampleVideo_1280x720_1mb.mp4"
	// There should be one video already encoded, and we should detect it.
	dupeContent, dupeErr := FindDuplicateVideos(man)
	as.NoError(dupeErr)
	as.NotNil(dupeContent)
	as.Equal(1, len(dupeContent), fmt.Sprintf("It should find %s", dupeSample))
}

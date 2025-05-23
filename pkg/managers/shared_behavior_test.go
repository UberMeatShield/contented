package managers

import (
	"contented/pkg/config"
	"contented/pkg/models"
	"contented/pkg/test_common"
	"contented/pkg/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TOTAL_IN_SCREENS = 4

// TODO: This naming is now bad with preview screens
func GetScreens() (*models.Container, models.Contents) {
	return test_common.GetContentByDirName("screens")
}

func SetupScreensPreview(t *testing.T) (*models.Container, models.Contents) {
	test_common.ResetConfig()
	models.ResetDB(models.InitGorm(false))

	c_pt, content := GetScreens()
	clear_err := utils.ClearContainerPreviews(c_pt)
	assert.NoError(t, clear_err, "And we should clear the preview dir", t)

	dstPath := utils.GetContainerPreviewDst(c_pt)
	dir_err := utils.MakePreviewPath(dstPath)
	assert.NoError(t, dir_err, "Did we create preview path", t)

	empty, read_err := os.ReadDir(dstPath)
	assert.Empty(t, empty, fmt.Sprintf("It has an empty directory %s", dstPath))
	assert.NoError(t, read_err, "It should be able to read the dst directory")
	return c_pt, content
}

func TestSharedInitialCreation(t *testing.T) {
	cfg := test_common.ResetConfig()
	db := models.ResetDB(models.InitGorm(false))

	dir := config.MustGetEnvString("DIR")
	assert.NotEmpty(t, dir, "The test must specify a directory to run on")

	assert.True(t, cfg.ExcContent(".DS_Store", "application/octet-stream"), "This should not be allowed")

	err := CreateInitialStructure(cfg)
	assert.NoError(t, err, "It should successfully create the full DB setup")

	cnts := models.Containers{}
	db.Find(&cnts)

	assert.Equal(t, test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(cnts), "The mocks have a specific expected number of items")

	content := models.Contents{}
	db.Find(&content)
	assert.Equal(t, test_common.TOTAL_MEDIA, len(content), "The mocks have a specific expected number of items")
}

// TODO: Is this even db ?
func TestSharedCfgIncExcContent(t *testing.T) {
	cfg := test_common.ResetConfig()
	db := models.ResetDB(models.InitGorm(false))

	// Exclude all images
	dir := config.MustGetEnvString("DIR")
	cfg.Dir = dir
	nope := config.CreateContentMatcher("DS_Store", "image|text", "OR")

	assert.True(t, nope(".DS_Store", "image/png"))
	cfg.ExcContent = nope
	config.SetCfg(*cfg)
	err := CreateInitialStructure(cfg)
	assert.NoError(t, err)

	content := models.Contents{}
	db.Find(&content)
	dbg, _ := json.Marshal(content)
	assert.Equal(t, test_common.TOTAL_VIDEO, len(content), fmt.Sprintf("There should be three: %s", dbg))
	assert.Equal(t, content[0].ContentType, "video/mp4", "It should be the video")

	models.ResetDB(models.InitGorm(false))
	cfg.ExcContent = config.ExcludeNoFiles
	cfg.IncContent = config.CreateContentMatcher("", "jpeg", "AND")

	err_png := CreateInitialStructure(cfg)
	assert.NoError(t, err_png)

	jpegContent := models.Contents{}
	db.Find(&jpegContent)
	assert.Equal(t, 2, len(jpegContent), fmt.Sprintf("There are 2 jpeg %s", jpegContent))
}

func TestSharedImgShouldCreatePreview(t *testing.T) {
	cfg := test_common.ResetConfig()
	dir := config.MustGetEnvString("DIR")

	cfg.Dir = dir

	id := utils.AssignNumerical(0, "containers")
	cnt := models.Container{
		Path:   dir,
		Name:   "screens",
		ID:     id,
		Active: true,
	}
	content := utils.FindContent(cnt, 4, 0)
	assert.Equal(t, len(content), 4)

	// Basic sanity check that things exist and can preview
	for _, m := range content {
		fq_path, err := utils.GetFilePathInContainer(m.Src, cnt.GetFqPath())
		assert.NoError(t, err, "It should not fail getting a full file path"+m.Src)

		f, open_err := os.Open(fq_path)
		assert.NoError(t, open_err, "And all the paths should exist")
		assert.Equal(t, utils.ShouldCreatePreview(f, 0), true, "Create all previews with no min size")
	}
}

// TODO: This should eventually move to utils/previews_test.go
func TestSharedCreatePreview(t *testing.T) {
	// Create one that is a fail
	cfg := test_common.ResetConfig()
	c_pt, content := GetScreens()
	err := utils.ClearContainerPreviews(c_pt)
	assert.NoError(t, err, "It should nuke out the preview directory")
	assert.Equal(t, TOTAL_IN_SCREENS, len(content), "There should be 4 of these in the screens dir")

	dstPath := utils.GetContainerPreviewDst(c_pt)
	dir_err := utils.MakePreviewPath(dstPath)
	assert.NoError(t, dir_err, "Did we createa preview path")

	// Create one that does create a preview
	cfg.PreviewOverSize = 0
	for _, mc := range content {
		preview_path, err := utils.CreateContentPreview(c_pt, &mc)
		assert.NoError(t, err, "It should be ble to create previews")
		assert.NotEqual(t, preview_path, "", "The path should be defined")
	}
	previews, read_err := os.ReadDir(dstPath)
	assert.Equal(t, TOTAL_IN_SCREENS, len(previews), "It should create 4 previews")
	assert.NoError(t, read_err, "It should be able to read the directory")
}

func TestSharedCreateBaseTags(t *testing.T) {
	models.ResetDB(models.InitGorm(false))
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = true
	cfg.TagFile = filepath.Join(cfg.Dir, "/dir2/tags.txt")
	config.SetCfg(*cfg)

	// TODO: Fix / nuke this
	man := GetManagerTestSuite(cfg)
	assert.Equal(t, man.CanEdit(), true)

	tags, createTagsErr := CreateTagsFromFile(man)
	assert.NoError(t, createTagsErr)
	assert.NotNil(t, tags)
	assert.Equal(t, test_common.TOTAL_TAGS, len(*tags), "It should keep track of what was created")

	tagsCheck, total, lErr := man.ListAllTags(TagQuery{PerPage: 50})
	assert.NoError(t, lErr, "It should not error out listing tags")
	assert.Greater(t, total, int64(0), "We should have a positive tag count")
	assert.NotNil(t, tags, "There should be tags")
	assert.Equal(t, test_common.TOTAL_TAGS, len(*tagsCheck), "The tags should be created and queried")

	tagsAgain, retryError := CreateTagsFromFile(man)
	assert.NoError(t, retryError, "We should not have any DB error")
	assert.Equal(t, test_common.TOTAL_TAGS, len(*tagsAgain), "The Tag count should be the same.")
	// Add test to ensure we do not try and create the same tag over and over
	// TODO: Need to do a test that actually checks memory
}

func TestSharedCreateContainerPreviews(t *testing.T) {
	// Get a local not in DB setup for the container and content
	// Create a bunch of previews
	db := models.ResetDB(models.InitGorm(false))

	// Now add the data into the database
	c_pt, content := SetupScreensPreview(t)
	cfg := config.GetCfg()
	cfg.UseDatabase = true
	cfg.PreviewOverSize = int64(1)
	config.SetCfg(*cfg) // Note the SetupScrensPreview resets the config

	cRes := db.Create(c_pt)
	test_common.NoError(cRes, "Failed to create content", t)
	assert.Greater(t, len(content), 0, "We should have content")

	// Check that we have a container preview at this point
	expect_c_preview := ""
	for idx, mc := range content {
		mc.Idx = idx
		mcRes := db.Create(&mc)
		assert.NoError(t, mcRes.Error)
		assert.Equal(t, mc.Preview, "", "There should be no preview at this point")
		if expect_c_preview == "" {
			expect_c_preview = fmt.Sprintf("/api/preview/%d", mc.ID)
		}
	}

	man := GetManagerTestSuite(cfg)
	cnts, count, c_err := man.ListContainers(ContainerQuery{Page: 1, PerPage: 2})
	assert.Greater(t, count, int64(0), "There should be a container")
	assert.Equal(t, len(*cnts), 1, "It should have containers")
	assert.NoError(t, c_err)

	p_err := CreateContainerPreviews(c_pt, man)
	assert.Equal(t, expect_c_preview, c_pt.PreviewUrl, "It should assign a mc preview to the container")

	assert.NoError(t, p_err, "An error happened creating the previews")
	dstPath := utils.GetContainerPreviewDst(c_pt)
	previews, read_err := os.ReadDir(dstPath)
	assert.Equal(t, TOTAL_IN_SCREENS, len(previews), fmt.Sprintf("It should create 6 previews %s", dstPath))
	assert.NoError(t, read_err, "It should be able to read the directory")

	// Validate that the content was updated in the DB
	content_check := models.Contents{}
	db.Where("container_id = ?", c_pt.ID).Order("created_at desc").Find(&content_check)
	assert.Equal(t, TOTAL_IN_SCREENS, len(content_check), "We should just have 6 things to check")
	for _, mc_check := range content_check {
		assert.NotEqual(t, mc_check.Preview, "", fmt.Sprintf("It should now have a preview %s", mc_check.Src))
	}
}

func TestSharedAsyncContainerPreviews(t *testing.T) {
	cfg := test_common.ResetConfig()
	c_pt, content := SetupScreensPreview(t)

	// On the DB side these would then need to be updated in the DB for linkage
	cfg.PreviewOverSize = 0
	previews, err := CreateContentPreviews(c_pt, content)
	assert.NoError(t, err, "It should be able to create all previews successfully")

	assert.Equal(t, len(previews), len(content), "With size zero we should have 4 previews")
	for _, p_mc := range previews {
		assert.NotEqual(t, p_mc.Preview, "", "All results should have a preview")
	}
	// TODO: Validate the previews are created on disk
	// Map the results back to the content containers
	// Maybe just return them vs update the DB
}

func TestSharedPreviewAllData(t *testing.T) {
	cfg := test_common.ResetConfig()
	db := models.ResetDB(models.InitGorm(false))
	assert.NoError(t, db.Error, "Couldn't clean the DB")

	dir := config.MustGetEnvString("DIR")
	assert.NotEmpty(t, dir, "The test must specify a directory to run on")

	cfg.UseDatabase = true
	cfg.Dir = dir
	cfg.ExcContent = config.CreateContentMatcher("corrupted", "", cfg.ExcludeOperator)

	cErr := CreateInitialStructure(cfg)
	assert.NoError(t, cErr, "No errors should occur during data creation")
	man := GetManagerTestSuite(cfg)

	cquery := ContainerQuery{Page: 1, PerPage: 3}
	cnts, count, c_err := man.ListContainers(cquery)
	assert.Greater(t, count, int64(1), "There should be a positive count")
	assert.Equal(t, cquery.PerPage, len(*cnts), "It should have containers")
	assert.NoError(t, c_err)

	assert.NoError(t, c_err, "Failed to build out the initial database")
	assert.Equal(t, true, man.CanEdit(), "It should be able to edit")

	// Exclude the corrupted files
	cfg.PreviewOverSize = 0
	all_created_err := CreateAllPreviews(man)
	assert.NoError(t, all_created_err, "Failed to create all previews")
}

func TestSharedPreviewsWithCorrupted(t *testing.T) {
	// Create all previews including the corrupted, and then do a search
	// for the two corrupted files in the DB
	cfg := test_common.ResetConfig()
	db := models.ResetDB(models.InitGorm(false))
	assert.NoError(t, db.Error, "Couldn't clean the DB")

	dir := config.MustGetEnvString("DIR")
	assert.NotEmpty(t, dir, "The test must specify a directory to run on")

	// Note for this test we DO allow corrupted files
	cfg.UseDatabase = true
	cfg.Dir = dir

	// Match only our corrupted files
	cfg.IncContent = config.CreateContentMatcher(".*corrupted.*", "", cfg.IncludeOperator)
	cfg.ExcContent = config.ExcludeNoFiles

	c_err := CreateInitialStructure(cfg)
	man := GetManagerTestSuite(cfg)
	assert.NoError(t, c_err, "Failed to build out the initial database")
	assert.Equal(t, true, man.CanEdit(), "It should be able to edit")

	content, count, m_err := man.ListContent(ContentQuery{PerPage: 100})
	assert.NoError(t, m_err)
	assert.Equal(t, 2, len(*content), "It should all be loaded in the db")
	assert.Equal(t, int64(2), count, "Count should be correct")
	for _, mc := range *content {
		if mc.Src != "nature-corrupted-free-use.jpg" && mc.Src != "snow-corrupted-free-use.png" {
			assert.Equal(t, mc.Corrupt, false, fmt.Sprintf("And at this point nothing is corrupt %s", mc.Src))
		}
	}

	// Exclude the corrupted files
	cfg.PreviewOverSize = 0
	all_created_err := CreateAllPreviews(man)
	assert.Error(t, all_created_err, "It should ACTUALLY have an error now.")

	content_check, count, m_err := man.ListContent(ContentQuery{PerPage: 100})
	assert.NoError(t, m_err)
	assert.Equal(t, 2, len(*content_check), "It should have two corrupt content items")
	assert.Equal(t, int64(2), count, "Count should be correct")

	for _, mc_check := range *content_check {
		assert.Equal(t, mc_check.Corrupt, true, "These images should actually be corrupt")
	}
}

func TestSharedFindDuplicateVideos(t *testing.T) {
	cfg := test_common.ResetConfig()
	db := models.ResetDB(models.InitGorm(false))
	assert.NoError(t, db.Error, "Couldn't clean the DB")

	dir := config.MustGetEnvString("DIR")
	assert.NotEmpty(t, dir, "The test must specify a directory to run on")

	cfg.UseDatabase = false
	cfg.Dir = dir

	c_err := CreateInitialStructure(cfg)
	assert.NoError(t, c_err, "Failed to create intial structure")
	man := GetManagerTestSuite(cfg)

	cquery := ContentQuery{ContentType: "video"}
	contents, total, cErr := man.ListContent(cquery)
	assert.NoError(t, cErr, "Could not list contents")
	assert.Equal(t, int64(test_common.TOTAL_VIDEO), total, "There should be this many videos")
	assert.Equal(t, int64(len(*contents)), total, "It should find all video content")
	assert.Greater(t, total, int64(0), "Ensure there are matches to detect duplicate video")

	dupeSample := "SampleVideo_1280x720_1mb.mp4"
	// There should be one video already encoded, and we should detect it.
	dupeContent, dupeErr := FindDuplicateVideos(man)
	assert.NoError(t, dupeErr, "Duplicate errors found")
	assert.NotNil(t, dupeContent)
	assert.Equal(t, 1, len(dupeContent), fmt.Sprintf("It should find %s", dupeSample))
}

func TestClearScreensMemory(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	man := GetManagerTestSuite(cfg)
	ValidateClearScreens(t, man)
}
func TestClearScreensDatabase(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	ValidateClearScreens(t, man)
}

func ValidateClearScreens(t *testing.T, man ContentManager) {
	content := models.Content{
		Src:    "ScreenContent",
		NoFile: true,
	}
	assert.NoError(t, man.CreateContent(&content), "Failed to create the content")
	assert.Greater(t, content.ID, int64(0), "It should create a valid content")

	contentID := content.ID
	screen1 := models.Screen{ContentID: contentID, Src: "https://some.url.com"}
	screen2 := models.Screen{ContentID: contentID, Src: "https://other.url.com"}

	assert.NoError(t, man.CreateScreen(&screen1), "Failed to create screen1")
	assert.NoError(t, man.CreateScreen(&screen2), "Failed to create screen2")

	sq := ScreensQuery{
		ContentID: strconv.FormatInt(contentID, 10),
	}
	screens, total, err := man.ListScreens(sq)
	assert.NoError(t, err, "Failed to list screens")
	assert.Equal(t, int64(2), total, "We should have our screens")
	assert.NotNil(t, screens)

	assert.NoError(t, man.ClearScreens(&content), "It should clear the screens")

	_, totalAfter, errAfter := man.ListScreens(sq)
	assert.Equal(t, int64(0), totalAfter, "Now it should be clear")
	assert.NoError(t, errAfter, "It should not fail the list")
}

func TestClearScreensDiskMemory(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	man := GetManagerTestSuite(cfg)
	ValidateClearScreensOnDisk(t, man)
}
func TestClearScreensDiskDatabase(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	ValidateClearScreensOnDisk(t, man)
}

func ValidateClearScreensOnDisk(t *testing.T, man ContentManager) {
	containerPreviews, testDir := test_common.CreateTestPreviewsContainerDirectory(t)
	utils.ResetPreviewDir(containerPreviews)
	container := models.Container{
		Name: "dir2",
		Path: testDir,
	}
	assert.NoError(t, man.CreateContainer(&container), "Failed to create the container %s", container.GetFqPath())

	content := models.Content{
		Src:         "donut_[special( gunk.mp4",
		ContainerID: &container.ID,
	}
	assert.NoError(t, man.CreateContent(&content), "Failed to create the content")
	assert.Greater(t, content.ID, int64(0), "It should create a valid content")

	// Create some content on disk so we can validate disk removal of these elements
	screenName1, err := test_common.WriteScreenFile(containerPreviews, "ScreenTestRemove", 1)
	assert.NoError(t, err, "Failed to write screen file: %v", err)
	screenName2, err := test_common.WriteScreenFile(containerPreviews, "ScreenTestRemove", 2)
	assert.NoError(t, err, "Failed to write screen file: %v", err)

	screen1 := models.Screen{ContentID: content.ID, Src: screenName1, Path: containerPreviews}
	screen2 := models.Screen{ContentID: content.ID, Src: screenName2, Path: containerPreviews}
	assert.NoError(t, man.CreateScreen(&screen1), "Failed to create screen1")
	assert.NoError(t, man.CreateScreen(&screen2), "Failed to create screen2")

	// Verify directory now has two files
	files, err := os.ReadDir(containerPreviews)
	assert.NoError(t, err, "Should be able to read directory")
	assert.NotEmpty(t, files, "Directory should be empty before test")
	assert.Equal(t, 2, len(files), fmt.Sprintf("Directory should have two files %s", containerPreviews))

	// Verify the screens exist in the managers
	sq := ScreensQuery{
		ContentID: strconv.FormatInt(content.ID, 10),
	}
	screens, total, err := man.ListScreens(sq)
	assert.NoError(t, err, "Failed to list screens")
	assert.Equal(t, int64(2), total, "We should have our screens")
	assert.NotNil(t, screens)

	err = RemoveScreensForContent(man, content.ID)
	assert.NoError(t, err, "Could not remove screens for content")

	// Check that the screens are gone
	_, totalAfter, errAfter := man.ListScreens(sq)
	assert.Equal(t, int64(0), totalAfter, "Now it should be clear")
	assert.NoError(t, errAfter, "It should not fail the list")

	// Check that the files are gone
	files, err = os.ReadDir(containerPreviews)
	assert.NoError(t, err, "Should be able to read directory")
	assert.Empty(t, files, "No files should be left on disk")
}

func Test_RemoveContentFromDiskDB(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(true)
	test_common.CreateContentByDirName("dir3")

	man := GetManagerTestSuite(cfg)
	ValidateRemoveContentFromDisk(t, man)
}

func Test_RemoveContentFromDiskMemory(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	man := GetManagerTestSuite(cfg)
	ValidateRemoveContentFromDisk(t, man)
}

func ValidateRemoveContentFromDisk(t *testing.T, man ContentManager) {
	cfg := man.GetCfg()
	removeLocation, mkErr := test_common.SetupRemovalLocation(cfg)
	assert.NoError(t, mkErr, fmt.Sprintf("Failed to create remove location %s", mkErr))

	cs := ContainerQuery{
		Name:    "dir3",
		PerPage: 100,
	}
	cnts, _, err := man.SearchContainers(cs)
	assert.NoError(t, err, fmt.Sprintf("Failed to search containers %s", err))
	assert.NotNil(t, cnts)
	assert.Equal(t, 1, len(*cnts), fmt.Sprintf("We should have 1 screens container %s", cnts))
	cnt := (*cnts)[0]

	content, err := test_common.CreateTestContentInContainer(&cnt, "test.txt", "test removal")
	assert.NoError(t, err, "Failed to create test file")

	_, verifyFileErr := os.Stat(filepath.Join(cnt.GetFqPath(), content.Src))
	assert.NoError(t, verifyFileErr, "Content should exist on disk")

	// Test removing content
	dst, err := RemoveContentFromContainer(man, content, nil)
	assert.NoError(t, err, fmt.Sprintf("Failed to remove content from disk %s", err))
	assert.NotEmpty(t, dst, fmt.Sprintf("Failed to remove content from disk no dst %s", dst))

	// Verify original file is gone
	_, err = os.Stat(filepath.Join(cnt.GetFqPath(), content.Src))
	assert.True(t, os.IsNotExist(err), fmt.Sprintf("Original file should be removed %s content %s", err, content.Src))

	// Verify file exists in new location
	_, err = os.Stat(dst)
	assert.NoError(t, err, fmt.Sprintf("File should exist in remove location %s", dst))

	// Clean up
	defer os.RemoveAll(removeLocation)
}

func Test_ContainerDuplicateRemovalMemory(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(false)
	man := GetManagerTestSuite(cfg)
	ValidateContainerDuplicateRemoval(t, man)
}

func Test_ContainerDuplicateRemovalDB(t *testing.T) {
	cfg, _ := test_common.InitFakeApp(true)
	man := GetManagerTestSuite(cfg)
	ValidateContainerDuplicateRemoval(t, man)
}

func ValidateContainerDuplicateRemoval(t *testing.T, man ContentManager) {
	cfg := man.GetCfg()

	removeLocation, mkErr := test_common.SetupRemovalLocation(cfg)
	assert.NoError(t, mkErr, fmt.Sprintf("Failed to create remove location %s", mkErr))

	total := 4
	cnt, contents, err := test_common.CreateTestRemovalContent(cfg, total)
	cnt.ID = 0
	assert.NoError(t, err, "Failed to create test removal content")
	assert.NotNil(t, cnt)
	assert.NotNil(t, contents)
	assert.Equal(t, total, len(contents))

	cntErr := man.CreateContainer(cnt)
	assert.NoError(t, cntErr, fmt.Sprintf("Failed to create container %s", cntErr))

	// Set all the contents except the first one as duplicates for removal purposes
	for idx, content := range contents {
		content.ContainerID = &cnt.ID
		if idx != 0 {
			content.Duplicate = true
		}
		contentErr := man.CreateContent(&content)
		assert.NoError(t, contentErr, fmt.Sprintf("Failed to create content %s", contentErr))
	}

	removed_count, err := RemoveDuplicateContents(man, cnt)
	assert.NoError(t, err, fmt.Sprintf("Failed to remove duplicate contents %s", err))
	assert.Equal(t, total-1, removed_count, "We should have removed 2 contents")

	// Verify the contents are gone
	for idx, content := range contents {
		if idx == 0 {
			continue
		}
		_, err = os.Stat(filepath.Join(cnt.GetFqPath(), content.Src))
		assert.True(t, os.IsNotExist(err), fmt.Sprintf("Content should be removed %s", content.Src))
	}

	defer os.RemoveAll(removeLocation)
	defer test_common.RemoveTestContent()
}

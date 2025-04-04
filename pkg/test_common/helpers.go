package test_common

/**
 * These are some common test helper files dealing mostly with unit test setup and
 * mock data counts and information.
 */
import (
	"contented/pkg/config"
	"contented/pkg/models"
	"contented/pkg/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const TOTAL_CONTAINERS = 9
const TOTAL_CONTAINERS_WITH_CONTENT = 7
const TOTAL_MEDIA = 35
const TOTAL_TAGS = 49
const TOTAL_VIDEO = 3
const VIDEO_FILENAME = "donut_[special( gunk.mp4"

const TEST_REMOVAL_LOCATION = "test_removal_content"

var EXPECT_CNT_COUNT = map[string]int{
	"dir1":            12,
	"dir2":            4,
	"dir3":            10,
	"screens":         4,
	"screens_sub_dir": 2,
	"test_encoding":   2,
	"not_empty":       1,
}

// Helper for a common block of video test code (duplicated in internals)
func Get_VideoAndSetupPaths(cfg *config.DirConfigEntry) (string, string, string) {
	// The video we use is only 10.08 seconds long.
	cfg.PreviewFirstScreenOffset = 2
	cfg.PreviewNumberOfScreens = 4
	config.SetCfg(*cfg)

	dir := config.MustGetEnvString("DIR")
	srcDir := filepath.Join(dir, "dir2")
	dstDir := utils.GetPreviewDst(srcDir)
	testFile := "donut_[special( gunk.mp4"

	// Ensure that the preview destination directory is clean
	utils.ResetPreviewDir(dstDir)
	return srcDir, dstDir, testFile
}

func CreateTestPreviewsContainerDirectory(t *testing.T) (string, string) {
	cfg := config.GetCfg()
	testDir := cfg.Dir
	_, containerPreviews, _ := Get_VideoAndSetupPaths(cfg)

	// Check we can write to the video destination directory (probably not needed)
	ok, err := utils.PathIsOk(containerPreviews, "", testDir)
	if err != nil {
		t.Errorf("Failed to check path %s", err)
	}
	if !ok {
		t.Errorf("Path was not ok %s", containerPreviews)
	}

	if _, err := os.Stat(containerPreviews); os.IsNotExist(err) {
		err := os.MkdirAll(containerPreviews, 0755)
		if err != nil {
			t.Fatalf("Failed to create container_previews directory: %v", err)
		}
	}
	ok, err = utils.PathIsOk(containerPreviews, "", testDir)
	if err != nil {
		t.Errorf("Failed to check container path %s", err)
	}
	if !ok {
		t.Errorf("container Path was not ok %s", containerPreviews)
	}
	return containerPreviews, testDir
}

func CreateTestContentInContainer(cnt *models.Container, fileName string, writeString string) (*models.Content, error) {
	testFile := "test.txt"
	contentPath := filepath.Join(cnt.GetFqPath(), testFile)

	err := os.WriteFile(contentPath, []byte(writeString), 0644)
	if err != nil {
		return nil, err
	}
	content := models.Content{
		ID:          42,
		ContainerID: &cnt.ID,
		Src:         testFile,
	}
	return &content, nil
}

// Remember to cleanup after the test
func SetupRemovalLocation(cfg *config.DirConfigEntry) (string, error) {
	if cfg.Dir == "" {
		cfg.Dir = config.MustGetEnvString("DIR")
	}

	removeLocation := filepath.Join(cfg.Dir, TEST_REMOVAL_LOCATION)
	cfg.RemoveLocation = removeLocation
	mkErr := os.MkdirAll(removeLocation, 0755)
	fmt.Printf("Remove location: %s\n", removeLocation)
	config.SetCfg(*cfg)
	return removeLocation, mkErr
}

// Generates files on disk in a temporary directory and returns the container and contents
func CreateTestRemovalContent(cfg *config.DirConfigEntry, amountToCreate int) (cnt *models.Container, contents models.Contents, err error) {
	if cfg.Dir == "" {
		cfg.Dir = config.MustGetEnvString("DIR")
	}
	locationPath := filepath.Join(cfg.Dir, TEST_REMOVAL_LOCATION)
	log.Printf("Creating %d test removal %s content in %s", amountToCreate, cfg.Dir, locationPath)

	mkErr := os.MkdirAll(locationPath, 0755)
	if mkErr != nil {
		return nil, nil, mkErr
	}

	for i := 0; i < amountToCreate; i++ {
		fileName := fmt.Sprintf("test_%d.txt", i)
		filePath := filepath.Join(locationPath, fileName)
		err := os.WriteFile(filePath, []byte(fmt.Sprintf("test %d", i)), 0644)
		if err != nil {
			return nil, nil, err
		}
	}
	return CreateContentByDirName(TEST_REMOVAL_LOCATION)
}

func RemoveTestContent() error {
	cfg := config.GetCfg()
	location := "test_removal_content"
	locationPath := filepath.Join(cfg.Dir, location)
	return os.RemoveAll(locationPath)
}

// duplicated in utils/previews_test.go
func WriteScreenFile(dstPath string, fileName string, count int) (string, error) {
	screenName := fmt.Sprintf("%s.screens.00%d.jpg", fileName, count)
	fqPath := filepath.Join(dstPath, screenName)
	f, err := os.Create(fqPath)
	if err != nil {
		return "", err
	}
	_, wErr := f.WriteString("Now something exists in the file")
	if wErr != nil {
		return "", wErr
	}
	return screenName, nil
}

func GetContext() *gin.Context {
	return GetContextParams("/containers", "1", "10")
}

func GetContextParams(url string, page string, per_page string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Params = []gin.Param{
		{Key: "page", Value: page},
		{Key: "per_page", Value: per_page},
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	c.Request = req
	return c
}

// TODO validate octet/stream
func IsValidContentType(content_type string) error {
	valid := map[string]bool{
		"image/png":                 true,
		"image/jpeg":                true,
		"image/gif":                 true,
		"image/webp":                true,
		"application/octet-stream":  true,
		"video/mp4":                 true,
		"text/plain; charset=utf-8": true,
	}
	if valid[content_type] {
		return nil
	}
	return errors.New("Invalid content type: " + content_type)
}

func ResetConfig() *config.DirConfigEntry {
	cfg := config.GetCfgDefaults()
	dir := config.MustGetEnvString("DIR")
	cfg.Dir = dir
	cfg.MaxSearchDepth = 3
	config.InitConfig(dir, &cfg)
	config.SetupContainerMatchers(&cfg, "", "DS_Store|container_previews")
	config.SetupContentMatchers(&cfg, "", "image|video|text", "DS_Store", "")
	config.SetCfg(cfg)
	return config.GetCfg()
}

// This function is now how the init method should function till caching is implemented
// As the internals / guts are functional using the new models the creation of models
// can be removed.
func InitFakeApp(use_db bool) (*config.DirConfigEntry, *gorm.DB) {
	dir := config.MustGetEnvString("DIR")
	fmt.Printf("Using directory %s\n", dir)

	cfg := ResetConfig()
	cfg.UseDatabase = use_db // Set via .env or USE_DATABASE as an environment var
	cfg.StaticResourcePath = "./public/build"
	cfg.StartQueueWorkers = false
	config.SetCfg(*cfg)

	// TODO: Assign the context into the manager (force it?)
	if !cfg.UseDatabase {

		log.Printf("Using memory storage for unit test")
		// TODO: This moves into managers.. is there a sane way of handling this?
		memStorage := utils.InitializeMemory(dir)

		// cnts := memStorage.ValidContainers
		// for _, c := range cnts {
		mcs := memStorage.ValidContent
		for _, mc := range mcs {
			if mc.Src == "this_is_p_ng" {
				mc.Preview = "preview_this_is_p_ng"
			}
		}

		// Fake somehidden content (put int on disk maybe?)
		hiddenContainer := models.Container{
			ID:     utils.AssignNumerical(0, "containers"),
			Name:   "hide",
			Hidden: true,
		}

		hiddenContent := models.Content{
			ID:          utils.AssignNumerical(0, "contents"),
			ContainerID: &hiddenContainer.ID,
			Hidden:      true,
			Src:         "hidden.txt",
		}
		memStorage.ValidContainers[hiddenContainer.ID] = hiddenContainer
		memStorage.ValidContent[hiddenContent.ID] = hiddenContent
	} else {
		db := models.ResetDB(models.InitGorm(false))
		return cfg, db
	}
	return cfg, nil
}

func NilError(err error, msg string, t *testing.T) {
	if err != nil {
		t.Errorf("%s error: %s", msg, err)
	}
}

func NoError(tx *gorm.DB, msg string, t *testing.T) {
	if tx.Error != nil {
		t.Errorf("%s error: %s", msg, tx.Error)
	}
}

// For loading up the memory app but not searching directories checking content etc.
func InitMemoryFakeAppEmpty() *config.DirConfigEntry {
	dir := config.MustGetEnvString("DIR")
	fmt.Printf("Using directory %s\n", dir)
	cfg := ResetConfig()
	cfg.UseDatabase = false
	cfg.StaticResourcePath = "./public/build"
	utils.InitializeEmptyMemory()
	return cfg
}

// Note this is DB only, init fake app creates the content for memory by default
func CreateContentByDirName(test_dir_name string) (*models.Container, models.Contents, error) {
	cnt, content := GetContentByDirName(test_dir_name)

	db := models.InitGorm(false)
	res := db.Create(cnt)
	if res.Error != nil {
		return nil, nil, res.Error
	}

	for _, mc := range content {
		mc.ContainerID = &cnt.ID
		mRes := db.Create(&mc)
		if mRes.Error != nil {
			return nil, nil, mRes.Error
		}
	}
	return cnt, content, nil
}

func GetContentByDirName(test_dir_name string) (*models.Container, models.Contents) {
	dir := config.MustGetEnvString("DIR")
	cfg := config.GetCfg()
	cfg.Dir = dir
	cnts := utils.FindContainers(cfg.Dir)

	var cnt *models.Container = nil
	for _, c := range cnts {
		if c.Name == test_dir_name {
			cnt = &c
			break
		}
	}
	if cnt == nil {
		log.Panic("Could not find the directory: " + test_dir_name)
	}
	content := utils.FindContentMatcher(*cnt, 42, 0, cfg.IncContent, cfg.ExcContent)
	cnt.Total = len(content)
	return cnt, content
}

// Cleanup paths that will be created by CreateContainerPath, note we IGNORE
// the c.Path because it will be reset and assigned in actions tests.
func CleanupContainer(c *models.Container) error {
	cfg := config.GetCfg()
	fqPath := filepath.Join(cfg.Dir, c.Name)
	fmt.Printf("CleanupContent() Test trying to cleanup %s", fqPath)
	if f, err := os.Stat(fqPath); !os.IsNotExist(err) {
		if f.IsDir() {
			err := os.Remove(fqPath)
			if err != nil {
				fmt.Printf("CleanupContent() Failed to cleanup test %s", err)
			}
			return err
		}
	}
	return nil
}

// Note that is ONLY for test purposes but could be used if we allow containers to
// 'create' their own directory (for uploads eventually.)
func CreateContainerPath(c *models.Container) (string, error) {
	cfg := config.GetCfg()
	fqPath := ""
	if cfg.Dir != "" && cfg.Dir != "~" {
		c.Path = cfg.Dir
		fqPath = c.GetFqPath() // Currently just ignore any path specified in the Container
		// fmt.Printf("It should be trying to create %s\n", fqPath)
		if _, err := os.Stat(fqPath); os.IsNotExist(err) {
			return fqPath, os.Mkdir(fqPath, 0655)
		}
	}
	return fqPath, nil
}

package test_common

/**
 * These are some common test helper files dealing mostly with unit test setup and
 * mock data counts and information.
 */
import (
	"contented/models"
	"contented/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)

const TOTAL_CONTAINERS = 9
const TOTAL_CONTAINERS_WITH_CONTENT = 7
const TOTAL_MEDIA = 35
const TOTAL_TAGS = 49
const TOTAL_VIDEO = 3
const VIDEO_FILENAME = "donut_[special( gunk.mp4"

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
func Get_VideoAndSetupPaths(cfg *utils.DirConfigEntry) (string, string, string) {
	// The video we use is only 10.08 seconds long.
	cfg.PreviewFirstScreenOffset = 2
	cfg.PreviewNumberOfScreens = 4
	utils.SetCfg(*cfg)

	var testDir, _ = envy.MustGet("DIR")
	srcDir := filepath.Join(testDir, "dir2")
	dstDir := utils.GetPreviewDst(srcDir)
	testFile := "donut_[special( gunk.mp4"

	// Ensure that the preview destination directory is clean
	utils.ResetPreviewDir(dstDir)
	return srcDir, dstDir, testFile
}

func GetContext(*buffalo.App) *gin.Context {
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

func ResetConfig() *utils.DirConfigEntry {
	cfg := utils.GetCfgDefaults()
	dir, _ := envy.MustGet("DIR")
	cfg.Dir = dir
	cfg.MaxSearchDepth = 3
	utils.InitConfig(dir, &cfg)
	utils.SetupContainerMatchers(&cfg, "", "DS_Store|container_previews")
	utils.SetupContentMatchers(&cfg, "", "image|video|text", "DS_Store", "")
	utils.SetCfg(cfg)
	return utils.GetCfg()
}

// This function is now how the init method should function till caching is implemented
// As the internals / guts are functional using the new models the creation of models
// can be removed.
func InitFakeApp(use_db bool) *utils.DirConfigEntry {
	dir, _ := envy.MustGet("DIR")
	fmt.Printf("Using directory %s\n", dir)

	cfg := ResetConfig()
	cfg.UseDatabase = use_db // Set via .env or USE_DATABASE as an environment var
	cfg.StaticResourcePath = "./public/build"
	cfg.StartQueueWorkers = false

	// TODO: Assign the context into the manager (force it?)
	if !cfg.UseDatabase {

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
			ContainerID: hiddenContainer.ID,
			Hidden:      true,
			Src:         "hidden.txt",
		}
		memStorage.ValidContainers[hiddenContainer.ID] = hiddenContainer
		memStorage.ValidContent[hiddenContent.ID] = hiddenContent
	} else {
		models.DB.TruncateAll()
	}
	return cfg
}

// For loading up the memory app but not searching directories checking content etc.
func InitMemoryFakeAppEmpty() *utils.DirConfigEntry {
	dir, _ := envy.MustGet("DIR")
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

	c_err := models.DB.Create(cnt)
	if c_err != nil {
		return nil, nil, c_err
	}
	for _, mc := range content {
		mc.ContainerID = cnt.ID
		m_err := models.DB.Create(&mc)
		if m_err != nil {
			return nil, nil, m_err
		}
	}
	return cnt, content, nil
}

func GetContentByDirName(test_dir_name string) (*models.Container, models.Contents) {
	dir, _ := envy.MustGet("DIR")
	cfg := utils.GetCfg()
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
	cfg := utils.GetCfg()
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
	cfg := utils.GetCfg()
	fqPath := ""
	if cfg.Dir != "" && cfg.Dir != "~" {
		c.Path = cfg.Dir
		fqPath = c.GetFqPath() // Currently just ignore any path specified in the Container
		// fmt.Printf("It should be trying to create %s\n", fqPath)
		if _, err := os.Stat(fqPath); os.IsNotExist(err) {
			return fqPath, os.Mkdir(fqPath, 0644)
		}
	}
	return fqPath, nil
}

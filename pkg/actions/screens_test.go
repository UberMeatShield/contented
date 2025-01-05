package actions

import (
	"contented/pkg/config"
	"contented/pkg/models"
	"contented/pkg/test_common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func CreatePreview(src string, contentID int64, t *testing.T, router *gin.Engine) models.Screen {

	screen := &models.Screen{
		Src:       src,
		ContentID: contentID,
		Idx:       1,
	}
	resObj := &models.Screen{}
	code, err := PostJson("/api/screens", screen, resObj, router)
	assert.NoError(t, err, "It should create a screen")
	assert.Equal(t, http.StatusCreated, code)
	return *resObj
}

// Kind of a pain in the ass to create all the way down to a valid preview screen
func CreateTestContainerWithContent(t *testing.T, db *gorm.DB) (*models.Container, *models.Content, string) {
	cfg := config.GetCfg()
	srcDir, dstDir, testFile := test_common.Get_VideoAndSetupPaths(cfg)
	c := &models.Container{
		Total: 4,
		Path:  filepath.Dir(srcDir),
		Name:  filepath.Base(srcDir),
	}
	assert.NoError(t, db.Create(c).Error, "Failed to create container")

	// TODO: Ensure that this path is actually correct, should actually make a REAL jpeg copy
	screenName := fmt.Sprintf("%s.screen.001.jpg", testFile)
	screenSrc := filepath.Join(dstDir, fmt.Sprintf("%s.screen.001.jpg", screenName))
	mc := &models.Content{
		Src:         testFile,
		ContentType: "video/mp4",
		Preview:     screenSrc,
		ContainerID: &c.ID,
	}
	assert.NoError(t, db.Create(mc).Error, "Failed to create content")

	fmt.Printf("Screen src %s", screenSrc)
	f, err := os.Create(screenSrc)
	assert.NoError(t, err, fmt.Sprintf("It could not create the screen on disk %s", screenSrc))

	_, wErr := f.WriteString("Totally a screen")
	assert.NoError(t, wErr, fmt.Sprintf("And be able to write to it %s", screenSrc))
	f.Sync()
	f.Close()
	return c, mc, screenSrc
}

func CreateScreen(t *testing.T, db *gorm.DB, router *gin.Engine) (*models.Container, *models.Content, *models.Screen) {
	c, mc, screenSrc := CreateTestContainerWithContent(t, db)
	ps := CreatePreview(screenSrc, mc.ID, t, router)
	return c, mc, &ps
}

func TestScreensResourceList(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	CreateScreen(t, db, router)
	CreateScreen(t, db, router)

	validate := ScreensResponse{}
	code, err := GetJson("/api/screens", "", &validate, router)
	assert.NoError(t, err, "It should be able to list screens")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, len(validate.Results), 2, "There should be two preview screens")
}

func TestScreensResourceListMedia(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)

	// This creates a preview screen making the total 3 in the DB
	// Note it also resets the container_preview dir right now
	CreateScreen(t, db, router)
	_, mc1, _ := CreateScreen(t, db, router)

	CreatePreview("A", mc1.ID, t, router)

	url := fmt.Sprintf("/api/contents/%d/screens", mc1.ID)
	validate := ScreensResponse{}
	code, err := GetJson(url, "", &validate, router)
	assert.NoError(t, err, "It should load up screens")
	assert.Equal(t, http.StatusOK, code, "It should get screens back")

	assert.Equal(t, 2, len(validate.Results), 2, "Note we should have only two screens")
	assert.Equal(t, int64(2), validate.Total, "Count should be correct")

	for _, ps := range validate.Results {
		assert.Equal(t, ps.ContentID, mc1.ID, fmt.Sprintf("Failed %s", ps))
		assert.Equal(t, ps.Path, "", fmt.Sprintf("Path visible %s", ps)) // Path should not be visible in the API
	}
}

func TestScreensResourceShow(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	_, _, ps := CreateScreen(t, db, router)

	url := fmt.Sprintf("/api/screens/%d", ps.ID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Need to make it host the file.
	header := w.Header()
	assert.Equal(t, "image/jpeg", header.Get("Content-Type"))
}

// TODO: Create a screen that is actually on disk.
func TestScreensResourceCreate(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(t, db)
	ps := CreatePreview(screenSrc, mc.ID, t, router)
	assert.Equal(t, ps.Src, filepath.Base(screenSrc))

	screens := models.Screens{}
	db.Where("content_id = ?", mc.ID).Find(&screens)
	assert.Equal(t, len(screens), 1, "There should be a screen in the DB")
}

func TestScreensResourceUpdate(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(t, db)
	ps := CreatePreview(screenSrc, mc.ID, t, router)
	ps.Src = "UP"

	resObj := models.Screen{}
	url := fmt.Sprintf("/api/screens/%d", ps.ID)
	code, err := PutJson(url, ps, &resObj, router)
	assert.NoError(t, err, "It should update")
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("It should update %s", err))
	assert.Equal(t, ps.Src, resObj.Src, "It should have updated")
}

func TestScreensResourceDestroy(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(t, db)
	ps := CreatePreview(screenSrc, mc.ID, t, router)
	assert.Greater(t, ps.ID, int64(0), "It should create an entry in the DB")

	url := fmt.Sprintf("/api/screens/%d", ps.ID)
	code, err := DeleteJson(url, router)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, err, "It should delete ok")

	check := models.Screen{}
	res := db.Find(&check, ps.ID)
	assert.NoError(t, res.Error, "Ensure we did not have an error")
	assert.Equal(t, check.ID, int64(0))
}

func TestScreensResourceCannotCreate(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ps := &models.Screen{
		Src: "ThisShouldBail",
		Idx: 1,
	}

	resObj := models.Screen{}
	code, err := PostJson("/api/screens", ps, &resObj, router)
	fmt.Printf("What the actual fuck %s", resObj)
	assert.NotEqual(t, http.StatusCreated, code)
	assert.Error(t, err, "It Should error on post")
}

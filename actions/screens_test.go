package actions

import (
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"fmt"
	"net/http"
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
	cfg := utils.GetCfg()
	srcDir, dstDir, testFile := test_common.Get_VideoAndSetupPaths(cfg)
	c := &models.Container{
		Total: 4,
		Path:  filepath.Dir(srcDir),
		Name:  filepath.Base(srcDir),
	}
	assert.NoError(t, db.Create(c).Error, "Failed to create container")

	// TODO: Ensure that this path is actually correct, should actually make a REAL jpeg copy
	screenSrc := filepath.Join(dstDir, fmt.Sprintf("%s.screen.001.jpg", testFile))
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

/*
func (as *ActionSuite) Test_ScreensResource_ListMC() {
	test_common.InitFakeApp(true)

	// This creates a preview screen making the total 3 in the DB
	// Note it also resets the container_preview dir right now
	CreateScreen(as)

	_, mc1, _ := CreateScreen(as)
	CreatePreview("A", mc1.ID, as)
	res := as.JSON(fmt.Sprintf("/contents/%s/screens", mc1.ID.String())).Get()
	as.Equal(http.StatusOK, res.Code)

	validate := ScreensResponse{}
	json.NewDecoder(res.Body).Decode(&validate)
	as.Equal(len(validate.Results), 2, "Note we should have only two screens")
	as.Equal(validate.Total, 2, "Count should be correct")
	for _, ps := range validate.Results {
		as.Equal(ps.ContentID, mc1.ID)
		as.Equal(ps.Path, "") // Path should not be visible in the API
	}
}

func (as *ActionSuite) Test_ScreensResource_Show() {
	test_common.InitFakeApp(true)
	_, _, ps := CreateScreen(as)

	res := as.JSON(fmt.Sprintf("/screens/%s", ps.ID.String())).Get()
	as.Equal(http.StatusOK, res.Code)

	// Need to make it host the file.
	header := res.Header()
	as.Equal("image/jpeg", header.Get("Content-Type"))
}

// TODO: Create a screen that is actually on disk.
func (as *ActionSuite) Test_ScreensResource_Create() {
	test_common.InitFakeApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(as)
	ps := CreatePreview(screenSrc, mc.ID, as)
	as.Equal(ps.Src, screenSrc)

	screens := models.Screens{}
	as.DB.Where("content_id = ?", mc.ID).All(&screens)
	as.Equal(len(screens), 1, "There should be a screen in the DB")
}

func (as *ActionSuite) Test_ScreensResource_Update() {
	test_common.InitFakeApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(as)
	ps := CreatePreview(screenSrc, mc.ID, as)
	ps.Src = "UP"
	res := as.JSON(fmt.Sprintf("/screens/%s", ps.ID.String())).Put(ps)
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_ScreensResource_Destroy() {
	test_common.InitFakeApp(true)
	_, mc, screenSrc := CreateTestContainerWithContent(as)
	ps := CreatePreview(screenSrc, mc.ID, as)

	del_res := as.JSON(fmt.Sprintf("/screens/%s", ps.ID.String())).Delete()
	as.Equal(http.StatusOK, del_res.Code)
}

func (as *ActionSuite) Test_ScreensResource_CannotCreate() {
	test_common.InitFakeApp(false)
	ps := &models.Screen{
		Src: "Shouldn't Allow Create",
		Idx: 1,
	}
	res := as.JSON("/screens/").Post(ps)
	as.Equal(http.StatusCreated, res.Code)
}
*/

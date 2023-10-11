package actions

import (
	"fmt"
	"net/http"
	"os"

	//    "net/url"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"path/filepath"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

func CreatePreview(src string, contentID uuid.UUID, as *ActionSuite) models.Screen {
	mc := &models.Screen{
		Src:       src,
		ContentID: contentID,
		Idx:       1,
	}
	res := as.JSON("/screens").Post(mc)
	as.Equal(http.StatusCreated, res.Code)

	resObj := models.Screen{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj

}

// Kind of a pain in the ass to create all the way down to a valid preview screen
func CreateTestContainerWithContent(as *ActionSuite) (*models.Container, *models.Content, string) {
	cfg := utils.GetCfg()
	srcDir, dstDir, testFile := test_common.Get_VideoAndSetupPaths(cfg)
	c := &models.Container{
		Total: 4,
		Path:  filepath.Dir(srcDir),
		Name:  filepath.Base(srcDir),
	}
	as.DB.Create(c)

	// TODO: Ensure that this path is actually correct, should actually make a REAL jpeg copy
	screenSrc := filepath.Join(dstDir, fmt.Sprintf("%s.screen.001.jpg", testFile))
	mc := &models.Content{
		Src:         testFile,
		ContentType: "video/mp4",
		Preview:     screenSrc,
		ContainerID: nulls.NewUUID(c.ID),
	}
	as.DB.Create(mc)

	fmt.Printf("Screen src %s", screenSrc)
	f, err := os.Create(screenSrc)
	if err != nil {
		as.T().Errorf("Couldn't write to %s", screenSrc)
	}
	_, wErr := f.WriteString("Totally a screen")
	if wErr != nil {
		as.T().Errorf("Create a fake screen file on disk %s", screenSrc)
	}
	f.Sync()
	f.Close()
	return c, mc, screenSrc
}

func CreateScreen(as *ActionSuite) (*models.Container, *models.Content, *models.Screen) {
	c, mc, screenSrc := CreateTestContainerWithContent(as)
	ps := CreatePreview(screenSrc, mc.ID, as)
	return c, mc, &ps
}

func (as *ActionSuite) Test_ScreensResource_List() {
	test_common.InitFakeApp(true)
	CreateScreen(as)
	CreateScreen(as)

	res := as.JSON("/screens/").Get()
	as.Equal(http.StatusOK, res.Code)

	validate := ScreensResponse{}
	json.NewDecoder(res.Body).Decode(&validate)
	as.Equal(len(validate.Screens), 2, "There should be two preview screens")
}

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
	as.Equal(len(validate.Screens), 2, "Note we should have only two screens")
	as.Equal(validate.Count, 2, "Count should be correct")
	for _, ps := range validate.Screens {
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

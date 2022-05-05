package actions

import (
    "fmt"
    "net/http"
//    "net/url"
    "encoding/json"
    "path/filepath"
    "contented/internals"
    "contented/models"
    "github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
)

func CreatePreview(src string, mediaID uuid.UUID, as *ActionSuite) models.PreviewScreen {
     mc := &models.PreviewScreen{
         Src:     src,
         MediaID: mediaID,
         Idx: 1,
     }
     res := as.JSON("/screens").Post(mc)
     as.Equal(http.StatusCreated, res.Code)

     resObj := models.PreviewScreen{}
     json.NewDecoder(res.Body).Decode(&resObj)
     return resObj

}

// Kind of a pain in the ass to create all the way down to a valid preview screen
func CreateTestContainerWithMedia(as *ActionSuite) (*models.Container, *models.MediaContainer) {
    srcDir, dstDir, testFile := internals.Get_VideoAndSetupPaths()
    c := &models.Container{
        Total: 4,
        Path:  filepath.Dir(srcDir),
        Name:  filepath.Base(srcDir),
    }
    as.DB.Create(c)
    mc := &models.MediaContainer{
      Src:         testFile,
      ContentType: "video/mp4",
      Preview:     fmt.Sprintf("%s.%s.png", dstDir, testFile),
      ContainerID: nulls.NewUUID(c.ID),
    }
    as.DB.Create(mc)
    return c, mc
}

/*
func (as *ActionSuite) Test_PreviewScreensResource_List() {
	as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_PreviewScreensResource_Show() {
	as.Fail("Not Implemented!")
}
*/

func (as *ActionSuite) Test_PreviewScreensResource_Create() {
    internals.InitFakeApp(true)
    c, mc := CreateTestContainerWithMedia(as)

    fmt.Printf("What is up with this %s %s with %s", c.ID, c.Path, mc.ID)

    // TODO: Create a screen that is actually on disk.
    screenSrc := "container_previews/fakescreen.png"
    ps := CreatePreview(screenSrc, mc.ID, as)
    as.Equal(ps.Src, screenSrc)
}

/*
func (as *ActionSuite) Test_PreviewScreensResource_Update() {
	as.Fail("Not Implemented!")
}

func (as *ActionSuite) Test_PreviewScreensResource_Destroy() {
	as.Fail("Not Implemented!")
}
*/

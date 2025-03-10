package actions

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"contented/pkg/config"
	"contented/pkg/managers"
	"contented/pkg/models"
	"contented/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Screen)
// DB Table: Plural (screens)
// Resource: Plural (Screens)
// Path: Plural (/screens)
// View Template Folder: Plural (/templates/screens/)
type ScreensResponse struct {
	Total   int64          `json:"total" default:"0"`
	Results models.Screens `json:"results" default:"[]"`
}

// List gets all Screens. This function is mapped to the path
// GET /screens
func ScreensResourceList(c *gin.Context) {
	// Get the DB connection from the context
	mcStrID := managers.StringDefault(c.Param("content_id"), "")
	log.Printf("Content ID specified %s", mcStrID)
	if mcStrID != "" {
		_, err := strconv.ParseInt(mcStrID, 10, 64)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	// TODO: Screens Response (total count provided)
	man := managers.GetManager(c)
	screens, total, err := man.ListScreensContext()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if screens == nil {
		screens = &models.Screens{}
	}
	res := ScreensResponse{
		Total:   total,
		Results: *screens,
	}
	c.JSON(200, res)
}

// Show gets the data for one Screen. This function is mapped to
// the path GET /screens/{screen_id}
func ScreensResourceShow(c *gin.Context) {
	psStrID := c.Param("screen_id")
	psID, badUUID := strconv.ParseInt(psStrID, 10, 64)
	if badUUID != nil {
		c.AbortWithError(400, badUUID)
		return
	}

	man := managers.GetManager(c)
	screen, err := man.GetScreen(psID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	// Check it exists
	fqPath := screen.GetFqPath()
	_, fErr := os.Stat(fqPath)
	if fErr != nil {
		log.Printf("Cannot download Screen file not on disk %s with err %s", fqPath, fErr)
		c.AbortWithError(404, fErr)
		return
	}

	// TODO: Figure out the headers better
	log.Printf("Preview Screen ID specified %s path %s", psStrID, fqPath)
	c.File(fqPath)
}

// Create adds a Screen to the DB. This function is mapped to the
// path POST /screens
func ScreensResourceCreate(c *gin.Context) {
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	// Bind previewScreen to the html form/JSON elements
	screen := &models.Screen{}
	if err := c.BindJSON(screen); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// We don't allow a src to provide a path
	src := filepath.Base(screen.Src)
	if utils.HasUpwardTraversal(src) {
		c.AbortWithError(http.StatusBadRequest, errors.New("src cannot contain upward traversal"))
		return
	}
	screen.Src = src

	// Ensure that the content exists and is under a container so we can find where the preview should exist
	content, err := man.GetContent(screen.ContentID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	cnt, err := man.GetContainer(*content.ContainerID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// TODO: make the preview directory name configurable or something that can be passed via the API
	dstPath := cnt.GetFqPath()
	screen.Path = filepath.Join(dstPath, config.PREVIEW_DIRECTORY)

	cErr := man.CreateScreen(screen)
	if cErr != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, cErr)
		return
	}
	c.JSON(http.StatusCreated, screen)
}

// Update changes a Screen in the DB. This function is mapped to
// the path PUT /screens/{screen_id}
func ScreensResourceUpdate(c *gin.Context) {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}

	man := managers.GetManager(c)
	id, idErr := strconv.ParseInt(c.Param("screen_id"), 10, 64)
	if idErr != nil {
		c.AbortWithError(http.StatusBadRequest, idErr)
		return
	}
	screen, notFoundErr := man.GetScreen(id)
	if notFoundErr != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	if err := c.BindJSON(screen); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
		return
	}
	if upErr := man.UpdateScreen(screen); upErr != nil {
		c.JSON(http.StatusInternalServerError, upErr)
		return
	}
	checkScreen, _ := man.GetScreen(id)
	c.JSON(http.StatusOK, checkScreen)
}

// Destroy deletes a Screen from the DB. This function is mapped
// to the path DELETE /screens/{screen_id}
func ScreensResourceDestroy(c *gin.Context) {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}

	screenId, sErr := strconv.ParseInt(c.Param("screen_id"), 10, 64)
	if sErr != nil {
		c.AbortWithError(http.StatusBadRequest, sErr)
		return
	}

	man := managers.GetManager(c)
	screen, dErr := man.DestroyScreen(screenId)
	if dErr != nil {
		c.JSON(http.StatusBadRequest, dErr)
		return
	}
	c.JSON(http.StatusOK, screen)
}

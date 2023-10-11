package actions

import (
	"log"
	"os"

	//"fmt"
	"net/http"
	// "errors"
	"contented/managers"
	"contented/models"

	"github.com/gobuffalo/buffalo"

	//"github.com/gobuffalo/pop/v5"

	"github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Screen)
// DB Table: Plural (screens)
// Resource: Plural (Screens)
// Path: Plural (/screens)
// View Template Folder: Plural (/templates/screens/)
type ScreensResponse struct {
	Count   int            `json:"count" default:"0"`
	Screens models.Screens `json:"screens" default:"[]"`
}

// ScreensResource is the resource for the Screen model
type ScreensResource struct {
	buffalo.Resource
}

// List gets all Screens. This function is mapped to the path
// GET /screens
func (v ScreensResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	mcStrID := managers.StringDefault(c.Param("content_id"), "")
	log.Printf("Content ID specified %s", mcStrID)
	if mcStrID != "" {
		_, err := uuid.FromString(mcStrID)
		if err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
	}
	// TODO: Screens Response (total count provided)
	man := managers.GetManager(&c)
	screens, count, err := man.ListScreensContext()
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	if screens == nil {
		screens = &models.Screens{}
	}
	res := ScreensResponse{
		Count:   count,
		Screens: *screens,
	}
	return c.Render(200, r.JSON(res))
}

// Show gets the data for one Screen. This function is mapped to
// the path GET /screens/{screen_id}
func (v ScreensResource) Show(c buffalo.Context) error {
	psStrID := c.Param("screen_id")
	psID, badUUID := uuid.FromString(psStrID)
	if badUUID != nil {
		return c.Error(400, badUUID)
	}

	man := managers.GetManager(&c)
	screen, err := man.GetScreen(psID)
	if err != nil {
		return c.Error(404, err)
	}

	// Check it exists
	fqPath := screen.GetFqPath()
	_, fErr := os.Stat(fqPath)
	if fErr != nil {
		log.Printf("Cannot download file not on disk %s with err %s", fqPath, fErr)
		return c.Error(404, err)
	}
	log.Printf("Preview Screen ID specified %s path %s", psStrID, fqPath)
	http.ServeFile(c.Response(), c.Request(), fqPath)
	return nil
}

// Create adds a Screen to the DB. This function is mapped to the
// path POST /screens
func (v ScreensResource) Create(c buffalo.Context) error {
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}
	// Bind previewScreen to the html form/JSON elements
	screen := &models.Screen{}
	if err := c.Bind(screen); err != nil {
		return err
	}
	man := managers.GetManager(&c)
	cErr := man.CreateScreen(screen)
	if cErr != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(cErr))
	}
	return c.Render(http.StatusCreated, r.JSON(screen))
}

// Update changes a Screen in the DB. This function is mapped to
// the path PUT /screens/{screen_id}
func (v ScreensResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}

	man := managers.GetManager(&c)
	id, idErr := uuid.FromString(c.Param("screen_id"))
	if idErr != nil {
		return c.Error(http.StatusBadRequest, idErr)
	}
	screen, notFoundErr := man.GetScreen(id)
	if notFoundErr != nil {
		return c.Error(http.StatusNotFound, err)
	}
	if err := c.Bind(screen); err != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(err))
	}
	checkScreen, _ := man.GetScreen(id)
	return c.Render(http.StatusOK, r.JSON(checkScreen))
}

// Destroy deletes a Screen from the DB. This function is mapped
// to the path DELETE /screens/{screen_id}
func (v ScreensResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}

	man := managers.GetManager(&c)
	screen, dErr := man.DestroyScreen(c.Param("screen_id"))
	if dErr != nil {
		return c.Render(http.StatusBadRequest, r.JSON(dErr))
	}
	return c.Render(http.StatusOK, r.JSON(screen))
}

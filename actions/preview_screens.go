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
    "github.com/gobuffalo/x/responder"
    "github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Screen)
// DB Table: Plural (preview_screens)
// Resource: Plural (Screens)
// Path: Plural (/preview_screens)
// View Template Folder: Plural (/templates/preview_screens/)

// ScreensResource is the resource for the Screen model
type ScreensResource struct {
    buffalo.Resource
}

// List gets all Screens. This function is mapped to the path
// GET /preview_screens
func (v ScreensResource) List(c buffalo.Context) error {
    // Get the DB connection from the context
    var previewScreens *models.Screens
    var err error

    mcStrID := c.Param("content_id")
    log.Printf("Content ID specified %s", mcStrID)

    man := managers.GetManager(&c)
    if mcStrID != "" {
        mcID, err := uuid.FromString(mcStrID)
        if err != nil {
            return c.Error(http.StatusBadRequest, err)
        }
        previewScreens, err = man.ListScreensContext(mcID)

    } else {
        previewScreens, err = man.ListAllScreensContext()
        if err != nil {
            return c.Error(http.StatusBadRequest, err)
        }
    }

    return responder.Wants("html", func(c buffalo.Context) error {
        // Add the paginator to the context so it can be used in the template.
        return c.Render(200, r.JSON(previewScreens))
    }).Wants("json", func(c buffalo.Context) error {
        return c.Render(200, r.JSON(previewScreens))
    }).Wants("xml", func(c buffalo.Context) error {
        return c.Render(200, r.XML(previewScreens))
    }).Respond(c)
}

// Show gets the data for one Screen. This function is mapped to
// the path GET /preview_screens/{preview_screen_id}
func (v ScreensResource) Show(c buffalo.Context) error {
    psStrID := c.Param("preview_screen_id")
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
// path POST /preview_screens
func (v ScreensResource) Create(c buffalo.Context) error {
    _, tx, err := managers.ManagerCanCUD(&c)
    if err != nil {
        return err
    }
    // Allocate an empty Screen
    previewScreen := &models.Screen{}

    // Bind previewScreen to the html form/JSON elements
    if err := c.Bind(previewScreen); err != nil {
        return err
    }

    // Validate the data from the html form
    verrs, err := tx.ValidateAndCreate(previewScreen)
    if err != nil {
        return err
    }

    if verrs.HasAny() {
        return responder.Wants("html", func(c buffalo.Context) error {
            // Make the errors available inside the html template
            c.Set("errors", verrs)

            // Render again the new.html template that the user can
            // correct the input.
            c.Set("previewScreen", previewScreen)

            return c.Render(http.StatusUnprocessableEntity, r.HTML("/preview_screens/new.plush.html"))
        }).Wants("json", func(c buffalo.Context) error {
            return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
        }).Wants("xml", func(c buffalo.Context) error {
            return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
        }).Respond(c)
    }

    return responder.Wants("html", func(c buffalo.Context) error {
        // If there are no errors set a success message
        return c.Render(http.StatusCreated, r.JSON(previewScreen))
    }).Wants("json", func(c buffalo.Context) error {
        return c.Render(http.StatusCreated, r.JSON(previewScreen))
    }).Wants("xml", func(c buffalo.Context) error {
        return c.Render(http.StatusCreated, r.XML(previewScreen))
    }).Respond(c)
}

// Update changes a Screen in the DB. This function is mapped to
// the path PUT /preview_screens/{preview_screen_id}
func (v ScreensResource) Update(c buffalo.Context) error {
    // Get the DB connection from the context
    _, tx, err := managers.ManagerCanCUD(&c)
    if err != nil {
        return err
    }

    // Allocate an empty Screen
    previewScreen := &models.Screen{}

    if err := tx.Find(previewScreen, c.Param("preview_screen_id")); err != nil {
        return c.Error(http.StatusNotFound, err)
    }

    // Bind Screen to the html form elements
    if err := c.Bind(previewScreen); err != nil {
        return err
    }

    verrs, err := tx.ValidateAndUpdate(previewScreen)
    if err != nil {
        return err
    }

    if verrs.HasAny() {
        return responder.Wants("html", func(c buffalo.Context) error {
            // Make the errors available inside the html template
            c.Set("errors", verrs)

            // Render again the edit.html template that the user can
            // correct the input.
            c.Set("previewScreen", previewScreen)

            return c.Render(http.StatusUnprocessableEntity, r.HTML("/preview_screens/edit.plush.html"))
        }).Wants("json", func(c buffalo.Context) error {
            return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
        }).Wants("xml", func(c buffalo.Context) error {
            return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
        }).Respond(c)
    }

    return responder.Wants("html", func(c buffalo.Context) error {
        // If there are no errors set a success message
        //c.Flash().Add("success", T.Translate(c, "previewScreen.updated.success"))

        // and redirect to the show page
        return c.Redirect(http.StatusSeeOther, "/preview_screens/%v", previewScreen.ID)
    }).Wants("json", func(c buffalo.Context) error {
        return c.Render(http.StatusOK, r.JSON(previewScreen))
    }).Wants("xml", func(c buffalo.Context) error {
        return c.Render(http.StatusOK, r.XML(previewScreen))
    }).Respond(c)
}

// Destroy deletes a Screen from the DB. This function is mapped
// to the path DELETE /preview_screens/{preview_screen_id}
func (v ScreensResource) Destroy(c buffalo.Context) error {
    // Get the DB connection from the context
    _, tx, err := managers.ManagerCanCUD(&c)
    if err != nil {
        return err
    }

    // Allocate an empty Screen
    previewScreen := &models.Screen{}

    // To find the Screen the parameter preview_screen_id is used.
    if err := tx.Find(previewScreen, c.Param("preview_screen_id")); err != nil {
        return c.Error(http.StatusNotFound, err)
    }

    if err := tx.Destroy(previewScreen); err != nil {
        return err
    }

    return responder.Wants("html", func(c buffalo.Context) error {
        // If there are no errors set a flash message
        //c.Flash().Add("success", T.Translate(c, "previewScreen.destroyed.success"))
        // Redirect to the index page
        return c.Redirect(http.StatusSeeOther, "/preview_screens")
    }).Wants("json", func(c buffalo.Context) error {
        return c.Render(http.StatusOK, r.JSON(previewScreen))
    }).Wants("xml", func(c buffalo.Context) error {
        return c.Render(http.StatusOK, r.XML(previewScreen))
    }).Respond(c)
}

package actions

import (
	"errors"

	//"fmt"
	"net/http"
	// "errors"
	"contented/managers"
	"contented/models"

	"github.com/gobuffalo/buffalo"
	//"github.com/gobuffalo/pop/v5"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Tag)
// DB Table: Plural (tags)
// Resource: Plural (Tags)
// Path: Plural (/tags)
// View Template Folder: Plural (/templates/tags/)

// TagsResource is the resource for the Tag model
type TagsResource struct {
	buffalo.Resource
}

// List gets all Tags. This function is mapped to the path
// GET /tags
func (v TagsResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	var previewTags *models.Tags
	var err error

	man := managers.GetManager(&c)
	previewTags, err = man.ListAllTagsContext()
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	return c.Render(200, r.JSON(previewTags))
}

// Show gets the data for one Tag. This function is mapped to
// the path GET /tags/{tag_id}
func (v TagsResource) Show(c buffalo.Context) error {

	tagID := c.Param("tag_id")
	if tagID == "" {
		return c.Error(400, errors.New("Requires ID"))
	}
	man := managers.GetManager(&c)
	tag, err := man.GetTag(tagID)
	if err != nil {
		return c.Error(404, err)
	}
	return c.Render(200, r.JSON(tag))
}

// Create adds a Tag to the DB. This function is mapped to the
// path POST /tags
func (v TagsResource) Create(c buffalo.Context) error {
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}
	// Bind previewTag to the html form/JSON elements
	tag := &models.Tag{}
	if err := c.Bind(tag); err != nil {
		return err
	}
	man := managers.GetManager(&c)
	cErr := man.CreateTag(tag)
	if cErr != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(cErr))
	}
	return c.Render(http.StatusCreated, r.JSON(tag))
}

// Update changes a Tag in the DB. This function is mapped to
// the path PUT /tags/{tag_id}
func (v TagsResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}

	man := managers.GetManager(&c)
	id := c.Param("tag_id")
	tag, notFoundErr := man.GetTag(id)
	if notFoundErr != nil {
		return c.Error(http.StatusNotFound, err)
	}
	if err := c.Bind(tag); err != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(err))
	}
	upErr := man.UpdateTag(tag)
	if upErr != nil {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(upErr))
	}
	checkTag, _ := man.GetTag(id)
	return c.Render(http.StatusOK, r.JSON(checkTag))
}

// Destroy deletes a Tag from the DB. This function is mapped
// to the path DELETE /tags/{tag_id}
func (v TagsResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}

	// TODO: Implement
	man := managers.GetManager(&c)
	tag, dErr := man.DestroyTag(c.Param("tag_id"))
	if dErr != nil {
		return c.Render(http.StatusBadRequest, r.JSON(dErr))
	}
	return c.Render(http.StatusOK, r.JSON(tag))
}

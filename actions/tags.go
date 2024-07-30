package actions

import (
	"errors"

	"contented/managers"
	"contented/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/buffalo"
)

// TagsResource is the resource for the Tag model
type TagsResource struct {
	buffalo.Resource
}

type TagResponse struct {
	Total   int         `json:"total"`
	Results models.Tags `json:"results"`
}

// List gets all Tags. This function is mapped to the path
// GET /tags
func (v TagsResource) List(c *gin.Context) {
	// Get the DB connection from the context
	man := managers.GetManager(c)
	previewTags, total, err := man.ListAllTagsContext()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	tr := TagResponse{
		Total:   total,
		Results: *previewTags,
	}
	c.JSON(200, r.JSON(tr))
}

// Show gets the data for one Tag. This function is mapped to
// the path GET /tags/{tag_id}
func (v TagsResource) Show(c *gin.Context) {
	tagID := c.Param("tag_id")
	if tagID == "" {
		c.AbortWithError(400, errors.New("Requires ID"))
		return
	}
	man := managers.GetManager(c)
	tag, err := man.GetTag(tagID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	c.JSON(200, r.JSON(tag))
}

// Create adds a Tag to the DB. This function is mapped to the
// path POST /tags
func (v TagsResource) Create(c *gin.Context) {
	_, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	// Bind previewTag to the html form/JSON elements
	tag := &models.Tag{}
	if err := c.Bind(tag); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	man := managers.GetManager(c)
	cErr := man.CreateTag(tag)
	if cErr != nil {
		c.JSON(http.StatusUnprocessableEntity, r.JSON(cErr))
		return
	}
	c.JSON(http.StatusCreated, r.JSON(tag))
}

// Update changes a Tag in the DB. This function is mapped to
// the path PUT /tags/{tag_id}
func (v TagsResource) Update(c *gin.Context) {
	// Get the DB connection from the context
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}

	id := c.Param("tag_id")
	tag, notFoundErr := man.GetTag(id)
	if notFoundErr != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	if err := c.Bind(tag); err != nil {
		c.JSON(http.StatusUnprocessableEntity, r.JSON(err))
		return
	}
	upErr := man.UpdateTag(tag)
	if upErr != nil {
		c.JSON(http.StatusUnprocessableEntity, r.JSON(upErr))
		return
	}
	checkTag, _ := man.GetTag(id)
	c.JSON(http.StatusOK, r.JSON(checkTag))
}

// Destroy deletes a Tag from the DB. This function is mapped
// to the path DELETE /tags/{tag_id}
func (v TagsResource) Destroy(c *gin.Context) {
	// Get the DB connection from the context
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	tag, dErr := man.DestroyTag(c.Param("tag_id"))
	if dErr != nil {
		c.JSON(http.StatusBadRequest, r.JSON(dErr))
		return
	}
	c.JSON(http.StatusOK, r.JSON(tag))
}

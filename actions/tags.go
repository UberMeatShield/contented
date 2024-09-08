package actions

import (
	"encoding/json"
	"errors"

	"contented/managers"
	"contented/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TagResponse struct {
	Total   int64       `json:"total"`
	Results models.Tags `json:"results"`
}

func (t TagResponse) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// List gets all Tags. This function is mapped to the path
// GET /tags
func TagsResourceList(c *gin.Context) {
	// Get the DB connection from the context
	man := managers.GetManager(c)

	previewTags, total, err := man.ListAllTagsContext()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if previewTags == nil {
		previewTags = &models.Tags{}
	}
	tr := TagResponse{
		Total:   total,
		Results: *previewTags,
	}
	c.JSON(200, tr)
}

// Show gets the data for one Tag. This function is mapped to
// the path GET /tags/{tag_id}
func TagsResourceShow(c *gin.Context) {
	tagID := c.Param("tag_id")
	if tagID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("requires ID"))
		return
	}
	man := managers.GetManager(c)
	tag, err := man.GetTag(tagID)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.JSON(200, tag)
}

// Create adds a Tag to the DB. This function is mapped to the
// path POST /tags
func TagsResourceCreate(c *gin.Context) {
	_, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	// Bind previewTag to the html form/JSON elements
	tag := &models.Tag{}
	if err := c.BindJSON(tag); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	man := managers.GetManager(c)
	if cErr := man.CreateTag(tag); cErr != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, cErr)
		return
	}
	c.JSON(http.StatusCreated, tag)
}

// Update changes a Tag in the DB. This function is mapped to
// the path PUT /tags/{tag_id}
func TagsResourceUpdate(c *gin.Context) {
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
	if err := c.BindJSON(tag); err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	upErr := man.UpdateTag(tag)
	if upErr != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, upErr)
		return
	}
	checkTag, _ := man.GetTag(id)
	c.JSON(http.StatusOK, checkTag)
}

// Destroy deletes a Tag from the DB. This function is mapped
// to the path DELETE /tags/{tag_id}
func TagsResourceDestroy(c *gin.Context) {
	// Get the DB connection from the context
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	tag, dErr := man.DestroyTag(c.Param("tag_id"))
	if dErr != nil {
		c.AbortWithError(http.StatusBadRequest, dErr)
		return
	}
	c.JSON(http.StatusOK, tag)
}

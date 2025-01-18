package actions

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"contented/pkg/managers"
	"contented/pkg/models"

	"github.com/gin-gonic/gin"
)

type ContentsResponse struct {
	Total   int64           `json:"total"`
	Results models.Contents `json:"results"`
}

func (t ContentsResponse) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// List gets all Contents. This function is mapped to the path
// GET /contents
func ContentsResourceList(c *gin.Context) {
	// Optional params suuuuck in GoLang
	cIDStr := c.Param("container_id")
	if cIDStr != "" {
		_, err := strconv.ParseInt(cIDStr, 10, 64)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	// The managers are going to be rough
	man := managers.GetManager(c)
	contents, total, cErr := man.ListContentContext()
	if cErr != nil {
		c.AbortWithError(http.StatusInternalServerError, cErr)
		return
	}
	log.Printf("Contents loaded found %d elements", total)
	if contents == nil {
		contents = &models.Contents{}
	}
	cr := ContentsResponse{Total: total, Results: *contents}
	c.JSON(200, cr)
}

// Show gets the data for one Content. This function is mapped to
// the path GET /contents/{content_id}
func ContentsResourceShow(c *gin.Context) {
	man := managers.GetManager(c)

	// TODO: Make it actually just handle /content (page, number)
	id, err := strconv.ParseInt(c.Param("content_id"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	contentContainer, missing_err := man.GetContent(id)
	if missing_err != nil {
		c.AbortWithError(http.StatusNotFound, missing_err)
		return
	}
	c.Header("Last-Modified", contentContainer.UpdatedAt.UTC().Format(http.TimeFormat))
	c.JSON(http.StatusOK, *contentContainer)
}

// Create adds a Content to the DB. This function is mapped to the
// path POST /contents
func ContentsResourceCreate(c *gin.Context) {
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	// Allocate an empty Content
	// Bind contentContainer to the html form elements (probably not required?)
	content := &models.Content{}
	if err := c.BindJSON(content); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err = man.CreateContent(content)
	if err != nil {
		log.Printf("Failed to create content with error %s container ID %s", err, content)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	validate, checkErr := man.GetContent(content.ID)
	if checkErr != nil {
		c.AbortWithError(http.StatusExpectationFailed, checkErr)
		return
	}
	c.JSON(http.StatusCreated, validate)
}

// Update changes a Content in the DB. This function is mapped to
// the path PUT /contents/{content_id}
func ContentsResourceUpdate(c *gin.Context) {
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}

	id, _ := strconv.ParseInt(c.Param("content_id"), 10, 64)
	exists, err := man.GetContent(id)
	if err != nil || exists == nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	// Bind Content to the html form elements (Nuke this? or change to json)
	content := *exists
	if err := c.BindJSON(&content); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	upErr := man.UpdateContent(&content)
	if upErr != nil {
		log.Printf("Failed to update resource %s", upErr)
		c.AbortWithError(http.StatusBadRequest, upErr)
		return
	}
	checkContent, checkErr := man.GetContent(content.ID)
	if checkErr != nil {
		c.AbortWithError(http.StatusExpectationFailed, checkErr)
		return
	}
	c.JSON(http.StatusOK, checkContent)
}

// Destroy deletes a Content from the DB. This function is mapped
// to the path DELETE /contents/{content_id}
func ContentsResourceDestroy(c *gin.Context) {
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	// TODO: Manager should ABSOLUTELY be the thing doing updates etc.
	// Allocate an empty Content
	id, argErr := strconv.ParseInt(c.Param("content_id"), 10, 64)
	if argErr != nil {
		c.AbortWithError(http.StatusBadRequest, argErr)
		return
	}
	content, err := man.DestroyContent(id)
	if err != nil {
		if content == nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		} else {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	c.JSON(http.StatusOK, content)
}

// Destroy deletes a Content from the DB. This function is mapped
// to the path DELETE /contents/{content_id}
func ContentScreensDestroy(c *gin.Context) {
	man, _, err := managers.ManagerCanCUD(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}

	// TODO: Manager should ABSOLUTELY be the thing doing updates etc.
	id, argErr := strconv.ParseInt(c.Param("content_id"), 10, 64)
	if argErr != nil {
		c.AbortWithError(http.StatusBadRequest, argErr)
		return
	}

	content, err := man.GetContent(id)
	if err != nil {
		if content == nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		} else {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	if content == nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	err = managers.RemoveScreensForContent(man, content.ID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	content.Screens = models.Screens{}
	c.JSON(http.StatusOK, content)
}

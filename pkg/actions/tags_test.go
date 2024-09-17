package actions

import (
	"contented/pkg/models"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateTag(name string, t *testing.T, router *gin.Engine) models.Tag {
	tag := &models.Tag{
		ID: name,
	}
	resObj := models.Tag{}
	code, err := PostJson("/api/tags", tag, &resObj, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Error creating %s", err))
	assert.NoError(t, err, "It should not throw an error")
	return resObj
}

func TestTagsResourceListDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagList(t, router)
}

func TestTagsResourceListMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagList(t, router)
}

func ValidateTagList(t *testing.T, router *gin.Engine) {
	CreateTag("A", t, router)
	CreateTag("B", t, router)

	tags := TagResponse{}
	code, err := GetJson("/api/tags", "", &tags, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed to load tags %s", err))
	assert.Equal(t, len(tags.Results), 2, fmt.Sprintf("There should be two tags %s", tags.Results))
}

func TestTagsResourceShowDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagShow(t, router)
}

func TestTagsResourceShowMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTagShow(t, router)
}

func ValidateTagShow(t *testing.T, router *gin.Engine) {
	tag := CreateTag("A", t, router)
	url := fmt.Sprintf("/api/tags/%s", tag.ID)
	check := models.Tag{}
	code, err := GetJson(url, "", &check, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("It should find the tag %s", err))
	assert.NoError(t, err)
}

func TestTagsResourceCreateDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagCreate(t, router)
}

func TestTagsResourceCreateMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTagCreate(t, router)
}

func ValidateTagCreate(t *testing.T, router *gin.Engine) {
	CreateTag("Monkey", t, router)
	CreateTag("Dupe", t, router)
	tag := &models.Tag{
		ID: "Dupe",
	}
	resObj := models.Tag{}
	codeDupe, errDupe := PostJson("/api/tags", tag, &resObj, router)
	assert.Equal(t, http.StatusUnprocessableEntity, codeDupe, fmt.Sprintf("Error creating %s", errDupe))

	checkTags := TagResponse{}
	code, err := GetJson("/api/tags", "", &checkTags, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("It should find the tag %s", err))
	assert.Equal(t, 2, len(checkTags.Results), fmt.Sprintf("We should have a tag %s", checkTags))
}

func TestTagsResourceUpdateDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagsUpdate(t, router)
}

func TestTagsResourceUpdateMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTagsUpdate(t, router)
}

func ValidateTagsUpdate(t *testing.T, router *gin.Engine) {
	tag := CreateTag("Tag", t, router)
	tag.Description = "Updated"
	url := fmt.Sprintf("/api/tags/%s", tag.ID)

	check := models.Tag{}
	code, err := PutJson(url, tag, &check, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("It should update %s", err))

	assert.NoError(t, err)
	assert.Equal(t, "Updated", check.Description)
}

func TestTagsResourceDestroyDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateTagsDelete(t, router)
}

func TestTagsResourceDestroyMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTagsDelete(t, router)
}

func ValidateTagsDelete(t *testing.T, router *gin.Engine) {
	tag := CreateTag("A", t, router)

	url := fmt.Sprintf("/api/tags/%s", tag.ID)
	code, err := DeleteJson(url, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed to delete %s", err))

	tagCheck := TagResponse{}
	codeCheck, cErr := GetJson("/api/tags", "", &tagCheck, router)
	assert.Equal(t, http.StatusOK, codeCheck, fmt.Sprintf("It should get tags fine %s", cErr))

	assert.NoError(t, cErr, "Getting tags should be fine")
	assert.Equal(t, len(tagCheck.Results), 0, "It should be deleted")
}

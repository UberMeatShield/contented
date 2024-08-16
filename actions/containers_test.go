package actions

// These tests are DB based tests, vs in memory manager test_common.InitFakeApp(true)

import (
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateContainer(name string, t *testing.T, router *gin.Engine) models.Container {
	cfg := utils.GetCfg()
	c := &models.Container{
		Total: 1,
		Name:  name,
		Path:  cfg.Dir,
	}
	fqPath, err := test_common.CreateContainerPath(c)
	log.Printf("What is the fqContainer path %s", fqPath)
	if err != nil {
		fmt.Printf("Failed to create path %s with err %s", fqPath, err)
		panic(err)
	}

	resObj := &models.Container{}
	code, err := PostJson("/api/containers", c, &resObj, router)

	assert.NoError(t, err, "The Post was not a success")
	assert.Greater(t, resObj.ID, int64(0), "A container ID should exist")
	assert.Equal(t, resObj.Name, name, fmt.Sprintf("Did we get a valid object back %s", resObj))
	assert.Equal(t, http.StatusCreated, code, "The http post call was successful")

	if resObj.Name != name {
		t.Fatalf("The container did not get created correctly %s", resObj)
	}
	return *resObj
}

func TestContainersResourceList(t *testing.T) {
	test_common.InitFakeApp(false)
	router := setupRouter()

	url := "/api/containers"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Failed to call %s", url))

	containers := ContainersResponse{}
	json.NewDecoder(w.Body).Decode(&containers)
	assert.NotEmpty(t, containers.Results, "No containers returned from the call")
}

func TestContainersResourceShow(t *testing.T) {
	test_common.InitFakeApp(true)
	router := setupRouter()
	name := "ShowTest"

	cnt := CreateContainer(name, t, router)
	defer test_common.CleanupContainer(&cnt)
	assert.NotZero(t, cnt.ID)

	url := fmt.Sprintf("/api/containers/%d", cnt.ID)
	req, _ := http.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	resObj := models.Container{}
	json.NewDecoder(w.Body).Decode(&resObj)
	assert.Equal(t, resObj.Name, name)
}

func TestContainersResourceCreate(t *testing.T) {
	cfg, db := test_common.InitFakeApp(true)
	router := setupRouter()

	cnt := &models.Container{
		Total: 1,
		Name:  "dir3",
		Path:  "ShouldGetReset",
	}
	resObj := &models.Container{}
	code, err := PostJson("/api/containers", cnt, resObj, router)
	assert.NoError(t, err, "Failure in creating the container")
	assert.Equal(t, http.StatusCreated, code, "It should be able to create")
	defer test_common.CleanupContainer(cnt)

	assert.Equal(t, resObj.Name, cnt.Name)
	assert.NotZero(t, resObj.ID)

	// Path does not come back from the API (hidden), check it updated.
	check := models.Container{}
	assert.NoError(t, db.Find(&check, resObj.ID).Error, "Could not pull back a container from the DB")
	assert.Equal(t, check.Path, cfg.Dir, "It should reset our path")
}

func TestContainersResourceUpdate(t *testing.T) {
	test_common.InitFakeApp(true)
	router := setupRouter()
	cnt := CreateContainer("Initial", t, router)
	defer test_common.CleanupContainer(&cnt)
	assert.NotZero(t, cnt.ID)
	assert.Equal(t, cnt.Name, "Initial")

	description := "UpdateTest"
	cnt.Description = description

	url := fmt.Sprintf("/api/containers/%d", cnt.ID)
	check := &models.Container{}
	code, err := PutJson(url, cnt, check, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Error %s", err))
	assert.Equal(t, description, check.Description, "It should update the description")
}

func TestContainersResourceDestroy(t *testing.T) {
	router := setupRouter() // move this into Fake App?
	test_common.InitFakeApp(true)
	cnt := CreateContainer("Nuke", t, router)
	defer test_common.CleanupContainer(&cnt)
	assert.Equal(t, cnt.Name, "Nuke")
	assert.NotZero(t, cnt.ID)

	url := fmt.Sprintf("/api/containers/%d", cnt.ID)
	deleteCode, err := DeleteJson(url, router)
	assert.NoError(t, err, "It should delete")
	assert.Equal(t, http.StatusOK, deleteCode, "The delete operation failed")

	check := &models.Container{}
	code, err := GetJson(url, cnt, check, router)
	assert.Error(t, err, "It should be in an error state")
	assert.Equal(t, http.StatusNotFound, code, "It should not get a container back")
}

func TestContainerList(t *testing.T) {
	router := setupRouter() // move this into Fake App?
	_, db := test_common.InitFakeApp(true)

	cnt1, _ := test_common.GetContentByDirName("dir1")
	cnt2, _ := test_common.GetContentByDirName("dir2")
	db.Create(cnt1)
	db.Create(cnt2)

	containers := &ContainersResponse{}
	code, err := GetJson("/api/containers", "", containers, router)
	assert.NoError(t, err, "It should not error")
	assert.Equal(t, http.StatusOK, code)

	assert.Equal(t, 2, len(containers.Results), "It should have loaded two fixtures")
	var found *models.Container
	for _, c := range containers.Results {
		if c.Name == "dir2" {
			found = &c
		}
	}
	assert.NotNil(t, found, "If it had the fixture loaded we should have this name")
	url := fmt.Sprintf("/api/containers/%d/contents", found.ID)
	contents := &ContentsResponse{}
	contentsCode, contentsErr := GetJson(url, "", contents, router)
	assert.NoError(t, contentsErr, "It should have no content but should work")
	assert.Equal(t, http.StatusOK, contentsCode)
	assert.Equal(t, int64(0), contents.Total, "There should not be content")
}

func TestMemoryReadOnlyDenyEdit(t *testing.T) {
	router := setupRouter() // move this into Fake App?
	cfg, _ := test_common.InitFakeApp(false)
	cfg.ReadOnly = true
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	containers, count, err := man.ListContainersContext()
	assert.NoError(t, err, "It should list containers")
	assert.Greater(t, count, int64(0), "The count should be positive")
	assert.Greater(t, len(*containers), 0, "There should be containers")

	for _, c := range *containers {
		c.Name = "Update Should fail"
		url := fmt.Sprintf("/api/containers/%d", c.ID)
		code, err := PutJson(url, c, &models.Container{}, router)
		assert.Equal(t, http.StatusNotImplemented, code)
		assert.Error(t, err, "It should not work")
	}
}

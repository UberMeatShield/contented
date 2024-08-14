package actions

// These tests are DB based tests, vs in memory manager test_common.InitFakeApp(true)

import (
	"contented/test_common"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
func CreateContainer(name string) models.Container {
	cfg := utils.GetCfg()
	c := &models.Container{
		Total: 1,
		Name:  name,
		Path:  cfg.Dir,
	}
	fqPath, err := test_common.CreateContainerPath(c)
	if err != nil {
		fmt.Printf("Failed to create path %s with err %s", fqPath, err)
		panic(err)
	}
	res := as.JSON("/containers").Post(c)
	resObj := models.Container{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj
}
*/

func TestContainersResourceList(t *testing.T) {
	test_common.InitFakeApp(false)
	router := setupRouter()

	url := "/api/containers"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

/*
func Test_ContainersResource_Show(t *testing.T) {
	test_common.InitFakeApp(true)
	name := "ShowTest"
	cnt := CreateContainer(name, as)
	defer test_common.CleanupContainer(&cnt)
	assert.NotZero(t, cnt.ID)

	validate := as.JSON("/containers/" + cnt.ID.String()).Get()
	assert.Equal(t, http.StatusOK, validate.Code)

	resObj := models.Container{}
	json.NewDecoder(validate.Body).Decode(&resObj)
	assert.Equal(t, resObj.Name, name)
}

func Test_ContainersResource_Create(t *testing.T) {
	cfg := test_common.InitFakeApp(true)
	cnt := &models.Container{
		Total: 1,
		Name:  "dir3",
		Path:  "ShouldGetReset",
	}
	res := as.JSON("/containers").Post(cnt)
	assert.Equal(t, http.StatusCreated, res.Code, "It should be able to create")
	defer test_common.CleanupContainer(cnt)

	resObj := models.Container{}
	json.NewDecoder(res.Body).Decode(&resObj)
	assert.Equal(t, resObj.Name, cnt.Name)
	assert.NotZero(t, resObj.ID)
	assert.Equal(t, http.StatusCreated, res.Code)

	// Path does not come back from the API (hidden), check it updated.
	check := models.Container{}
	assert.NoError(t, as.DB.Find(&check, resObj.ID))
	assert.Equal(t, check.Path, cfg.Dir, "It should reset our path")
}

func Test_ContainersResource_Update(t *testing.T) {
	test_common.InitFakeApp(true)
	cnt := CreateContainer("Initial", as)
	test_common.CleanupContainer(&cnt)
	assert.NotZero(t, cnt.ID)

	name := "UpdateTest"
	cnt.Name = name
	test_common.CreateContainerPath(&cnt)

	url := fmt.Sprintf("/api/containers/%d", cnt.ID)
	res := as.JSON(url).Put(cnt)
	defer test_common.CleanupContainer(&cnt)
	assert.Equal(t, http.StatusOK, res.Code, fmt.Sprintf("Error %s", res.Body.String()))

	check := models.Container{}
	json.NewDecoder(res.Body).Decode(&check)
	assert.Equal(t, check.Name, name, "It should update the name")
}

func Test_ContainersResource_Destroy(t *testing.T) {
	test_common.InitFakeApp(true)
	cnt := CreateContainer("Nuke", as)
	defer test_common.CleanupContainer(&cnt)
	assert.NotZero(t, cnt.ID)

	url := fmt.Sprintf("/api/containers/%d", cnt.ID)
	res := as.JSON(url).Delete()
	assert.Equal(t, http.StatusOK, res.Code)

	notFoundRes := as.JSON(url).Get()
	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)
}

func Test_ContainerList(t *testing.T) {
	test_common.InitFakeApp(true)
	db := models.InitGorm(false)

	cnt1, _ := test_common.GetContentByDirName("dir1")
	cnt2, _ := test_common.GetContentByDirName("dir2")
	db.Create(cnt1)
	db.Create(cnt2)
	res := as.JSON("/containers").Get()
	assert.Equal(t, http.StatusOK, res.Code)

	containers := ContainersResponse{}
	json.NewDecoder(res.Body).Decode(&containers)

	assert.Equal(t, 2, len(containers.Results), "It should have loaded two fixtures")
	var found *models.Container
	for _, c := range containers.Results {
		if c.Name == "dir2" {
			found = &c
		}
	}
	as.NotNil(found, "If it had the fixture loaded we should have this name")
	url := fmt.Sprintf("/api/containers/%d/contents", found.ID)
	contentRes := as.JSON(url).Get()
	assert.Equal(t, http.StatusOK, contentRes.Code)
}

func Test_Memory_ReadOnlyDenyEdit(t *testing.T) {
	cfg := test_common.InitFakeApp(false)
	cfg.ReadOnly = true
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	containers, count, err := man.ListContainersContext()
	assert.NoError(t, err, "It should list containers")
	assert.Greater(t, count, 0, "The count should be positive")
	assert.Greater(t, len(*containers), 0, "There should be containers")

	for _, c := range *containers {
		c.Name = "Update Should fail"
		url := fmt.Sprintf("/api/containers/%d", c.ID)
		res := as.JSON(url).Put(&c)
		assert.Equal(t, http.StatusNotImplemented, res.Code)
	}
}
*/

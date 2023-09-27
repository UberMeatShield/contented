package actions

// These tests are DB based tests, vs in memory manager test_common.InitFakeApp(true)

import (
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func (as *ActionSuite) Test_ContainersResource_List() {
	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
}

func CreateContainer(name string, as *ActionSuite) models.Container {
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

func (as *ActionSuite) Test_ContainersResource_Show() {
	test_common.InitFakeApp(true)
	name := "ShowTest"
	cnt := CreateContainer(name, as)
	defer test_common.CleanupContainer(&cnt)
	as.NotZero(cnt.ID)

	validate := as.JSON("/containers/" + cnt.ID.String()).Get()
	as.Equal(http.StatusOK, validate.Code)

	resObj := models.Container{}
	json.NewDecoder(validate.Body).Decode(&resObj)
	as.Equal(resObj.Name, name)
}

func (as *ActionSuite) Test_ContainersResource_Create() {
	cfg := test_common.InitFakeApp(true)
	cnt := &models.Container{
		Total: 1,
		Name:  "dir3",
		Path:  "ShouldGetReset",
	}
	res := as.JSON("/containers").Post(cnt)
	as.Equal(http.StatusCreated, res.Code, "It should be able to create")
	defer test_common.CleanupContainer(cnt)

	resObj := models.Container{}
	json.NewDecoder(res.Body).Decode(&resObj)
	as.Equal(resObj.Name, cnt.Name)
	as.NotZero(resObj.ID)
	as.Equal(http.StatusCreated, res.Code)

	// Path does not come back from the API (hidden), check it updated.
	check := models.Container{}
	as.NoError(as.DB.Find(&check, resObj.ID))
	as.Equal(check.Path, cfg.Dir, "It should reset our path")
}

func (as *ActionSuite) Test_ContainersResource_Update() {
	test_common.InitFakeApp(true)
	cnt := CreateContainer("Initial", as)
	test_common.CleanupContainer(&cnt)
	as.NotZero(cnt.ID)

	name := "UpdateTest"
	cnt.Name = name
	test_common.CreateContainerPath(&cnt)

	res := as.JSON("/containers/" + cnt.ID.String()).Put(cnt)
	defer test_common.CleanupContainer(&cnt)
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Error %s", res.Body.String()))

	check := models.Container{}
	json.NewDecoder(res.Body).Decode(&check)
	as.Equal(check.Name, name, "It should update the name")
}

func (as *ActionSuite) Test_ContainersResource_Destroy() {
	test_common.InitFakeApp(true)
	cnt := CreateContainer("Nuke", as)
	defer test_common.CleanupContainer(&cnt)
	as.NotZero(cnt.ID)

	res := as.JSON("/containers/" + cnt.ID.String()).Delete()
	as.Equal(http.StatusOK, res.Code)

	notFoundRes := as.JSON("/containers/" + cnt.ID.String()).Get()
	as.Equal(http.StatusNotFound, notFoundRes.Code)
}

func (as *ActionSuite) Test_ContainerList() {
	test_common.InitFakeApp(true)

	cnt1, _ := test_common.GetContentByDirName("dir1")
	cnt2, _ := test_common.GetContentByDirName("dir2")
	models.DB.Create(cnt1)
	models.DB.Create(cnt2)
	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)

	containers := models.Containers{}
	json.NewDecoder(res.Body).Decode(&containers)

	as.Equal(2, len(containers), "It should have loaded two fixtures")
	var found *models.Container
	for _, c := range containers {
		if c.Name == "dir2" {
			found = &c
		}
	}
	as.NotNil(found, "If it had the fixture loaded we should have this name")

	contentRes := as.JSON("/containers/" + found.ID.String() + "/content").Get()
	as.Equal(http.StatusOK, contentRes.Code)
}

func (as *ActionSuite) Test_Memory_ReadOnlyDenyEdit() {
	cfg := test_common.InitFakeApp(false)
	cfg.ReadOnly = true
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	containers, err := man.ListContainersContext()
	as.NoError(err, "It should list containers")
	as.Greater(len(*containers), 0, "There should be containers")

	for _, c := range *containers {
		c.Name = "Update Should fail"
		res := as.JSON("/containers/" + c.ID.String()).Put(&c)
		as.Equal(http.StatusNotImplemented, res.Code)
	}
}

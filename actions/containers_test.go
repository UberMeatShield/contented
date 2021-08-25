package actions
// These tests are DB based tests, vs in memory manager internals.InitFakeApp(true)

import (
	"contented/models"
	"contented/internals"
	"contented/managers"
	"encoding/json"
	"net/http"
)

func (as *ActionSuite) Test_ContainersResource_List() {
	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
}

func CreateContainer(name string, as *ActionSuite) models.Container {
	c := &models.Container{
		Total: 1,
		Name:  name,
		Path:  "test/thing",
	}
	res := as.JSON("/containers").Post(c)

	resObj := models.Container{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj
}

func (as *ActionSuite) Test_ContainersResource_Show() {
    internals.InitFakeApp(true)
	name := "Show Test"
	s := CreateContainer(name, as)
	as.NotZero(s.ID)

	validate := as.JSON("/containers/" + s.ID.String()).Get()
	as.Equal(http.StatusOK, validate.Code)

	resObj := models.Container{}
	json.NewDecoder(validate.Body).Decode(&resObj)
	as.Equal(resObj.Name, name)
}

func (as *ActionSuite) Test_ContainersResource_Create() {
    internals.InitFakeApp(true)
	c := &models.Container{
		Total: 1,
		Name:  "Derp",
		Path:  "test/thing",
	}
	res := as.JSON("/containers").Post(c)
    as.Equal(http.StatusCreated, res.Code, "It should be able to create")

	resObj := models.Container{}
	json.NewDecoder(res.Body).Decode(&resObj)

	as.Equal(resObj.Name, c.Name)
	as.NotZero(resObj.ID)
	as.Equal(http.StatusCreated, res.Code)
}

func (as *ActionSuite) Test_ContainersResource_Update() {
    internals.InitFakeApp(true)
	s := CreateContainer("Initial Title", as)
	as.NotZero(s.ID)

	name := "Update test"
	s.Name = name
	res := as.JSON("/containers/" + s.ID.String()).Put(s)
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_ContainersResource_Destroy() {
    internals.InitFakeApp(true)
	s := CreateContainer("Initial Title", as)
	as.NotZero(s.ID)

	res := as.JSON("/containers/" + s.ID.String()).Delete()
	as.Equal(http.StatusOK, res.Code)

	notFoundRes := as.JSON("/containers/" + s.ID.String()).Get()
	as.Equal(http.StatusNotFound, notFoundRes.Code)
}

func (as *ActionSuite) Test_ContainerFixture() {
    internals.InitFakeApp(true)
    as.LoadFixture("base")

    res := as.JSON("/containers").Get()
    as.Equal(http.StatusOK, res.Code)

    containers := models.Containers{}
	json.NewDecoder(res.Body).Decode(&containers)

    as.Equal(2, len(containers), "It should have loaded two fixtures")

    var found *models.Container
    for _, c := range containers {
        if c.Name == "contain1" {
            found = &c
        }
    }
    as.NotNil(found, "If it had the fixture loaded we should have this name")

    mediaRes := as.JSON("/containers/" + found.ID.String() + "/media").Get()
    as.Equal(http.StatusOK, mediaRes.Code)
}

func (as *ActionSuite) Test_MemoryDenyEdit() {
    cfg := internals.InitFakeApp(false)
    cfg.UseDatabase = false
    ctx := internals.GetContext(as.App)
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

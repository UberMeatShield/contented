package actions
// These tests are DB based tests, vs in memory manager init_fake_app(true)

import (
	"contented/models"
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
    init_fake_app(true)
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
    init_fake_app(true)
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
    init_fake_app(true)
	s := CreateContainer("Initial Title", as)
	as.NotZero(s.ID)

	name := "Update test"
	s.Name = name
	res := as.JSON("/containers/" + s.ID.String()).Put(s)
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_ContainersResource_Destroy() {
    init_fake_app(true)
	s := CreateContainer("Initial Title", as)
	as.NotZero(s.ID)

	res := as.JSON("/containers/" + s.ID.String()).Delete()
	as.Equal(http.StatusOK, res.Code)

	notFoundRes := as.JSON("/containers/" + s.ID.String()).Get()
	as.Equal(http.StatusNotFound, notFoundRes.Code)
}

func (as *ActionSuite) Test_ContainerFixture() {
    init_fake_app(true)
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

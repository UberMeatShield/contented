package actions

import (
  "encoding/json"
  "net/http"
  "contented/models"
)

func (as *ActionSuite) Test_ContainersResource_List() {
    res := as.JSON("/containers").Get()
    as.Equal(http.StatusOK, res.Code)
}

func CreateContainer(name string, as *ActionSuite) models.Container {
    c := &models.Container{
        Total: 1,
        Name: name,
        Path: "test/thing",
    }
    res := as.JSON("/containers").Post(c)

    resObj := models.Container{}
    json.NewDecoder(res.Body).Decode(&resObj)
    return resObj
}

func (as *ActionSuite) Test_ContainersResource_Show() {
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
    c := &models.Container{
        Total: 1,
        Name: "Derp",
        Path: "test/thing",
    }
    res := as.JSON("/containers").Post(c)

    resObj := models.Container{}
    json.NewDecoder(res.Body).Decode(&resObj)

    as.Equal(resObj.Name, c.Name)
    as.NotZero(resObj.ID)
    as.Equal(http.StatusCreated, res.Code)
}

func (as *ActionSuite) Test_ContainersResource_Update() {
    s := CreateContainer("Initial Title", as)
    as.NotZero(s.ID)

    name := "Update test"
    s.Name = name
    res := as.JSON("/containers/" + s.ID.String()).Put(s)
    as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_ContainersResource_Destroy() {
    s := CreateContainer("Initial Title", as)
    as.NotZero(s.ID)

    res := as.JSON("/containers/" + s.ID.String()).Delete()
    as.Equal(http.StatusOK, res.Code)
}


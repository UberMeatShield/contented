package actions

import (
    "github.com/gobuffalo/nulls"
	"contented/models"
	"encoding/json"
	"net/http"
)

func CreateResource(src string, container_id nulls.NewUUID, as *ActionSuite) models.MediaContainer {
	mc := &models.MediaContainer{
		Src:     src,
		Type:    "test",
		Preview: "",
        ContainerID: container_id,
	}
	res := as.JSON("/media").Post(mc)
	as.Equal(http.StatusCreated, res.Code)

	resObj := models.MediaContainer{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj
}


func (as *ActionSuite) Test_MediaSubQuery() {
    // Create 2 containers
    c1 := &models.Container{
         Total: 2,
         Path:  "container/1/media",
         Name:  "Trash1",
    }
    c2 := &models.Container{
         Total: 2,
         Path:  "container/2/media",
         Name:  "Trash2",
    }
    as.DB.Create(&c1)
    as.DB.Create(&c2)
    as.NotZero(c1.ID)
    as.NotZero(c2.ID)
    // Add resources to both
    // Filter based on container
}

func (as *ActionSuite) Test_MediaContainersResource_List() {
	res := as.JSON("/media").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Show() {
	src := "test_query"
	mc := CreateResource(src, nil, as)
	check := as.JSON("/media/" + mc.ID.String()).Get()
	as.Equal(http.StatusOK, check.Code)

	validate := models.MediaContainer{}
	json.NewDecoder(check.Body).Decode(&validate)
	as.Equal(src, validate.Src)
}

func (as *ActionSuite) Test_MediaContainersResource_Create() {
	mc := CreateResource("test_create", nil, as)
	as.NotZero(mc.ID)
}

func (as *ActionSuite) Test_MediaContainersResource_Update() {
	mc := CreateResource("test_update", nil, as)
	mc.Type = "Update Test"
	up_res := as.JSON("/media/" + mc.ID.String()).Put(mc)
	as.Equal(http.StatusOK, up_res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Destroy() {
	mc := CreateResource("Nuke Test", nil, as)
	del_res := as.JSON("/media/" + mc.ID.String()).Delete()
	as.Equal(http.StatusOK, del_res.Code)
}

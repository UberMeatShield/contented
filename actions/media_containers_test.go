package actions

import (
	"contented/models"
	"encoding/json"
	"net/http"
)

func CreateResource(src string, as *ActionSuite) models.MediaContainer {
	mc := &models.MediaContainer{
		Src:     src,
		Type:    "test",
		Preview: "",
	}
	res := as.JSON("/media").Post(mc)
	as.Equal(http.StatusCreated, res.Code)

	resObj := models.MediaContainer{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj
}

func (as *ActionSuite) Test_MediaContainersResource_List() {
	res := as.JSON("/media").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Show() {
	src := "test_query"
	mc := CreateResource(src, as)
	check := as.JSON("/media/" + mc.ID.String()).Get()
	as.Equal(http.StatusOK, check.Code)

	validate := models.MediaContainer{}
	json.NewDecoder(check.Body).Decode(&validate)
	as.Equal(src, validate.Src)
}

func (as *ActionSuite) Test_MediaContainersResource_Create() {
	mc := CreateResource("test_create", as)
	as.NotZero(mc.ID)
}

func (as *ActionSuite) Test_MediaContainersResource_Update() {
	mc := CreateResource("test_update", as)
	mc.Type = "Update Test"
	up_res := as.JSON("/media/" + mc.ID.String()).Put(mc)
	as.Equal(http.StatusOK, up_res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Destroy() {
	mc := CreateResource("Nuke Test", as)
	del_res := as.JSON("/media/" + mc.ID.String()).Delete()
	as.Equal(http.StatusOK, del_res.Code)
}

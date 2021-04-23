package actions

import (
    "github.com/gobuffalo/nulls"
	"contented/models"
	"encoding/json"
	"net/http"
)

func CreateResource(src string, container_id nulls.UUID, as *ActionSuite) models.MediaContainer {
    init_fake_app(true)
	mc := &models.MediaContainer{
		Src:     src,
		ContentType:    "test",
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
    init_fake_app(true)
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
    as.DB.Create(c1)
    as.DB.Create(c2)
    as.NotZero(c1.ID)
    as.NotZero(c2.ID)

    CreateResource("a", nulls.NewUUID(c1.ID), as)
    CreateResource("b", nulls.NewUUID(c1.ID), as)
    CreateResource("c", nulls.NewUUID(c2.ID), as)
    CreateResource("d", nulls.NewUUID(c2.ID), as)
    CreateResource("e", nulls.NewUUID(c2.ID), as)

    res1 := as.JSON("/containers/" +  c1.ID.String() +  "/media").Get()
    res2 := as.JSON("/containers/" +  c2.ID.String() +  "/media").Get()

    as.Equal(http.StatusOK, res1.Code)
    as.Equal(http.StatusOK, res2.Code)
    // Add resources to both
    // Filter based on container
	validate1 := models.MediaContainers{}
	validate2 := models.MediaContainers{}
	json.NewDecoder(res1.Body).Decode(&validate1)
	json.NewDecoder(res2.Body).Decode(&validate2)

    as.Equal(len(validate1), 2, "There should be 2 media containers found")
    as.Equal(len(validate2), 3, "There should be 3 in this one")
}

func (as *ActionSuite) Test_MediaContainersResource_List() {
    init_fake_app(true)
	res := as.JSON("/media").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Show() {
    init_fake_app(true)
	src := "test_query"
	mc := CreateResource(src, nulls.UUID{}, as)
	check := as.JSON("/media/" + mc.ID.String()).Get()
	as.Equal(http.StatusOK, check.Code)

	validate := models.MediaContainer{}
	json.NewDecoder(check.Body).Decode(&validate)
	as.Equal(src, validate.Src)
}

func (as *ActionSuite) Test_MediaContainersResource_Create() {
	mc := CreateResource("test_create", nulls.UUID{}, as)
	as.NotZero(mc.ID)
}

func (as *ActionSuite) Test_MediaContainersResource_Update() {
	mc := CreateResource("test_update", nulls.UUID{}, as)
	mc.ContentType = "Update Test"
	up_res := as.JSON("/media/" + mc.ID.String()).Put(mc)
	as.Equal(http.StatusOK, up_res.Code)
}

func (as *ActionSuite) Test_MediaContainersResource_Destroy() {
    init_fake_app(true)
	mc := CreateResource("Nuke Test", nulls.UUID{}, as)
	del_res := as.JSON("/media/" + mc.ID.String()).Delete()
	as.Equal(http.StatusOK, del_res.Code)
}

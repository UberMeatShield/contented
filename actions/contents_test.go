package actions

import (
    "contented/internals"
    "contented/models"
    "encoding/json"
    "github.com/gobuffalo/nulls"
    "net/http"
    "net/url"
)

func CreateResource(src string, container_id nulls.UUID, as *ActionSuite) models.Content {
    internals.InitFakeApp(true)
    mc := &models.Content{
        Src:         src,
        ContentType: "test",
        Preview:     "",
        ContainerID: container_id,
    }
    res := as.JSON("/content").Post(mc)
    as.Equal(http.StatusCreated, res.Code)

    resObj := models.Content{}
    json.NewDecoder(res.Body).Decode(&resObj)
    return resObj
}

func (as *ActionSuite) Test_ContentSubQuery() {
    // Create 2 containers
    internals.InitFakeApp(true)
    c1 := &models.Container{
        Total: 2,
        Path:  "container/1/content",
        Name:  "Trash1",
    }
    c2 := &models.Container{
        Total: 2,
        Path:  "container/2/content",
        Name:  "Trash2",
    }
    as.DB.Create(c1)
    as.DB.Create(c2)
    as.NotZero(c1.ID)
    as.NotZero(c2.ID)

    CreateResource("a", nulls.NewUUID(c1.ID), as)
    CreateResource("b", nulls.NewUUID(c1.ID), as)
    CreateResource("c", nulls.NewUUID(c2.ID), as)
    CreateResource("donut", nulls.NewUUID(c2.ID), as)
    CreateResource("e", nulls.NewUUID(c2.ID), as)

    res1 := as.JSON("/containers/" + c1.ID.String() + "/content").Get()
    res2 := as.JSON("/containers/" + c2.ID.String() + "/content").Get()

    as.Equal(http.StatusOK, res1.Code)
    as.Equal(http.StatusOK, res2.Code)
    // Add resources to both
    // Filter based on container
    validate1 := models.Contents{}
    validate2 := models.Contents{}
    json.NewDecoder(res1.Body).Decode(&validate1)
    json.NewDecoder(res2.Body).Decode(&validate2)

    as.Equal(len(validate1), 2, "There should be 2 content containers found")
    as.Equal(len(validate2), 3, "There should be 3 in this one")

    // Add in a test that uses the search interface via the actions via DB
    params := url.Values{}
    params.Add("text", "donut")

    res3 := as.JSON("/search?%s", params.Encode()).Get()
    as.Equal(http.StatusOK, res3.Code)
    validate3 := SearchResult{}
    json.NewDecoder(res3.Body).Decode(&validate3)
    as.Equal(1, len(*validate3.Content), "We have one donut")
}

func (as *ActionSuite) Test_ManagerDB_Preview() {
    models.DB.TruncateAll()
    internals.ResetConfig()
    internals.InitFakeApp(true)

    cnt, content := internals.GetContentByDirName("dir2")
    as.Equal(3, len(content), "Dir2 should have 3 items")
    as.Equal("dir2", cnt.Name, "It should have loaded the right item")

    as.DB.Create(cnt)
    as.NotZero(cnt.ID, "We should have an ID now for the container")
    for _, mc := range content {
        mc.ContainerID = nulls.NewUUID(cnt.ID)
        as.DB.Create(&mc)
        as.NotZero(mc.ID, "It should have a content container ID and id")
        previewRes := as.JSON("/preview/%s", mc.ID).Get()
        as.Equal(http.StatusOK, previewRes.Code)
    }
}

func (as *ActionSuite) Test_MemoryAPIBasics() {
    internals.InitFakeApp(false)
    res := as.JSON("/content").Get()
    as.Equal(http.StatusOK, res.Code)

    validate := models.Contents{}
    json.NewDecoder(res.Body).Decode(&validate)
    as.Equal(internals.TOTAL_MEDIA, len(validate), "It should have a known set of mock data")

    validate_search := models.Contents{}
    res_search := as.JSON("/search?text=Large").Get()
    json.NewDecoder(res_search.Body).Decode(&validate_search)
    as.Equal(internals.TOTAL_MEDIA, len(validate), "In memory should have these")
}

func (as *ActionSuite) Test_ContentsResource_List() {
    internals.InitFakeApp(true)
    src := "test_list"
    CreateResource(src, nulls.UUID{}, as)
    res := as.JSON("/content").Get()
    as.Equal(http.StatusOK, res.Code)

    validate := models.Contents{}
    json.NewDecoder(res.Body).Decode(&validate)
    as.Equal(src, validate[0].Src)
    as.Equal(1, len(validate), "One item should be in the DB")
}

func (as *ActionSuite) Test_ContentsResource_Show() {
    internals.InitFakeApp(true)
    src := "test_query"
    mc := CreateResource(src, nulls.UUID{}, as)
    check := as.JSON("/content/" + mc.ID.String()).Get()
    as.Equal(http.StatusOK, check.Code)

    validate := models.Content{}
    json.NewDecoder(check.Body).Decode(&validate)
    as.Equal(src, validate.Src)
}

func (as *ActionSuite) Test_ContentsResource_Create() {
    mc := CreateResource("test_create", nulls.UUID{}, as)
    as.NotZero(mc.ID)
}

func (as *ActionSuite) Test_ContentsResource_Update() {
    mc := CreateResource("test_update", nulls.UUID{}, as)
    mc.ContentType = "Update Test"
    up_res := as.JSON("/content/" + mc.ID.String()).Put(mc)
    as.Equal(http.StatusOK, up_res.Code)
}

func (as *ActionSuite) Test_ContentsResource_Destroy() {
    internals.InitFakeApp(true)
    mc := CreateResource("Nuke Test", nulls.UUID{}, as)
    del_res := as.JSON("/content/" + mc.ID.String()).Delete()
    as.Equal(http.StatusOK, del_res.Code)
}
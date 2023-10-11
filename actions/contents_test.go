package actions

import (
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gobuffalo/nulls"
)

func CreateResource(src string, container_id nulls.UUID, as *ActionSuite) models.Content {
	mc := &models.Content{
		Src:         src,
		ContentType: "test",
		Preview:     "",
		ContainerID: container_id,
		NoFile:      true,
	}
	res := as.JSON("/contents").Post(mc)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Error creating %s", res.Body.String()))

	resObj := models.Content{}
	json.NewDecoder(res.Body).Decode(&resObj)
	return resObj
}

func (as *ActionSuite) Test_ContentSubQuery_DB() {
	// Create 2 containers
	test_common.InitFakeApp(true)
	c1 := &models.Container{
		Total: 2,
		Path:  "container/1/contents",
		Name:  "Trash1",
	}
	c2 := &models.Container{
		Total: 2,
		Path:  "container/2/contents",
		Name:  "Trash2",
	}
	c3 := &models.Container{
		Total:  1,
		Path:   "/container/3/contents",
		Name:   "Hidden",
		Hidden: true,
	}
	as.DB.Create(c1)
	as.DB.Create(c2)
	as.DB.Create(c3)
	as.NotZero(c1.ID)
	as.NotZero(c2.ID)
	as.NotZero(c3.ID)

	CreateResource("a", nulls.NewUUID(c1.ID), as)
	CreateResource("b", nulls.NewUUID(c1.ID), as)
	CreateResource("c", nulls.NewUUID(c2.ID), as)
	CreateResource("donut", nulls.NewUUID(c2.ID), as)
	CreateResource("e", nulls.NewUUID(c2.ID), as)

	// Check that hidden resources stay hidden
	up := CreateResource("donut_2_search_should_fail", nulls.NewUUID(c3.ID), as)
	up.Hidden = true
	upErr := as.DB.Update(&up)
	as.NoError(upErr, fmt.Sprintf("It should have updated %s", upErr))

	res1 := as.JSON("/containers/" + c1.ID.String() + "/contents").Get()
	res2 := as.JSON("/containers/" + c2.ID.String() + "/contents").Get()

	as.Equal(http.StatusOK, res1.Code)
	as.Equal(http.StatusOK, res2.Code)
	// Add resources to both
	// Filter based on container
	validate1 := ContentsResponse{}
	validate2 := ContentsResponse{}
	json.NewDecoder(res1.Body).Decode(&validate1)
	json.NewDecoder(res2.Body).Decode(&validate2)

	as.Equal(2, len(validate1.Contents), "There should be 2 content containers found")
	as.Equal(3, len(validate2.Contents), "There should be 3 in this one")

	// Add in a test that uses the search interface via the actions via DB
	params := url.Values{}
	params.Add("text", "donut")
	res3 := as.JSON("/search?%s", params.Encode()).Get()
	as.Equal(http.StatusOK, res3.Code)
	validate3 := SearchResult{}
	json.NewDecoder(res3.Body).Decode(&validate3)
	as.Equal(1, len(*validate3.Content), "We have one donut that is not hidden")
}

func (as *ActionSuite) Test_ManagerDB_Preview() {
	models.DB.TruncateAll()
	test_common.ResetConfig()
	test_common.InitFakeApp(true)

	cnt, content := test_common.GetContentByDirName("dir2")

	as.Equal("dir2", cnt.Name, "It should have loaded the right item")
	as.Equal(test_common.EXPECT_CNT_COUNT["dir2"], len(content), fmt.Sprintf("Dir2 should have 3 items %s", content))

	as.DB.Create(cnt)
	as.NotZero(cnt.ID, "We should have an ID now for the container")
	for _, mc := range content {
		mc.ContainerID = nulls.NewUUID(cnt.ID)
		err := as.DB.Create(&mc)
		as.NoError(err, fmt.Sprintf("It should create item %s with err %s", mc.Src, err))
		as.NotZero(mc.ID, "It should have a content container ID and id")
		previewRes := as.JSON("/preview/%s", mc.ID).Get()
		as.Equal(http.StatusOK, previewRes.Code, fmt.Sprintf("Failed to find preview for %s preview (%s)", mc.Src, previewRes.Response.Body))
	}
}

func (as *ActionSuite) Test_MemoryAPIBasics() {
	test_common.ResetConfig()
	test_common.InitFakeApp(false)
	res := as.JSON("/contents").Get()
	as.Equal(http.StatusOK, res.Code)

	// Also validates that hidden content doesn't come back from the main listing API
	validate := ContentsResponse{}
	json.NewDecoder(res.Body).Decode(&validate)
	as.Equal(test_common.TOTAL_MEDIA, len(validate.Contents), "It should have a known set of mock data")

	// I feel like this should be failing?
	res_search := as.JSON("/search/?text=Large").Get()
	as.Equal(res_search.Code, http.StatusOK, "It should search")

	validate_search := SearchResult{}
	json.NewDecoder(res_search.Body).Decode(&validate_search)
	as.Equal(5, len(*validate_search.Content), fmt.Sprintf("In memory should have these %s", res_search.Body.String()))
}

func (as *ActionSuite) Test_ContentsResource_List() {
	test_common.InitFakeApp(true)
	src := "test_list"
	CreateResource(src, nulls.UUID{}, as)
	res := as.JSON("/contents").Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Failed %s", res.Body.String()))

	validate := ContentsResponse{}
	json.NewDecoder(res.Body).Decode(&validate)
	as.Equal(src, validate.Contents[0].Src)
	as.Equal(1, len(validate.Contents), "One item should be in the DB")
}

func (as *ActionSuite) Test_ContentsResource_Show() {
	test_common.InitFakeApp(true)
	src := "test_query"
	mc := CreateResource(src, nulls.UUID{}, as)
	check := as.JSON("/contents/" + mc.ID.String()).Get()
	as.Equal(http.StatusOK, check.Code)

	validate := models.Content{}
	json.NewDecoder(check.Body).Decode(&validate)
	as.Equal(src, validate.Src)
}

func (as *ActionSuite) Test_ContentsResource_Create() {
	test_common.InitFakeApp(true)
	mc := CreateResource("test_create", nulls.UUID{}, as)
	as.NotZero(mc.ID)
}

func (as *ActionSuite) Test_ContentsResource_Update_DB() {
	test_common.InitFakeApp(true)
	mc := CreateResource("test_update", nulls.UUID{}, as)

	tag := models.Tag{ID: "TAG"}
	invalid := models.Tag{ID: "NOT IN DAB"}
	as.NoError(as.DB.Create(&tag))
	mc.Tags = models.Tags{tag, invalid}

	mc.ContentType = "Update Test Memory"
	up_res := as.JSON("/contents/" + mc.ID.String()).Put(mc)
	as.Equal(http.StatusOK, up_res.Code, fmt.Sprintf("Err %s", up_res.Body.String()))

	validate := models.Content{}
	json.NewDecoder(up_res.Body).Decode(&validate)
	as.Equal(validate.ContentType, "Update Test Memory")
	tags := validate.Tags
	as.NotNil(tags)
	as.Equal(len(tags), 1, "There should be 1 tag actually in the DB")
}

func (as *ActionSuite) Test_ContentsResource_Update_Memory() {
	test_common.InitFakeApp(false)
	mc := CreateResource("test_update", nulls.UUID{}, as)
	mc.ContentType = "Update Test Memory"
	up_res := as.JSON("/contents/" + mc.ID.String()).Put(mc)
	as.Equal(http.StatusOK, up_res.Code, fmt.Sprintf("Err %s", up_res.Body.String()))
}

func (as *ActionSuite) Test_ContentsResource_Destroy() {
	test_common.InitFakeApp(true)
	mc := CreateResource("Nuke Test", nulls.UUID{}, as)
	del_res := as.JSON("/contents/" + mc.ID.String()).Delete()
	as.Equal(http.StatusOK, del_res.Code)
}

// Also a good bit of testing the creation logic.
func (as *ActionSuite) Test_ActionsMemory_TagSearch() {
	cfg := test_common.InitMemoryFakeAppEmpty()
	as.Equal(cfg.ReadOnly, false)
	ActionsTagSearchValidation(as)
}

func (as *ActionSuite) Test_ActionsDb_TagSearch() {
	cfg := test_common.InitFakeApp(true)
	as.Equal(cfg.ReadOnly, false)
	ActionsTagSearchValidation(as)
}

func ActionsTagSearchValidation(as *ActionSuite) {
	cnt := models.Container{Name: "Tagged"}
	test_common.CreateContainerPath(&cnt)
	defer test_common.CleanupContainer(&cnt)

	cRes := as.JSON("/containers/").Post(&cnt)
	as.Equal(http.StatusCreated, cRes.Code, fmt.Sprintf("Failed to create container %s", cRes.Body.String()))
	cntCheck := models.Container{}
	json.NewDecoder(cRes.Body).Decode(&cntCheck)

	t := models.Tag{ID: "Zug"}
	tRes := as.JSON("/tags/").Post(&t)
	as.Equal(http.StatusCreated, tRes.Code, fmt.Sprintf("Tags Failed %s", tRes.Body.String()))

	a := models.Content{ContainerID: nulls.NewUUID(cntCheck.ID), Src: "AFile"}
	b := models.Content{ContainerID: nulls.NewUUID(cntCheck.ID), Src: "BFile"}
	b.Tags = models.Tags{t}

	aRes := as.JSON("/contents/").Post(&a)
	as.Equal(http.StatusCreated, aRes.Code)
	bRes := as.JSON("/contents/").Post(&b)
	as.Equal(http.StatusCreated, bRes.Code)

	params := url.Values{}
	params.Add("tags", "Zug")
	res := as.JSON("/search?%s", params.Encode()).Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Search failed %s", res.Body.String()))

	validate := SearchResult{}
	json.NewDecoder(res.Body).Decode(&validate)
	as.Equal(1, len(*validate.Content), fmt.Sprintf("Searching tags return content"))
}

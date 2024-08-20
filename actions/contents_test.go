package actions

import (
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateResource(src string, containerID *int64, t *testing.T, router *gin.Engine) models.Content {
	mc := &models.Content{
		Src:         src,
		ContentType: "test",
		Preview:     "",
		ContainerID: containerID,
		NoFile:      true,
	}

	resObj := models.Content{}
	code, err := PostJson("/api/contents", mc, &resObj, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Error creating %s", err))
	return resObj
}

/*
func (as *ActionSuite) Test_ContentSubQuery_DB() {
	// Create 2 containers
	InitFakeRouterApp(true)
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
	assert.NoError(t, upErr, fmt.Sprintf("It should have updated %s", upErr))

	res1 := as.JSON("/containers/" + c1.ID.String() + "/contents").Get()
	res2 := as.JSON("/containers/" + c2.ID.String() + "/contents").Get()

	assert.Equal(t, http.StatusOK, res1.Code)
	assert.Equal(t, http.StatusOK, res2.Code)
	// Add resources to both
	// Filter based on container
	validate1 := ContentsResponse{}
	validate2 := ContentsResponse{}
	json.NewDecoder(res1.Body).Decode(&validate1)
	json.NewDecoder(res2.Body).Decode(&validate2)

	assert.Equal(t, 2, len(validate1.Results), "There should be 2 content containers found")
	assert.Equal(t, 3, len(validate2.Results), "There should be 3 in this one")

	// Add in a test that uses the search interface via the actions via DB
	params := url.Values{}
	params.Add("search", "donut")
	res3 := as.JSON("/api/search/contents?%s", params.Encode()).Get()
	assert.Equal(t, http.StatusOK, res3.Code)
	validate3 := SearchContentsResult{}
	json.NewDecoder(res3.Body).Decode(&validate3)
	assert.Equal(t, 1, len(*validate3.Results), "We have one donut that is not hidden")
}

func (as *ActionSuite) Test_ManagerDB_Preview() {
	models.DB.TruncateAll()
	test_common.ResetConfig()
	InitFakeRouterApp(true)

	cnt, content := test_common.GetContentByDirName("dir2")

	assert.Equal(t, "dir2", cnt.Name, "It should have loaded the right item")
	assert.Equal(t, test_common.EXPECT_CNT_COUNT["dir2"], len(content), fmt.Sprintf("Dir2 should have 3 items %s", content))

	as.DB.Create(cnt)
	as.NotZero(cnt.ID, "We should have an ID now for the container")
	for _, mc := range content {
		mc.ContainerID = nulls.NewUUID(cnt.ID)
		err := as.DB.Create(&mc)
		assert.NoError(t, err, fmt.Sprintf("It should create item %s with err %s", mc.Src, err))
		as.NotZero(mc.ID, "It should have a content container ID and id")
		previewRes := as.JSON("/preview/%s", mc.ID).Get()
		assert.Equal(t, http.StatusOK, previewRes.Code, fmt.Sprintf("Failed to find preview for %s preview (%s)", mc.Src, previewRes.Response.Body))
	}
}
*/

// Would be better to have these call the same test code after an init to ensure they are the same
func TestMemoryAPIBasics(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)

	validate := ContentsResponse{}
	code, err := GetJson("/api/contents", nil, &validate, router)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, err, "It should make the call")

	// Also validates that hidden content doesn't come back from the main listing API
	assert.Equal(t, test_common.TOTAL_MEDIA, len(validate.Results), "It should have a known set of mock data")

	// I feel like this should be failing?
	validateSearch := SearchContentsResult{}
	searchCode, searchErr := GetJson("/api/search/contents?search=Large", "", &validateSearch, router)
	assert.NoError(t, searchErr, "It shoould search")

	assert.Equal(t, http.StatusOK, searchCode, "It should search")
	assert.Equal(t, 5, len(*validateSearch.Results), fmt.Sprintf("In memory should have these %s", validateSearch))
}

func TestContentsResourceListDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	src := "test_list"
	CreateResource(src, nil, t, router)
	validate := ContentsResponse{}
	code, err := GetJson("/api/contents", "", &validate, router)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, err)

	assert.Equal(t, src, validate.Results[0].Src)
	assert.Equal(t, 1, len(validate.Results), "One item should be in the DB")
}

func TestContentsResourceShow(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	src := "test_query"
	content := CreateResource(src, nil, t, router)

	url := fmt.Sprintf("/api/contents/%d", content.ID)
	validate := models.Content{}
	code, err := GetJson(url, "", &validate, router)
	assert.NoError(t, err, "Failed to get valid content")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, src, validate.Src)
}

func TestContentsResourceCreateDb(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	mc := CreateResource("test_create", nil, t, router)
	assert.Greater(t, mc.ID, int64(0))
}

func TestContentsResourceCreateMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	mc := CreateResource("test_create", nil, t, router)
	assert.Greater(t, mc.ID, int64(0))
}

func TestContentsResourceUpdateDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateUpdateContent(t, router)
}

func TestContentsResourceUpdateMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateUpdateContent(t, router)
}

func ValidateUpdateContent(t *testing.T, router *gin.Engine) {
	mc := CreateResource("test_update", nil, t, router)
	tag := CreateTag("TAG", t, router)

	invalid := models.Tag{ID: "Notinthesystem"}
	mc.Tags = models.Tags{tag, invalid}

	mc.ContentType = "Update Test Memory"
	url := fmt.Sprintf("/api/contents/%d", mc.ID)
	validate := models.Content{}
	code, err := PutJson(url, mc, &validate, router)

	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Err %s", err))
	assert.Equal(t, validate.ContentType, "Update Test Memory")

	tags := validate.Tags
	assert.NotNil(t, tags, "There should be tags associated now")
	assert.Equal(t, 1, len(tags), "There should be 1 tag actually in the DB")
}

func TestContentsResourceDestroyDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	ValidateDestroyContent(t, router)
}

func TestContentsResourceDestroyMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateDestroyContent(t, router)
}

func ValidateDestroyContent(t *testing.T, router *gin.Engine) {
	mc := CreateResource("NukeTest", nil, t, router)
	url := fmt.Sprintf("/api/contents/%d", mc.ID)
	code, err := DeleteJson(url, router)
	assert.Equal(t, http.StatusOK, code, "It should delete")
	assert.NoError(t, err)

	dead := models.Content{}
	codeDead, errDead := GetJson(url, "", &dead, router)
	assert.Equal(t, http.StatusNotFound, codeDead)
	assert.Error(t, errDead, "It should be not found")
}

// Also a good bit of testing the creation logic.
func TestActionsMemoryTagSearch(t *testing.T) {
	cfg := test_common.InitMemoryFakeAppEmpty()
	router := setupRouter()
	assert.Equal(t, cfg.ReadOnly, false)
	ActionsTagSearchValidation(t, router)
}

func TestActionsDbTagSearch(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(true)
	assert.Equal(t, cfg.ReadOnly, false)
	ActionsTagSearchValidation(t, router)
}

func ActionsTagSearchValidation(t *testing.T, router *gin.Engine) {
	cnt := models.Container{Name: "Tagged"}
	test_common.CreateContainerPath(&cnt)
	defer test_common.CleanupContainer(&cnt)

	cntCheck := models.Container{}
	code, cErr := PostJson("/api/containers", cnt, &cntCheck, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Failed to create container %s", cntCheck))
	assert.NoError(t, cErr)

	tagCheck := models.Tag{}
	tag := models.Tag{ID: "Zug"}
	tCode, tErr := PostJson("/api/tags", tag, &tagCheck, router)
	assert.Equal(t, http.StatusCreated, tCode, fmt.Sprintf("Tags Failed %s", tErr))

	a := models.Content{ContainerID: &cntCheck.ID, Src: "AFile"}
	b := models.Content{ContainerID: &cntCheck.ID, Src: "BFile"}
	b.Tags = models.Tags{tagCheck}

	aRes := models.Content{}
	aCode, aErr := PostJson("/api/contents", a, &aRes, router)
	assert.Equal(t, http.StatusCreated, aCode, fmt.Sprintf("Error creating A %s", aErr))
	bRes := models.Content{}
	bCode, bErr := PostJson("/api/contents", b, &bRes, router)
	assert.Equal(t, http.StatusCreated, bCode, fmt.Sprintf("Failed to create B %s", bErr))

	params := url.Values{}

	tagParam, _ := json.Marshal([]string{"Zug"})
	params.Add("tags", string(tagParam))

	url := fmt.Sprintf("/api/search/contents?%s", params.Encode())
	validate := SearchContentsResult{}
	vCode, vErr := GetJson(url, "", &validate, router)
	assert.Equal(t, http.StatusOK, vCode, fmt.Sprintf("Search failed %s", vErr))
	assert.Equal(t, 1, len(*validate.Results), "Searching tags return content")
}

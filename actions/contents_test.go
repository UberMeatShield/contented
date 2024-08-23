package actions

import (
	"bytes"
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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

func ValidateContentPreview(contentID int64, router *gin.Engine) (int, *httptest.ResponseRecorder, error) {
	url := fmt.Sprintf("/api/preview/%d", contentID)
	w := httptest.NewRecorder()
	userJson, _ := json.Marshal("")
	req, _ := http.NewRequest("GET", url, bytes.NewReader(userJson))
	router.ServeHTTP(w, req)

	result := w.Result()
	if !(result.StatusCode == 200 || result.StatusCode == 201) {
		log.Printf("Likely error getting data back from the server %s", w.Body)
		return result.StatusCode, w, fmt.Errorf("%s, %s", url, w.Body)
	}
	defer req.Body.Close()
	return result.StatusCode, w, nil
}

func TestContentSubQueryDB(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)

	c1 := CreateContainer("Trash1", t, router)
	c2 := CreateContainer("Trash2", t, router)
	c3 := CreateContainer("Trash3", t, router)
	c3.Hidden = true
	uCode, uErr := PutJson(fmt.Sprintf("/api/containers/%d", c3.ID), c3, &models.Container{}, router)
	assert.Equal(t, http.StatusOK, uCode, fmt.Sprintf("Failed to update %s", uErr))

	defer test_common.CleanupContainer(&c1)
	defer test_common.CleanupContainer(&c2)
	defer test_common.CleanupContainer(&c3)

	CreateResource("a", &c1.ID, t, router)
	CreateResource("b", &c1.ID, t, router)
	CreateResource("c", &c2.ID, t, router)
	CreateResource("donut", &c2.ID, t, router)
	CreateResource("e", &c2.ID, t, router)

	// TODO: Update this to work with the memory setup too (no good way to update the Hidden)
	hiddenDonut := CreateResource("donut_2_search_should_fail", &c3.ID, t, router)
	hiddenDonut.Hidden = true
	upRes := db.Save(hiddenDonut)
	assert.NoError(t, upRes.Error, "We should be able to update the hidden param")

	validate1 := ContentsResponse{}
	validate2 := ContentsResponse{}
	code1, err1 := GetJson(fmt.Sprintf("/api/containers/%d/contents", c1.ID), "", &validate1, router)
	code2, err2 := GetJson(fmt.Sprintf("/api/containers/%d/contents", c2.ID), "", &validate2, router)

	assert.Equal(t, http.StatusOK, code1, fmt.Sprintf("Failed to load container %s", err1))
	assert.Equal(t, http.StatusOK, code2, fmt.Sprintf("Failed to load container %s", err2))

	assert.Equal(t, 2, len(validate1.Results), "There should be 2 content containers found")
	assert.Equal(t, 3, len(validate2.Results), "There should be 3 in this one")

	// Test that search also respets hidden content
	params := url.Values{}
	params.Add("search", "donut")

	searchUrl := fmt.Sprintf("/api/search/contents?%s", params.Encode())
	fmt.Printf("What did we search %s", searchUrl)
	validate3 := SearchContentsResult{}
	searchCode, searchErr := GetJson(searchUrl, "", &validate3, router)

	assert.Equal(t, http.StatusOK, searchCode, fmt.Sprintf("Failed to search %s", searchErr))
	assert.Equal(t, 1, len(*validate3.Results), fmt.Sprintf("We have one donut that is not hidden %s", validate3))
}

func TestManagerPreviewDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)

	container, contents := test_common.GetContentByDirName("dir2")
	assert.Equal(t, "dir2", container.Name, "It should have loaded the right item")
	assert.Equal(t, test_common.EXPECT_CNT_COUNT["dir2"], len(contents), fmt.Sprintf("Dir2 should have 3 items %s", contents))

	cnt := CreateContainer(container.Name, t, router)
	assert.NotZero(t, cnt.ID, "We should have an ID now for the container")

	for _, mc := range contents {

		content := CreateResource(mc.Src, &cnt.ID, t, router)
		assert.NotZero(t, content.ID, fmt.Sprintf("Failed creating should create item %s", mc.Src))

		pCode, _, pErr := ValidateContentPreview(content.ID, router)
		assert.Equal(t, http.StatusOK, pCode, fmt.Sprintf("Failed to find preview for %s preview (%s)", content.Src, pErr))
	}
}

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

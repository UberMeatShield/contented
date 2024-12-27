package actions

import (
	"bytes"
	"contented/pkg/models"
	"contented/pkg/test_common"
	"contented/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateContentNamed(src string, containerID *int64, t *testing.T, router *gin.Engine) models.Content {
	mc := &models.Content{
		Src:         src,
		ContentType: "test",
		Preview:     "",
		ContainerID: containerID,
		NoFile:      true,
	}
	return CreateContent(mc, t, router)
}

func CreateContent(content *models.Content, t *testing.T, router *gin.Engine) models.Content {
	resObj := models.Content{}
	code, err := PostJson("/api/contents", content, &resObj, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Error creating %s", err))
	assert.Greater(t, resObj.ID, int64(0), fmt.Sprintf("It should have an ID %s", resObj))
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

	c1 := CreateNamedContainer("Trash1", t, router)
	c2 := CreateNamedContainer("Trash2", t, router)
	c3 := CreateNamedContainer("Trash3", t, router)
	c3.Hidden = true
	uCode, uErr := PutJson(fmt.Sprintf("/api/containers/%d", c3.ID), c3, &models.Container{}, router)
	assert.Equal(t, http.StatusOK, uCode, fmt.Sprintf("Failed to update %s", uErr))

	defer test_common.CleanupContainer(&c1)
	defer test_common.CleanupContainer(&c2)
	defer test_common.CleanupContainer(&c3)

	CreateContentNamed("a", &c1.ID, t, router)
	CreateContentNamed("b", &c1.ID, t, router)
	CreateContentNamed("c", &c2.ID, t, router)
	CreateContentNamed("donut", &c2.ID, t, router)
	CreateContentNamed("e", &c2.ID, t, router)

	// TODO: Update this to work with the memory setup too (no good way to update the Hidden)
	hiddenDonut := CreateContentNamed("donut_2_search_should_fail", &c3.ID, t, router)
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

	cnt := CreateNamedContainer(container.Name, t, router)
	assert.NotZero(t, cnt.ID, "We should have an ID now for the container")

	for _, mc := range contents {

		content := CreateContentNamed(mc.Src, &cnt.ID, t, router)
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
	CreateContentNamed(src, nil, t, router)
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
	content := CreateContentNamed(src, nil, t, router)

	url := fmt.Sprintf("/api/contents/%d", content.ID)
	validate := models.Content{}
	code, err := GetJson(url, "", &validate, router)
	assert.NoError(t, err, "Failed to get valid content")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, src, validate.Src)
}

func TestContentsResourceCreateDb(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	mc := CreateContentNamed("test_create", nil, t, router)
	assert.Greater(t, mc.ID, int64(0))
}

func TestContentsResourceCreateMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	mc := CreateContentNamed("test_create", nil, t, router)
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
	mc := CreateContentNamed("test_update", nil, t, router)
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
	mc := CreateContentNamed("NukeTest", nil, t, router)
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

func TestActionsMemoryScreensDestroy(t *testing.T) {
	cfg := test_common.InitMemoryFakeAppEmpty()
	router := setupRouter()
	assert.Equal(t, cfg.ReadOnly, false)
	ValidateContentScreensDestroy(t, router)
}

func TestActionsDbContentScreensDestroy(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(true)
	assert.Equal(t, cfg.ReadOnly, false)
	ValidateContentScreensDestroy(t, router)
}

func ValidateContentScreensDestroy(t *testing.T, router *gin.Engine) {
	containerPreviews, testDir := test_common.CreateTestPreviewsContainerDirectory(t)
	utils.ResetPreviewDir(containerPreviews)
	print("WHAT IS THE PREVIEW DIRECTORY %s", containerPreviews)

	container := models.Container{
		Name: "dir2",
		Path: testDir,
	}
	cnt := models.Container{}
	code, cErr := PostJson("/api/containers", container, &cnt, router)
	assert.NoError(t, cErr, "Failed to create the container %s", container.GetFqPath())
	assert.Equal(t, http.StatusCreated, code, "It should create the container %d", code)

	content := models.Content{
		Src:         "donut_[special( gunk.mp4",
		ContainerID: &cnt.ID,
	}
	contentCheck := models.Content{}
	code, cErr = PostJson("/api/contents", content, &contentCheck, router)
	assert.NoError(t, cErr, "It should create content %d", code)
	assert.Greater(t, contentCheck.ID, int64(0), "It should create a valid content")
	assert.Equal(t, http.StatusCreated, code, "It should create the content %d", code)

	// Create some content on disk so we can validate disk removal of these elements
	screenName1, err := test_common.WriteScreenFile(containerPreviews, "ScreenTestRemove", 1)
	assert.NoError(t, err, "Failed to write screen file: %v", err)
	screenName2, err := test_common.WriteScreenFile(containerPreviews, "ScreenTestRemove", 2)
	assert.NoError(t, err, "Failed to write screen file: %v", err)

	screen1 := models.Screen{ContentID: contentCheck.ID, Src: screenName1, Path: containerPreviews}
	screen1Check := models.Screen{}
	screen2 := models.Screen{ContentID: contentCheck.ID, Src: screenName2, Path: containerPreviews}
	screen2Check := models.Screen{}

	s1Code, s1Err := PostJson("/api/screens", screen1, &screen1Check, router)
	assert.NoError(t, s1Err, "It should create the screen %d", code)
	assert.Equal(t, http.StatusCreated, s1Code, "Screen create status code was unexpected %d", code)
	assert.Equal(t, screen1Check.ContentID, contentCheck.ID, "It should have the same ID")
	s2Code, s2Err := PostJson("/api/screens", screen2, &screen2Check, router)
	assert.NoError(t, s2Err, "It should create the second screen %d", code)
	assert.Equal(t, http.StatusCreated, s2Code, "Bad status code back from screen creation %d", code)
	assert.Equal(t, screen2Check.ContentID, contentCheck.ID, "It should have the same ID")

	// Verify directory now has two files
	files, err := os.ReadDir(containerPreviews)
	assert.NoError(t, err, "Should be able to read directory")
	assert.NotEmpty(t, files, "Directory should be empty before test")
	assert.Equal(t, 2, len(files), "Directory should have two files")

	screenCheck := ScreensResponse{}
	code, err = GetJson(fmt.Sprintf("/api/contents/%d/screens", contentCheck.ID), "", &screenCheck, router)
	assert.NoError(t, err, "It should get the screens")
	assert.Equal(t, http.StatusOK, code, "It should get the screens")
	assert.Equal(t, 2, len(screenCheck.Results), "It should have two screens")

	// Now delete the screens
	code, err = DeleteJson(fmt.Sprintf("/api/contents/%d/screens", contentCheck.ID), router)
	assert.NoError(t, err, "It should delete the screens")
	assert.Equal(t, http.StatusOK, code, "It should delete the screens")

	// Verify directory now has no files remaining
	files, err = os.ReadDir(containerPreviews)
	assert.NoError(t, err, "Should be able to read directory")
	assert.Equal(t, 0, len(files), "Directory should have two files %s", containerPreviews)
}

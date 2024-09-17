package actions

import (
	"bytes"
	"contented/pkg/managers"
	"contented/pkg/models"
	"contented/pkg/test_common"
	"contented/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	models.RebuildDatabase("test")
	code := m.Run()
	os.Exit(code)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	SetupRoutes(r)
	SetupWorkers()
	return r
}

func setupStatic() *gin.Engine {
	r := setupRouter()
	SetupStatic(r)
	return r
}

func InitFakeRouterApp(useDb bool) (*utils.DirConfigEntry, *gorm.DB, *gin.Engine) {
	cfg, db := test_common.InitFakeApp(useDb)
	return cfg, db, setupRouter()
}

// resObj is a &models.Content|Container|Screen|etc
func GetJson(url string, obj any, resObj any, router *gin.Engine) (int, error) {
	return HttpJson(url, obj, resObj, router, "GET")
}

func PostJson(url string, obj any, resObj any, router *gin.Engine) (int, error) {
	return HttpJson(url, obj, resObj, router, "POST")
}

func PutJson(url string, obj any, resObj any, router *gin.Engine) (int, error) {
	return HttpJson(url, obj, resObj, router, "PUT")
}

func DeleteJson(url string, router *gin.Engine) (int, error) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", url, nil)
	router.ServeHTTP(w, req)

	result := w.Result()
	if result.StatusCode != 200 {
		return result.StatusCode, fmt.Errorf("Error on delete %s", w.Body)
	}
	return result.StatusCode, nil
}

func HttpJson(url string, obj any, resObj any, router *gin.Engine, method string) (int, error) {
	w := httptest.NewRecorder()
	userJson, _ := json.Marshal(obj)
	req, _ := http.NewRequest(method, url, bytes.NewReader(userJson))
	router.ServeHTTP(w, req)

	result := w.Result()
	if !(result.StatusCode == 200 || result.StatusCode == 201) {
		log.Printf("Likely error getting data back from the server %s", w.Body)
		return result.StatusCode, fmt.Errorf("%s, %s", url, w.Body)
	}
	defer req.Body.Close()

	if result.ContentLength == 0 {
		log.Printf("Probably something went really wrong as there is no content body %d", result.StatusCode)
		return result.StatusCode, fmt.Errorf("No content from server %s", url)
	} else {
		log.Printf("%s response %s", url, w.Body)
		err := json.NewDecoder(w.Body).Decode(resObj)
		return result.StatusCode, err
	}
}

func MakeHttpRequest(url string, router *gin.Engine, method string) (int, *httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewReader([]byte{}))
	router.ServeHTTP(w, req)

	result := w.Result()
	if !(result.StatusCode == 200 || result.StatusCode == 201) {
		log.Printf("Likely error getting data back from the server %s", w.Body)
		return result.StatusCode, w, fmt.Errorf("%s, %s", url, w.Body)
	}
	defer req.Body.Close()

	if result.ContentLength == 0 {
		log.Printf("Probably something went really wrong as there is no content body %d", result.StatusCode)
		return result.StatusCode, w, fmt.Errorf("No content from server %s", url)
	} else {
		//log.Printf("%s response %s", url, w.Body)
		return result.StatusCode, w, nil
	}
}

// Removing action suite for all the tests is going to suck pretty hard
// Is there an AFTER all test option?  Just hard code the delete
func TestContainersList(t *testing.T) {
	test_common.InitFakeApp(false)

	router := setupRouter()
	req, _ := http.NewRequest("GET", "/api/containers", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	resObj := ContainersResponse{}
	json.NewDecoder(res.Body).Decode(&resObj)
	assert.Equal(t, test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(resObj.Results), "We should have this many dirs present")
}

func TestContentApplicationDirLoadMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)

	containerResponse := ContainersResponse{}
	code, err := GetJson("/api/containers", "", &containerResponse, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed to load containers %s", err))
	assert.Equal(t, test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(containerResponse.Results), "We should have this many dirs present")

	for _, c := range containerResponse.Results {
		url := fmt.Sprintf("/api/containers/%d/contents", c.ID)
		cntRes := ContentsResponse{}
		contentCode, contentErr := GetJson(url, "", &cntRes, router)
		assert.Equal(t, http.StatusOK, contentCode, fmt.Sprintf("Failed to load content %s", contentErr))

		expectCount := test_common.EXPECT_CNT_COUNT[c.Name]
		assert.Equal(t, test_common.EXPECT_CNT_COUNT[c.Name], len(cntRes.Results), fmt.Sprintf("Known content sizes %s", c))
		assert.Equal(t, int64(expectCount), cntRes.Total, "The count should be correct")
	}
}

// Tests that we can actually view the content directly vs as json responses
func TestViewHeaderCheck(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)

	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	mcs, count, err := man.ListContent(managers.ContentQuery{Page: 2, PerPage: 2})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(*mcs), "It should have only two results")
	assert.Greater(t, count, int64(2), "But the count should be the total")

	for _, mc := range *mcs {
		url := fmt.Sprintf("/api/view/%d", mc.ID)
		code, w, htmlErr := MakeHttpRequest(url, router, "GET")
		assert.Equal(t, http.StatusOK, code, fmt.Sprintf("View failed %s err %s", url, htmlErr))

		header := w.Header()
		assert.NoError(t, test_common.IsValidContentType(header.Get("Content-Type")))
	}
}

// Oof, that is rough... need a better way to select the file not by index but ID
func TestContentDirDownload(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)

	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	mcs, count, err := man.ListContent(managers.ContentQuery{Page: 2, PerPage: 2})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(*mcs), "It should have only two results")
	assert.Greater(t, count, int64(2), "It should have more content")

	for _, mc := range *mcs {

		url := fmt.Sprintf("/api/download/%d", mc.ID)
		code, w, err := MakeHttpRequest(url, router, "GET")
		assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed to download %s", err))

		header := w.Header()
		assert.NoError(t, test_common.IsValidContentType(header.Get("Content-Type")))
	}

	// Prevent evil check (is this working?)
	content := models.Content{NoFile: true, Description: "not a real boy", Src: "~/.ssh/id_rsa"}
	assert.NoError(t, man.CreateContent(&content))

	evilUrl := fmt.Sprintf("/api/download/%d", content.ID)
	evilCode, w, evilErr := MakeHttpRequest(evilUrl, router, "GET")
	assert.Equal(t, http.StatusOK, evilCode, fmt.Sprintf("Download should not work %s", evilErr))

	expectContent := "application/json; charset=utf-8"
	assert.Equal(t, expectContent, w.Header().Get("Content-Type"), fmt.Sprintf("This should be json %s", w.Body))

	checkNoFile := models.Content{}
	json.NewDecoder(w.Body).Decode(&checkNoFile)
	assert.Equal(t, content.Description, checkNoFile.Description)
}

// Test if we can get the actual file using just a file ID
func TestFindAndLoadFile(t *testing.T) {
	cfg, _, _ := InitFakeRouterApp(false)
	assert.Equal(t, true, cfg.Initialized)

	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	assert.NoError(t, err)

	// TODO: Should make the hidden file actually a file somehow
	for _, mc := range *mcs {
		if mc.Hidden == false {
			mc_ref, fc_err := man.GetContent(mc.ID)
			assert.NoError(t, fc_err, "And an initialized app should index correctly")

			fq_path, err := man.FindActualFile(mc_ref)
			assert.NoError(t, err, "It should find all these files")

			_, o_err := os.Stat(fq_path)
			assert.NoError(t, o_err, "The fully qualified path did not exist")
		}
	}
}

// This checks that a preview loads when defined and otherwise falls back to the MC itself
func TestPreviewFileMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	assert.NoError(t, err)

	for _, mc := range *mcs {
		url := fmt.Sprintf("/api/preview/%d", +mc.ID)
		code, w, err := MakeHttpRequest(url, router, "GET")
		assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed to preview %s", err))

		header := w.Header()
		assert.NoError(t, test_common.IsValidContentType(header.Get("Content-Type")))
	}
}

// This checks if previews are actually used if defined
func TestPreviewApiLoadsNormalContent(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	assert.NoError(t, err)

	for _, mc := range *mcs {
		if mc.Preview != "" {
			url := fmt.Sprintf("/api/preview/%d", mc.ID)
			code, _, err := MakeHttpRequest(url, router, "GET")
			assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed preview %s", err))
		}
	}
}

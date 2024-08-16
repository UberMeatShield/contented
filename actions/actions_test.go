package actions

import (
	"bytes"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	SetupRoutes(r)
	return r
}

func setupStatic() *gin.Engine {
	r := gin.Default()
	SetupStatic(r)
	SetupRoutes(r)
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

/*
func (as *ActionSuite) Test_ContentDirLoad() {
	test_common.InitFakeApp(false)

	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
	cnts := ContainersResponse{}
	json.NewDecoder(res.Body).Decode(&cnts)
	as.Equal(test_common.TOTAL_CONTAINERS_WITH_CONTENT, len(cnts.Results), "We should have this many dirs present")

	for _, c := range cnts.Results {
		url := fmt.Sprintf("/containers/%d/contents", c.ID)
		res := as.JSON(url).Get()
		as.Equal(http.StatusOK, res.Code)

		cntRes := ContentsResponse{}
		json.NewDecoder(res.Body).Decode(&cntRes)
		if c.Name == "dir1" {
			as.Equal(12, len(cntRes.Results), fmt.Sprintf("Known content sizes %s", res.Body.String()))
			as.Equal(12, cntRes.Total, "The count should be correct")
		}
	}
}

func (as *ActionSuite) Test_ViewRef() {
	// Oof, that is rough... need a better way to select the file not by index but ID
	test_common.InitFakeApp(false)

	app := as.App
	ctx := test_common.GetContext(app)
	man := managers.GetManager(&ctx)
	mcs, count, err := man.ListContent(managers.ContentQuery{Page: 2, PerPage: 2})
	as.NoError(err)
	as.Equal(2, len(*mcs), "It should have only two results")
	as.Greater(count, 2, "But the count should be the total")

	// TODO: Make it better about the type checking
	// TODO: Make it always pass in the file ID
	for _, mc := range *mcs {
		res := as.HTML("/view/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)
		header := res.Header()
		as.NoError(test_common.IsValidContentType(header.Get("Content-Type")))
	}
}

// Oof, that is rough... need a better way to select the file not by index but ID
func (as *ActionSuite) Test_ContentDirDownload() {
	test_common.InitFakeApp(false)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, count, err := man.ListContent(managers.ContentQuery{Page: 2, PerPage: 2})
	as.NoError(err)
	as.Equal(2, len(*mcs), "It should have only two results")
	as.Greater(count, 2, "It should have more content")

	for _, mc := range *mcs {
		res := as.HTML("/download/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)
		header := res.Header()
		as.NoError(test_common.IsValidContentType(header.Get("Content-Type")))
	}

	// Prevent evil check
	content := models.Content{NoFile: true, Description: "not a real boy", Src: "~/.ssh/id_rsa"}
	as.NoError(man.CreateContent(&content))

	noFileRes := as.JSON(fmt.Sprintf("/download/%s", content.ID.String())).Get()
	as.Equal(http.StatusOK, noFileRes.Code)

	checkNoFile := models.Content{}
	json.NewDecoder(noFileRes.Body).Decode(&checkNoFile)
	as.Equal(content.Description, checkNoFile.Description)
}

// Test if we can get the actual file using just a file ID
func (as *ActionSuite) Test_FindAndLoadFile() {
	cfg := test_common.InitFakeApp(false)

	as.Equal(true, cfg.Initialized)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	as.NoError(err)

	// TODO: Should make the hidden file actually a file somehow
	for _, mc := range *mcs {
		if mc.Hidden == false {
			mc_ref, fc_err := man.GetContent(mc.ID)
			as.NoError(fc_err, "And an initialized app should index correctly")

			fq_path, err := man.FindActualFile(mc_ref)
			as.NoError(err, "It should find all these files")

			_, o_err := os.Stat(fq_path)
			as.NoError(o_err, "The fully qualified path did not exist")
		}
	}
}

// This checks that a preview loads when defined and otherwise falls back to the MC itself
func (as *ActionSuite) Test_PreviewFile() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	as.NoError(err)

	for _, mc := range *mcs {
		res := as.HTML("/preview/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)

		header := res.Header()
		as.NoError(test_common.IsValidContentType(header.Get("Content-Type")))
	}
}

func (as *ActionSuite) Test_FullFile() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	as.NoError(err)

	for _, mc := range *mcs {
		if mc.Hidden == false {
			res := as.HTML("/view/" + mc.ID.String()).Get()
			as.Equal(http.StatusOK, res.Code)

			header := res.Header()
			as.NoError(test_common.IsValidContentType(header.Get("Content-Type")))
		}
	}
}

// This checks if previews are actually used if defined
func (as *ActionSuite) Test_PreviewWorking() {
	test_common.InitFakeApp(false)
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, _, err := man.ListContent(managers.ContentQuery{})
	as.NoError(err)

	for _, mc := range *mcs {
		if mc.Preview != "" {
			res := as.HTML("/preview/" + mc.ID.String()).Get()
			as.Equal(http.StatusOK, res.Code)
			fmt.Println("Not modified")
		}
	}
}

func Test_ManagerSuite(t *testing.T) {
	action := suite.NewAction(App(true))
	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

*/

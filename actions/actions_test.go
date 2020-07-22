package actions

import (
  "os"
  "fmt"
  "encoding/json"
  "net/http"
  "testing"
  "contented/utils"
  "github.com/gobuffalo/envy"
  "github.com/gobuffalo/packr/v2"
  "github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func TestMain(m *testing.M) {
    dir, err := envy.MustGet("DIR")
    cfg.Dir = dir  // Absolute path to our known test data
    cfg.ValidDirs = utils.GetDirectoriesLookup(cfg.Dir)

    if err != nil {
        fmt.Println("DIR ENV REQUIRED$ export=DIR=`pwd`/mocks/content/ && buffalo test")
        panic(err)
    }
    code := m.Run() 
    os.Exit(code)
}

func (as *ActionSuite) Test_ContentList() {
    res := as.JSON("/content/").Get()
    as.Equal(http.StatusOK, res.Code)

    resObj := PreviewResults{}
    json.NewDecoder(res.Body).Decode(&resObj)
    as.Equal(4, len(resObj.Results), "We should have this many directories present")

    // Check all the directories have contents
    for _, dir := range resObj.Results {
        as.Greater(dir.Total, 0, "There should be multiple results")
        as.NotNil(dir.Id, "There should be a directory ID")
        as.Greater(len(dir.Contents), 0,  "And all our test mocks have content")
    }
}

func (as *ActionSuite) Test_ContentDirLoad() {
    res := as.JSON("/content/dir1").Get()
    as.Equal(http.StatusOK, res.Code)

    resObj := utils.DirContents{}
    json.NewDecoder(res.Body).Decode(&resObj)
    as.Equal(resObj.Total, 11, "It should have a known number of images")
}

func (as *ActionSuite) Test_ViewRef() {
    res := as.HTML("/view/dir1/1").Get()
    as.Equal(http.StatusOK, res.Code)
    header := res.Header()
    as.Equal("image/jpeg", header.Get("Content-Type"))
}

func (as *ActionSuite) Test_ContentDirDownload() {
    res := as.HTML("/download/dir1/6DPrYve.jpg").Get()
    as.Equal(http.StatusOK, res.Code)
    
    header := res.Header()
    as.Equal("image/jpeg", header.Get("Content-Type"))
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

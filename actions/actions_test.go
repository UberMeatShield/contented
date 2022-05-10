package actions

import (
	"contented/internals"
	"contented/managers"
	"contented/models"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite/v3"
	"net/http"
	"os"
	"testing"
)

const ExpectCntCount = 5

type ActionSuite struct {
	*suite.Action
}

func TestMain(m *testing.M) {
	dir, err := envy.MustGet("DIR")
	fmt.Printf("Using this test directory %s", dir)
	if err != nil {
		fmt.Println("DIR ENV REQUIRED$ export=DIR=`pwd`/mocks/content/ && buffalo test")
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func (as *ActionSuite) Test_ContentList() {
	internals.InitFakeApp(false)

	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
	resObj := models.Containers{}
	json.NewDecoder(res.Body).Decode(&resObj)
	as.Equal(ExpectCntCount, len(resObj), "We should have this many dirs present")
}

func (as *ActionSuite) Test_ContentDirLoad() {
	internals.InitFakeApp(false)

	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
	cnts := models.Containers{}
	json.NewDecoder(res.Body).Decode(&cnts)
	as.Equal(ExpectCntCount, len(cnts), "We should have this many dirs present")

	for _, c := range cnts {
		res := as.JSON("/containers/" + c.ID.String() + "/media").Get()
		as.Equal(http.StatusOK, res.Code)

		resObj := []models.Containers{}
		json.NewDecoder(res.Body).Decode(&resObj)

		fmt.Printf("What was the result %s\n", resObj)

		if c.Name == "dir1" {
			as.Equal(len(resObj), 12, "It should have a known number of images")
		}
	}
}

func (as *ActionSuite) Test_ViewRef() {
	// Oof, that is rough... need a better way to select the file not by index but ID
	internals.InitFakeApp(false)

	app := as.App
	ctx := internals.GetContext(app)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(2, 2)
	as.NoError(err)
	as.Equal(2, len(*mcs), "It should have only two results")

	// TODO: Make it better about the type checking
	// TODO: Make it always pass in the file ID
	for _, mc := range *mcs {
		res := as.HTML("/view/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)
		header := res.Header()
		as.NoError(internals.IsValidContentType(header.Get("Content-Type")))

	}
}

// Oof, that is rough... need a better way to select the file not by index but ID
func (as *ActionSuite) Test_ContentDirDownload() {
	internals.InitFakeApp(false)

	ctx := internals.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(2, 2)
	as.NoError(err)
	as.Equal(2, len(*mcs), "It should have only two results")

	// Hate
	for _, mc := range *mcs {
		res := as.HTML("/download/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)
		header := res.Header()
		as.NoError(internals.IsValidContentType(header.Get("Content-Type")))
	}
}

// Test if we can get the actual file using just a file ID
func (as *ActionSuite) Test_FindAndLoadFile() {
	cfg := internals.InitFakeApp(false)

	as.Equal(true, cfg.Initialized)

	ctx := internals.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(1, 200)
	as.NoError(err)

	for _, mc := range *mcs {
		mc_ref, fc_err := man.FindFileRef(mc.ID)
		as.NoError(fc_err, "And an initialized app should index correctly")

		fq_path, err := man.FindActualFile(mc_ref)
		as.NoError(err, "It should find all these files")

		_, o_err := os.Stat(fq_path)
		as.NoError(o_err, "The fully qualified path did not exist")
	}
}

// This checks that a preview loads when defined and otherwise falls back to the MC itself
func (as *ActionSuite) Test_PreviewFile() {
	internals.InitFakeApp(false)
	ctx := internals.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(1, 200)
	as.NoError(err)

	for _, mc := range *mcs {
		res := as.HTML("/preview/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)

		header := res.Header()
		as.NoError(internals.IsValidContentType(header.Get("Content-Type")))
	}
}

func (as *ActionSuite) Test_FullFile() {
	internals.InitFakeApp(false)
	ctx := internals.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(1, 200)
	as.NoError(err)

	for _, mc := range *mcs {
		res := as.HTML("/view/" + mc.ID.String()).Get()
		as.Equal(http.StatusOK, res.Code)

		header := res.Header()
		as.NoError(internals.IsValidContentType(header.Get("Content-Type")))
	}
}

// This checks if previews are actually used if defined
func (as *ActionSuite) Test_PreviewWorking() {
	internals.InitFakeApp(false)
	ctx := internals.GetContext(as.App)
	man := managers.GetManager(&ctx)
	mcs, err := man.ListAllMedia(1, 200)
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
	action, err := suite.NewActionWithFixtures(App(true), packr.New("Test_ManagerSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
    "context"
	"testing"
	"contented/models"
	"contented/utils"
	"contented/managers"
    "contented/internals"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func getContext(app *buffalo.App) buffalo.Context {
    return getContextParams(app, "/containers", "1", "10")
}

func getContextParams(app *buffalo.App, url string, page string, per_page string) buffalo.Context {
    req, _ := http.NewRequest("GET", url, nil)

    // Setting a value here like this doesn't seem to work correctly.  The context will NOT
    // Actually keep the param.  GoLang made this a huge pain to test vs a nice simple SetParam
    ctx := req.Context()
    ctx = context.WithValue(ctx, "page", page)
    ctx = context.WithValue(ctx, "per_page", per_page)
    ctx = context.WithValue(ctx, "tx", models.DB)

    return &buffalo.DefaultContext{
        Context: ctx,
    }
}


func ResetConfig() *utils.DirConfigEntry {
    cfg := utils.GetCfgDefaults()
    dir, _ := envy.MustGet("DIR")
    cfg.Dir = dir
    utils.InitConfig(dir, &cfg)
    utils.SetCfg(cfg)
    return utils.GetCfg()
}

// This function is now how the init method should function till caching is implemented
// As the internals / guts are functional using the new models the creation of models
// can be removed.
func init_fake_app(use_db bool) *utils.DirConfigEntry {
	dir, _ := envy.MustGet("DIR")
	fmt.Printf("Using directory %s\n", dir)

	cfg := ResetConfig()
	utils.InitConfig(dir, cfg)
    cfg.UseDatabase = use_db  // Set via .env or USE_DATABASE as an environment var
    cfg.StaticResourcePath = "./public/build"

    // TODO: Assign the context into the manager (force it?)
    if cfg.UseDatabase == false {
        memStorage := utils.InitializeMemory(dir)

        // cnts := memStorage.ValidContainers
        // for _, c := range cnts {
        mcs := memStorage.ValidMedia
        for _, mc := range mcs {
           if mc.Src == "this_is_p_ng" {
               mc.Preview = "preview_this_is_p_ng"
           }
        }
       // }
    }
	return cfg
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
    init_fake_app(false)

	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
	resObj := models.Containers{}
	json.NewDecoder(res.Body).Decode(&resObj)
	as.Equal(4, len(resObj), "We should have this many dirs present")
}

func (as *ActionSuite) Test_ContentDirLoad() {
	init_fake_app(false)

	res := as.JSON("/containers").Get()
	as.Equal(http.StatusOK, res.Code)
	cnts := models.Containers{}
	json.NewDecoder(res.Body).Decode(&cnts)
	as.Equal(4, len(cnts), "We should have this many dirs present")

	for _, c := range cnts {
		res := as.JSON("/containers/" + c.ID.String()  + "/media").Get()
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
	init_fake_app(false)

    app := as.App
    ctx := getContext(app)
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
	init_fake_app(false)

    ctx := getContext(as.App)
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
	cfg := init_fake_app(false)

	as.Equal(true, cfg.Initialized)

    ctx := getContext(as.App)
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
	init_fake_app(false)
    ctx := getContext(as.App)
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
	init_fake_app(false)
    ctx := getContext(as.App)
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
	init_fake_app(false)
    ctx := getContext(as.App)
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

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(true), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

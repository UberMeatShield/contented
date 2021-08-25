package internals

/**
 * These are some common test helper files dealing mostly with unit test setup.
 */
import (
    "log"
	"fmt"
    "errors"
	"net/http"
    "context"
	"contented/models"
	"contented/utils"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
    contenttype "github.com/gobuffalo/mw-contenttype"
    paramlogger "github.com/gobuffalo/mw-paramlogger"
    "github.com/gobuffalo/x/sessions"
    "github.com/rs/cors"
    "github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
)

// Create the basic app but without routes, useful for testing the managers but not routes
func CreateBuffaloApp(UseDatabase bool, env string) *buffalo.App {
    app := buffalo.New(buffalo.Options{
        Env:          env,
        SessionStore: sessions.Null{},
        PreWares: []buffalo.PreWare{
            cors.Default().Handler,
        },
        SessionName: "_content_session",
    })

    // Log request parameters (filters apply) this should maybe be only in dev
    app.Use(paramlogger.ParameterLogger)

    // Wraps each request in a transaction. Remove to disable this.
    if UseDatabase == true {
        log.Printf("Connecting to the database\n")
        app.Use(popmw.Transaction(models.DB))
    } else {
        log.Printf("This code will attempt to use memory management \n")
    }

    // Set the request content type to JSON
    app.Use(contenttype.Set("application/json"))
    return app
}


func GetContext(app *buffalo.App) buffalo.Context {
    return GetContextParams(app, "/containers", "1", "10")
}

func GetContextParams(app *buffalo.App, url string, page string, per_page string) buffalo.Context {
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

// TODO validate octet/stream
func IsValidContentType(content_type string) error {
    valid := map[string]bool{
        "image/png": true,
        "image/jpeg": true,
        "application/octet-stream": true,
        "video/mp4": true,
    }
    if valid[content_type] {
        return nil 
    }
    return errors.New("Invalid content type: " + content_type)
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
func InitFakeApp(use_db bool) *utils.DirConfigEntry {
	dir, _ := envy.MustGet("DIR")
	fmt.Printf("Using directory %s\n", dir)

	cfg := ResetConfig()
	utils.InitConfig(dir, cfg)
    cfg.UseDatabase = use_db  // Set via .env or USE_DATABASE as an environment var
    cfg.StaticResourcePath = "./public/build"

    // TODO: Assign the context into the manager (force it?)
    if cfg.UseDatabase == false {

        // TODO: This moves into managers.. is there a sane way of handling this?
        memStorage := utils.InitializeMemory(dir)

        // cnts := memStorage.ValidContainers
        // for _, c := range cnts {
        mcs := memStorage.ValidMedia
        for _, mc := range mcs {
           if mc.Src == "this_is_p_ng" {
               mc.Preview = "preview_this_is_p_ng"
           }
        }
    }
	return cfg
}

func GetMediaByDirName(test_dir_name string) (*models.Container, models.MediaContainers) {
    dir, _ := envy.MustGet("DIR")
    cfg := utils.GetCfg()
    cfg.Dir = dir
    cnts := utils.FindContainers(cfg.Dir)

    var cnt *models.Container = nil
    for _, c := range cnts {
        if c.Name == test_dir_name {
            cnt = &c
            break
        }
    }
    if cnt == nil {
        log.Panic("Could not find the directory: " +  test_dir_name)
    }
    media := utils.FindMedia(*cnt, 42, 0)
    cnt.Total = len(media)
    return cnt, media
}

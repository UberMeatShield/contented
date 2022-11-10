package internals

/**
 * These are some common test helper files dealing mostly with unit test setup and
 * mock data counts and information.
 */
import (
    "contented/models"
    "contented/utils"
    "context"
    "errors"
    "fmt"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
    "github.com/gobuffalo/envy"
    contenttype "github.com/gobuffalo/mw-contenttype"
    paramlogger "github.com/gobuffalo/mw-paramlogger"
    "github.com/gobuffalo/nulls"
    "github.com/gobuffalo/x/sessions"
    "github.com/rs/cors"
    "log"
    "net/http"
    "path/filepath"
)

const TOTAL_CONTAINERS = 5
const TOTAL_MEDIA = 31
const VIDEO_FILENAME = "donut_[special( gunk.mp4"

// Helper for a common block of video test code (duplicated in the utils test)
func Get_VideoAndSetupPaths() (string, string, string) {
    var testDir, _ = envy.MustGet("DIR")
    srcDir := filepath.Join(testDir, "dir2")
    dstDir := utils.GetPreviewDst(srcDir)
    testFile := "donut.mp4"

    // Ensure that the preview destination directory is clean
    utils.ResetPreviewDir(dstDir)
    return srcDir, dstDir, testFile
}

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
        "image/png":                true,
        "image/jpeg":               true,
        "image/gif":                true,
        "image/webp":               true,
        "application/octet-stream": true,
        "video/mp4":                true,
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
    utils.SetupContainerMatchers(&cfg, "", "DS_Store|container_previews")
    utils.SetupContentMatchers(&cfg, "", "image|video", "DS_Store", "")
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
    cfg.UseDatabase = use_db // Set via .env or USE_DATABASE as an environment var
    cfg.StaticResourcePath = "./public/build"

    // TODO: Assign the context into the manager (force it?)
    if cfg.UseDatabase == false {

        // TODO: This moves into managers.. is there a sane way of handling this?
        memStorage := utils.InitializeMemory(dir)

        // cnts := memStorage.ValidContainers
        // for _, c := range cnts {
        mcs := memStorage.ValidContent
        for _, mc := range mcs {
            if mc.Src == "this_is_p_ng" {
                mc.Preview = "preview_this_is_p_ng"
            }
        }
    }
    return cfg
}

func CreateContentByDirName(test_dir_name string) (*models.Container, models.Contents, error) {
    cnt, content := GetContentByDirName(test_dir_name)

    c_err := models.DB.Create(cnt)
    if c_err != nil {
        return nil, nil, c_err
    }
    for _, mc := range content {
        mc.ContainerID = nulls.NewUUID(cnt.ID)
        m_err := models.DB.Create(&mc)
        if m_err != nil {
            return nil, nil, m_err
        }
    }
    return cnt, content, nil
}

func GetContentByDirName(test_dir_name string) (*models.Container, models.Contents) {
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
        log.Panic("Could not find the directory: " + test_dir_name)
    }
    content := utils.FindContentMatcher(*cnt, 42, 0, cfg.IncContent, cfg.ExcContent)
    cnt.Total = len(content)
    return cnt, content
}

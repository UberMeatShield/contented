package managers

import (
    "contented/internals"
    "contented/models"
    "contented/utils"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/pop/v6"
    "github.com/gobuffalo/suite/v3"
    "log"
    "net/url"
    "os"
    "testing"
)

var expect_len = map[string]int{
    "dir1":            12,
    "dir2":            3,
    "dir3":            10,
    "screens":         4,
    "screens_sub_dir": 2,
}

func GetManagerActionSuite(cfg *utils.DirConfigEntry, as *ActionSuite) ContentManager {
    ctx := internals.GetContext(as.App)
    get_params := func() *url.Values {
        vals := ctx.Params().(url.Values)
        return &vals
    }
    get_conn := func() *pop.Connection {
        // as.DB should work, but it is of a type pop.v5.Connection instead of pop.Connection
        return models.DB
    }
    return CreateManager(cfg, get_conn, get_params)
}

// Why are no tests working?
func TestMain(m *testing.M) {
    _, err := envy.MustGet("DIR")
    if err != nil {
        log.Println("DIR ENV REQUIRED$ export=DIR=`pwd`/mocks/content/ && buffalo test")
        panic(err)
    }
    code := m.Run()
    os.Exit(code)
}

func Test_ManagerSuite(t *testing.T) {
    app := internals.CreateBuffaloApp(true, "test")
    action := suite.NewAction(app)
    as := &ActionSuite{
        Action: action,
    }
    suite.Run(t, as)
}

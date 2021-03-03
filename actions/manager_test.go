package actions

import (
	//"fmt"
	//"contented/models"
	"contented/utils"
    /*
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite"
	"github.com/gofrs/uuid"
	"github.com/gobuffalo/envy"
    "github.com/gobuffalo/buffalo"
    */
    "context"
    //"sync"
    //"github.com/gobuffalo/logger"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/buffalo"
)

var expect_len = map[string]int{
    "dir1": 12,
    "dir2": 2,
    "dir3": 6,
    "screens": 3,
}

func basicContext() buffalo.DefaultContext {
	return buffalo.DefaultContext{
		Context: context.Background(),
		//logger:  buffalo.logger.New(logger.DebugLevel),
		//data:    &sync.Map{},
		//flash:   &Flash{data: make(map[string][]string)},
	}
}

func (as *ActionSuite) Test_ManagerContainers() {
    init_fake_app(false)
    ctx := getContext(app)
    man := GetManager(&ctx)
    containers, err := man.ListContainersContext()
    as.NoError(err)

    for _, c := range *containers {
        c_mem, err := man.FindDirRef(c.ID)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(c_mem.ID, c.ID)
    }
}


func (as *ActionSuite) Test_ManagerMediaContainer() {
    init_fake_app(false)
    ctx := getContext(app)
    man := GetManager(&ctx)
    mcs, err := man.ListAllMedia(1, 9001)
    as.NoError(err)

    for _, mc := range *mcs {
        cm, err := man.FindFileRef(mc.ID)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(cm.ID, mc.ID)
    }
}

func (as *ActionSuite) Test_AssignManager() {
    dir, _ := envy.MustGet("DIR")
    cfg := GetCfg()
    cfg.UseDatabase = false
    utils.InitConfig(dir, cfg)

    mem := ContentManagerMemory{}
    mem.validate = "Memory"
    mem.SetCfg(cfg)
    mem.Initialize()

    memCfg := mem.GetCfg()
    as.NotNil(memCfg, "It should be defined")
    mcs, err := mem.ListAllMedia(1, 9001)
    as.NoError(err)
    as.Greater(len(*mcs), 0, "It should have valid files in the manager")

    appCfg.UseDatabase = false
    ctx := getContext(app)
    man := GetManager(&ctx)  // New Reference but should have the same count of media
    mcs_2, _ := man.ListAllMedia(1, 9000)

    as.Equal(len(*mcs), len(*mcs_2), "A new instance should use the same storage")
}



func (as *ActionSuite) Test_ManagerInitialize() {
    cfg := init_fake_app(false)
    cfg.UseDatabase = false

    ctx := getContext(app)
    man := GetManager(&ctx)
    as.NotNil(man, "It should have a manager defined after init")

    containers, err := man.ListContainersContext()
    as.NoError(err, "It should list all containers")
    as.NotNil(containers, "It should have containers")
    as.Equal(len(*containers), 4, "It should have 4 of them")

    // Memory test working
    for _, c := range *containers {
        // fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
        media, err := man.ListMediaContext(c.ID)
        as.NoError(err)
        as.NotNil(media)

        media_len := len(*media)
        // fmt.Printf("Media length was %d\n", media_len)
        as.Greater(media_len, 0, "There should be a number of media")
        as.Equal(expect_len[c.Name], media_len, "It should have this many instances: " + c.Name )
    }
}

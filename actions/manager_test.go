package actions

import (
	"fmt"
	//"contented/models"
	//"contented/utils"
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
    cfg := init_fake_app()
    mem := ContentManagerMemory{}
    mem.cfg = cfg

    for id, _ := range cfg.ValidContainers {
        c_mem, err := mem.FindDirRef(id)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(c_mem.ID, id)
    }
}


func (as *ActionSuite) Test_ManagerMediaContainer() {
    cfg := init_fake_app()
    mem := ContentManagerMemory{}
    mem.cfg = cfg

    for id, _ := range cfg.ValidFiles {
        cm, err := mem.FindFileRef(id)
        if err != nil {
            as.Fail("It should not have an issue finding valid containers")
        }
        as.Equal(cm.ID, id)
    }
}

func (as *ActionSuite) Test_AssignManager() {
    mem := ContentManagerMemory{}
    mem.validate = "Memory"
    cfg := init_fake_app()
    mem.SetCfg(cfg)
    SetManager(mem)

    as.Greater(len(cfg.ValidFiles), 0, "It should have valid files in the config")

    var man ContentManager = GetManager()
    memCfg := man.GetCfg()
    as.NotNil(memCfg, "It should be defined")
    //as.Greater(len(memCfg.ValidFiles), 0, "There should be a config entry")
    //as.Equal(len(memCfg.ValidFiles), len(cfg.ValidFiles))
}



func (as *ActionSuite) Test_ManagerInitialize() {
    cfg := init_fake_app()
    cfg.UseDatabase = false
    SetupManager(cfg)

    man := GetManager()
    as.NotNil(man, "It should have a manager defined after init")

    containers := man.ListContainersContext()
    as.NotNil(containers, "It should have containers")
    as.Equal(len(*containers), 4, "It should have 4 of them")

    for _, c := range *containers {
        // fmt.Printf("Searching for this container %s with name %s\n", c.ID, c.Name)
        media := man.ListMediaContext(c.ID)
        as.NotNil(media)

        media_len := len(*media)
        // fmt.Printf("Media length was %d\n", media_len)
        as.Greater(media_len, 0, "There should be a number of media")
        as.Equal(expect_len[c.Name], media_len, "It should have this many instances: " + c.Name )
    }
    // as.Greater(len(containers), 0, "There should be valid containers")
    // as.Greater(len(man.ListMedia()), 0, "There should be valid files")

}

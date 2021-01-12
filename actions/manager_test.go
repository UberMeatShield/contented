package actions

import (
	//"contented/models"
	//"contented/utils"
    /*
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite"
	"github.com/gofrs/uuid"
	"github.com/gobuffalo/envy"
    */
)


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

    containers := man.ListContainers()
    as.NotNil(containers, "It should have containers")
    as.Greater(len(*containers), 0, "It should have")

    /*
    for idx, c := range containers {
        as.NotNil(man.ListMedia(c.ID))
    }
    */
    // as.Greater(len(containers), 0, "There should be valid containers")
    // as.Greater(len(man.ListMedia()), 0, "There should be valid files")

}

package grifts

import (
    "fmt"
//    "contented/models"
    "contented/actions"
    "contented/utils"
	"github.com/markbates/grift/grift"
    "github.com/pkg/errors"
    "github.com/gobuffalo/envy"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Populate the DB with a set of directory content.")
	grift.Add("seed", func(c *grift.Context) error {
        cfg := utils.GetCfg()
        utils.InitConfigEnvy(cfg)

        // Require the directory which we want to process (maybe just trust it is set)
        _, d_err := envy.MustGet("DIR")
        if d_err != nil {
            return errors.WithStack(d_err)
        }
        // Clean out the current DB setup then builds a new one
        fmt.Printf("Configuration is loaded %s Starting import", cfg)
        return actions.CreateInitialStructure(cfg.Dir)
	})
    // Then add the content for the entire directory structure

	grift.Add("preview", func(c *grift.Context) error {
        cfg := utils.GetCfg()
        utils.InitConfigEnvy(cfg)
        fmt.Printf("Configuration is loaded %s doing preview creation", cfg)

        // TODO: This should use the managers probably (allow for memory manager previews)
        return actions.CreateAllPreviews(cfg.PreviewOverSize)
	})
})

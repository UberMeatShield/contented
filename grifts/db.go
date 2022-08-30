package grifts

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"fmt"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/grift/grift"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"net/url"
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
		fmt.Printf("Configuration is loaded %s Starting import", cfg.Dir)
		return managers.CreateInitialStructure(cfg)
	})

	// It really feels like a grift should do a better command line handler
	// TODO: It must right?  Find it at some point
	grift.Add("nuke", func(c *grift.Context) error {
		nuke, d_err := envy.MustGet("NUKE_IT")
		if d_err != nil || nuke != "y" {
			return errors.New("NUKE_IT env must be defined and set to 'y' to delete")
		}
		return models.DB.TruncateAll()
	})

	grift.Add("preview", func(c *grift.Context) error {
		cfg := utils.GetCfg()
		utils.InitConfigEnvy(cfg)
		fmt.Printf("Configuration is loaded %s doing preview creation", cfg.Dir)

		get_params := func() *url.Values {
			vals := url.Values{} // TODO: Maybe set this via something sensible
			return &vals
		}
		if cfg.UseDatabase {
			// The scope of transactions is a bit odd.  Seems like this is handled in
			// buffalo via the magical buffalo middleware.
			fmt.Printf("DB Manager setup")
			return models.DB.Transaction(func(tx *pop.Connection) error {
				get_connection := func() *pop.Connection {
					return tx
				}
				man := managers.CreateManager(cfg, get_connection, get_params)
				fmt.Printf("Creating previews %t", man.CanEdit())
				return managers.CreateAllPreviews(man)
			})
		} else {
			get_connection := func() *pop.Connection {
				return nil // Do not do anything with the DB
			}
			man := managers.CreateManager(cfg, get_connection, get_params)
			fmt.Printf("Use memory manager %t", man.CanEdit())
			return managers.CreateAllPreviews(man)
		}
	})
})

package grifts

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"fmt"
	"net/url"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/grift/grift"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
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

	get_params := func() *url.Values {
		vals := url.Values{} // TODO: Maybe set this via something sensible
		return &vals
	}
	no_connection := func() *pop.Connection {
		return nil // Do not do anything with the DB
	}
	get_worker := func() worker.Worker {
		return nil
	}

	grift.Add("preview", func(c *grift.Context) error {
		cfg := utils.GetCfg()
		utils.InitConfigEnvy(cfg)
		fmt.Printf("Configuration is loaded %s doing preview creation", cfg.Dir)

		if cfg.UseDatabase {
			// The scope of transactions is a bit odd.  Seems like this is handled in
			// buffalo via the magical buffalo middleware
			fmt.Printf("DB Manager setup")
			return models.DB.Transaction(func(tx *pop.Connection) error {
				get_connection := func() *pop.Connection {
					return tx
				}
				man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
				fmt.Printf("Creating previews images %t with db manager", man.CanEdit())
				return managers.CreateAllPreviews(man)
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
			fmt.Printf("Use memory manager %t for preview images", man.CanEdit())
			return managers.CreateAllPreviews(man)
		}
	})

	grift.Add("encode", func(c *grift.Context) error {
		cfg := utils.GetCfg()
		utils.InitConfigEnvy(cfg)
		fmt.Printf("Configuration is loaded %s Starting Video conversion", cfg.Dir)

		if cfg.UseDatabase {
			// The scope of transactions is a bit odd.  Seems like this is handled in
			// buffalo via the magical buffalo middleware
			fmt.Printf("DB Manager setup Video encoding lookup")
			return models.DB.Transaction(func(tx *pop.Connection) error {
				get_connection := func() *pop.Connection {
					return tx
				}
				man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
				fmt.Printf("Starting to encode %t TODO: Summary of config", man.CanEdit())
				return managers.EncodeVideos(man)
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
			fmt.Printf("Use memory manager %t for video encoding lookups", man.CanEdit())
			return managers.EncodeVideos(man)
		}
	})

	grift.Add("tags", func(c *grift.Context) error {
		cfg := utils.GetCfg()
		utils.InitConfigEnvy(cfg)
		fmt.Printf("Configuration is loaded %s attempting to init tags", cfg.TagFile)

		if cfg.UseDatabase {
			// The scope of transactions is a bit odd.  Seems like this is handled in
			// buffalo via the magical buffalo middleware
			fmt.Printf("DB Manager is being used, will write to postgres")
			return models.DB.Transaction(func(tx *pop.Connection) error {
				get_connection := func() *pop.Connection {
					return tx
				}
				man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
				tags, err := managers.CreateTagsFromFile(man)
				if tags != nil {
					fmt.Printf("Created a set of tags %d", len(*tags))
				}
				return err
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params, get_worker)
			fmt.Printf("Memory Manager looking for tags.\n")
			tags, err := managers.CreateTagsFromFile(man)
			if tags != nil {
				fmt.Printf("Created a set of tags %d\n", len(*tags))
			}
			return err
		}
	})
})

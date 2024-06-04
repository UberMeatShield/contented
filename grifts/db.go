package grifts

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"fmt"
	"net/url"
	"os"

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
				man := managers.CreateManager(cfg, get_connection, get_params)
				fmt.Printf("Creating previews images %t with db manager", man.CanEdit())
				return managers.CreateAllPreviews(man)
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params)
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
				man := managers.CreateManager(cfg, get_connection, get_params)
				fmt.Printf("Starting to encode %t TODO: Summary of config", man.CanEdit())
				return managers.EncodeVideos(man)
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params)
			fmt.Printf("Use memory manager %t for video encoding lookups", man.CanEdit())
			return managers.EncodeVideos(man)
		}
	})

	/*
	 * Attempt to remove duplicate videos that are already encoded (safely).
	 */
	grift.Add("removeDuplicates", func(c *grift.Context) error {
		cfg := utils.GetCfg()
		utils.InitConfigEnvy(cfg)
		fmt.Printf("Configuration is loaded %s Looking for duplicates", cfg.Dir)

		var dupes managers.DuplicateContents
		var err error
		if cfg.UseDatabase {
			fmt.Printf("DB Manager is being used trying to find duplicates.")
			models.DB.Transaction(func(tx *pop.Connection) error {
				get_connection := func() *pop.Connection {
					return tx
				}
				man := managers.CreateManager(cfg, get_connection, get_params)
				dupes, err = managers.FindDuplicateVideos(man)
				return nil
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params)
			fmt.Printf("Memory Manager looking for duplicates.\n")
			dupes, err = managers.FindDuplicateVideos(man)
		}

		if err != nil {
			return err
		}
		if len(dupes) > 0 {
			dupeFile := envy.Get("DUPE_FILE", "")
			if dupeFile != "" {
				fi, err := os.Create(dupeFile)
				if err != nil {
					fmt.Printf("Error creating duplicate output file %s", dupeFile)
					panic(err)
				}
				defer fi.Close()
				for _, dupe := range dupes {
					fi.WriteString(dupe.FqPath)
				}
				fmt.Printf("Wrote duplicate information to %s\n", dupeFile)
			} else {
				for _, dupe := range dupes {
					fmt.Printf("DUPLICATE FOUND %s", dupe.FqPath)
				}
			}
		} else {
			fmt.Printf("No Duplicates were found.")
		}
		return nil
	})

	// Adds a tag task but this does not tag the content itself
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
				man := managers.CreateManager(cfg, get_connection, get_params)
				tags, err := managers.CreateTagsFromFile(man)
				if tags != nil {
					fmt.Printf("Created a set of tags %d", len(*tags))
				}
				tErr := managers.AssignTagsAndUpdate(man, *tags)
				if tErr != nil {
					fmt.Printf("Failed to assign tags %s", tErr)
				}
				return err
			})
		} else {
			get_connection := no_connection
			man := managers.CreateManager(cfg, get_connection, get_params)
			fmt.Printf("Memory Manager looking for tags.\n")
			tags, err := managers.CreateTagsFromFile(man)
			if tags != nil {
				fmt.Printf("Created a set of tags %d\n", len(*tags))
			}
			return err
		}
	})
})

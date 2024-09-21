package main

import (
	"contented/pkg/managers"
	"contented/pkg/models"
	"contented/pkg/utils"
	"flag"
	"fmt"
	"net/url"
	"os"

	"gorm.io/gorm"
)

func main() {
	// Use the env setup (make it possible to pass env location?)
	actionFlag := flag.String("action", "help", "Directory where we search for content")
	flag.Parse()

	//dirDefault := utils.GetEnvString("DIR", "")
	//dirFlag := flag.String("dir", dirDefault, "Directory where we search for content")
	//fmt.Printf("Reading %s the directory paths %s", *actionFlag, *dirFlag)
	// A set of actions that require the DB operations
	action := *actionFlag
	switch action {
	case "reset":
		reset()
	case "populate":
		populate(CreateScriptManager())
	case "preview":
		preview(CreateScriptManager())
	case "encode":
		encode(CreateScriptManager())
	case "tags":
		tags(CreateScriptManager())
	case "duplicates":
		duplicates(CreateScriptManager())
	default:
		// TODO: Print the arg options
		fmt.Printf("Bit on the ugly side compared to grifts but less code")
	}
	// Should bail if things fail
}

func CreateScriptManager() managers.ContentManager {
	cfg := utils.GetCfgDefaults()
	utils.InitConfigEnvy(&cfg)
	utils.SetCfg(cfg)

	//TODO: Override with various flags and arguments
	get_params := func() *url.Values {
		return &url.Values{}
	}
	get_connection := func() *gorm.DB {
		return models.InitGorm(false)
	}
	return managers.CreateManager(&cfg, get_connection, get_params)
}

func reset() {
	models.RebuildDatabase("development")
}

func populate(man managers.ContentManager) error {
	cfg := man.GetCfg()
	// Clean out the current DB setup then builds a new one
	fmt.Printf("Manager setup and attempting to read from %s", cfg.Dir)
	return managers.CreateInitialStructure(cfg)
}

func preview(man managers.ContentManager) error {
	cfg := man.GetCfg()
	fmt.Printf("Creating previews under %s", cfg.Dir)
	return managers.CreateAllPreviews(man)
}

func encode(man managers.ContentManager) error {
	cfg := man.GetCfg()
	fmt.Printf("Encoding task started %s", cfg.Dir)
	return managers.EncodeVideos(man)
}

func tags(man managers.ContentManager) error {
	cfg := man.GetCfg()
	fmt.Printf("Attempting to create tags from file %s", cfg.TagFile)
	tags, err := managers.CreateTagsFromFile(man)
	if tags != nil {
		fmt.Printf("Created a set of tags %d", len(*tags))
	}
	tErr := managers.AssignTagsAndUpdate(man, *tags)
	if tErr != nil {
		fmt.Printf("Failed to assign tags %s", tErr)
	}
	return err
}

func duplicates(man managers.ContentManager) error {
	dupes, err := managers.FindDuplicateVideos(man)
	if len(dupes) > 0 {
		fmt.Printf("Found some duplicates")
		dupeFile := models.GetEnvString("DUPE_FILE", "./duplicates.txt")
		if dupeFile != "" {
			fi, err := os.Create(dupeFile)
			if err != nil {
				fmt.Printf("Error creating duplicate output file %s", dupeFile)
				panic(err)
			}
			defer fi.Close()
			for _, dupe := range dupes {
				fi.WriteString(fmt.Sprintf("%s\n", dupe.FqPath))
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
	if err != nil {
		fmt.Printf("There were errors finding dupes %s", err)
	}
	return err
}

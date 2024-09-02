package main

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"flag"
	"fmt"
	"net/url"

	"gorm.io/gorm"
)

func main() {

	//databaseFlag := flag.Bool("database", false, "Should we override the environment and use the DB")

	// Use the env setup (make it possible to pass env location?)
	dirDefault := utils.GetEnvString("DIR", "")
	actionFlag := flag.String("action", "help", "Directory where we search for content")
	dirFlag := flag.String("dir", dirDefault, "Directory where we search for content")
	flag.Parse()

	fmt.Printf("Reading %s the directory paths %s", *actionFlag, *dirFlag)

	// A set of actions that require the DB operations
	action := *actionFlag
	switch action {
	case "rebuild":
		rebuild()
	case "populate":
		cfg := utils.GetCfgDefaults()
		man := CreateScriptManager(&cfg)
		utils.SetCfg(cfg)
		utils.InitConfigEnvy(&cfg)
		populate(man)
	default:
		fmt.Printf("Bit on the ugly side")
	}

	// Should bail if things fail
}

func CreateScriptManager(cfg *utils.DirConfigEntry) managers.ContentManager {
	get_params := func() *url.Values {
		return &url.Values{}
	}
	get_connection := func() *gorm.DB {
		return models.InitGorm(false)
	}
	return managers.CreateManager(cfg, get_connection, get_params)
}

func rebuild() {
	models.RebuildDatabase("development")
}

func populate(man managers.ContentManager) error {

	cfg := man.GetCfg()
	// Clean out the current DB setup then builds a new one
	fmt.Printf("Manager setup and attempting to read from %s", cfg.Dir)
	return managers.CreateInitialStructure(cfg)
}

package main

import (
	"contented/pkg/actions"
	"contented/pkg/config"
	"contented/pkg/models"
	"log"

	"github.com/gin-gonic/gin"
)

/**
 * Initialize the Gin Application.
 */
func main() {
	cfg := config.GetCfg()
	config.InitConfigEnvy(cfg)
	if cfg.UseDatabase {
		models.InitializeAppDatabase(config.GetEnvString("GO_ENV", "development"))
	}

	// Set them up side by side?
	r := gin.Default()
	actions.GinApp(r)

	actions.SetupContented(r, "", 0, 0)
	if err := r.Run(); err != nil {
		log.Fatalf("Crashed out %s", err)
	}
}

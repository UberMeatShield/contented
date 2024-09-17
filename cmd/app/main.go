package main

import (
	"contented/pkg/actions"
	"contented/pkg/models"
	"contented/pkg/utils"
	"log"

	"github.com/gin-gonic/gin"
)

/**
 * Initialize the Gin Application.
 */
func main() {
	cfg := utils.GetCfg()
	utils.InitConfigEnvy(cfg)
	models.InitializeAppDatabase(utils.GetEnvString("GO_ENV", "development"))

	// Set them up side by side?
	r := gin.Default()
	actions.GinApp(r)

	actions.SetupContented(r, "", 0, 0)
	if err := r.Run(); err != nil {
		log.Fatalf("Crashed out %s", err)
	}
}

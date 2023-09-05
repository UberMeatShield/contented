package grifts

import (
	"contented/internals"
	"contented/utils"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)

// We do need a DB if you are going to run the grifts
func init() {
	cfg := utils.GetCfg()
	env := envy.Get("GO_ENV", "development")
	utils.InitConfigEnvy(cfg)
	buffalo.Grifts(internals.CreateBuffaloApp(cfg.UseDatabase, env))
}

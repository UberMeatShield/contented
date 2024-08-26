package main

import (
	"contented/actions"
	"contented/utils"
	"log"

	"github.com/gin-gonic/gin"
)

/**
 * Initialize the Gin Application.
 */
func main() {
	cfg := utils.GetCfg()
	utils.InitConfigEnvy(cfg)

	// Set them up side by side?
	r := gin.Default()
	actions.GinApp(r)

	actions.SetupContented(r, "", 0, 0)
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err := r.Run(); err != nil {
		log.Fatalf("Crashed out %s", err)
	}
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy
*/

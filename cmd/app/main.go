package main

import (
	"contented/actions"
	"contented/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	cfg := utils.GetCfg()
	utils.InitConfigEnvy(cfg)

	// Set them up side by side?
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})

	})

	r.LoadHTMLGlob(fmt.Sprintf("%s/*.html", cfg.StaticResourcePath))

	r.StaticFS("/public/build", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/css", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/static", http.Dir(cfg.StaticLibraryPath))
	//r.LoadHTMLGlob(fmt.Sprintf("%s/*", cfg.StaticResourcePath))
	actions.GinApp(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Printf("What hmmm shouldnt' this work")

	utils.InitConfigEnvy(cfg)
	app := actions.App(cfg.UseDatabase)

	// TODO: Update or delete this method as it is not really doing anything
	// Potentially just do the static hosting in the actions.App bit.
	actions.SetupContented(app, "", 0, 0)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}

}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/

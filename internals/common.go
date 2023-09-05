package internals

/**
 * These are some common test helper files dealing mostly with unit test setup and
 * mock data counts and information.
 */
import (
	"contented/models"
	"log"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"

	contenttype "github.com/gobuffalo/mw-contenttype"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/rs/cors"
)

// Create the basic app but without routes, useful for testing the managers but not routes
func CreateBuffaloApp(UseDatabase bool, env string) *buffalo.App {
	log.Printf("Calling CreateBuffaloApp with environment %s", env)
	app := buffalo.New(buffalo.Options{
		Env: env,
		PreWares: []buffalo.PreWare{
			cors.Default().Handler,
		},
		SessionName: "_content_session",
	})

	// Log request parameters (filters apply) this should maybe be only in dev
	app.Use(paramlogger.ParameterLogger)

	// Wraps each request in a transaction. Remove to disable this.
	if UseDatabase == true {
		log.Printf("Connecting to the database\n")
		app.Use(popmw.Transaction(models.DB))
	} else {
		log.Printf("This code will attempt to use memory management \n")
	}
	// Set the request content type to JSON
	app.Use(contenttype.Set("application/json"))
	return app
}

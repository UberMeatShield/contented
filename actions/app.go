package actions

import (
    "net/http"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/unrolled/secure"
	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	contenttype "github.com/gobuffalo/mw-contenttype"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"

	"contented/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_content_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Set the request content type to JSON
		app.Use(contenttype.Set("application/json"))

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

        app.GET("/content/", ListDefaultHandler)
        app.GET("/content/{dir_to_list}", ListSpecificHandler)

        // TODO: Just make the full path show the file directly?
        app.GET("/view/{dir_id}/{file_id}", ViewHandler)
        app.GET("/content/{dir_to_list}/{filename}", DownloadHandler)  // This needs to change
        app.GET("/download/{dir_to_list}/{filename}", DownloadHandler)

         // Host the index.html, also assume that all angular UI routes are going to be under contented
         app.GET("/", AngularIndex)
         app.GET("/ui/{path}", AngularIndex)
         app.GET("/ui/{path}/{idx}", AngularIndex)

         // Need to make the file serving location smarter (serve the dir + serve static?)
         app.ServeFiles("/public/build", http.Dir("public/build"))
         app.ServeFiles("/public/css", http.Dir("public/css"))

         // The DIR env environment is then served under /static (see actions.SetupContented)
         // app.ServeFiles("/static", http.Dir("X"))
	}

	return app
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

package actions

import (
    "log"
    "contented/models"
    "contented/utils"
    "net/http"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
    "github.com/gobuffalo/envy"
    contenttype "github.com/gobuffalo/mw-contenttype"
    forcessl "github.com/gobuffalo/mw-forcessl"
    paramlogger "github.com/gobuffalo/mw-paramlogger"
    "github.com/gobuffalo/x/sessions"
    "github.com/rs/cors"
    "github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// var T *i18n.Translator  // TODO: Add in internationalization

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
func App(UseDatabase bool) *buffalo.App {
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

        // Log request parameters (filters apply) this should maybe be only in dev
        app.Use(paramlogger.ParameterLogger)

        // Wraps each request in a transaction. Remove to disable this.
        //  c.Value("tx").(*pop.Connection)
        if UseDatabase == true {
            log.Printf("Connecting to the database\n")
            app.Use(popmw.Transaction(models.DB))
        } else {
            log.Printf("This code will attempt to use memory management \n")
        }

        // Set the request content type to JSON
        app.Use(contenttype.Set("application/json"))

        // Run grift?  Do dev from an actual DB instance?
        app.GET("/preview/{file_id}", PreviewHandler)
        app.GET("/view/{file_id}", FullHandler)
        app.GET("/download/{file_id}", DownloadHandler)

        // Host the index.html, also assume that all angular UI routes are going to be under contented
        app.GET("/", AngularIndex)
        app.GET("/ui/{path}", AngularIndex)
        app.GET("/ui/{path}/{idx}", AngularIndex)

        // Need to make the file serving location smarter (serve the dir + serve static?)
        cfg := utils.GetCfg()
        app.ServeFiles("/public/build", http.Dir(cfg.StaticResourcePath))
        app.ServeFiles("/public/css", http.Dir(cfg.StaticResourcePath))

        // The DIR env environment is then served under /static (see actions.SetupContented)
        cr := app.Resource("/containers", ContainersResource{})
        cr.Resource("/media", MediaContainersResource{})
        app.Resource("/media", MediaContainersResource{})
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

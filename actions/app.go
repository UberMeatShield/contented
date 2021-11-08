package actions

import (
    "contented/utils"
    "contented/internals"
    "net/http"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/envy"
    forcessl "github.com/gobuffalo/mw-forcessl"
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
        app = internals.CreateBuffaloApp(UseDatabase, ENV)
        app.Use(forceSSL())

        // Run grift?  Do dev from an actual DB instance?
        app.GET("/preview/{mc_id}", PreviewHandler)
        app.GET("/view/{mc_id}", FullHandler)
        app.GET("/download/{mc_id}", DownloadHandler)
        app.GET("/search", SearchHandler)

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

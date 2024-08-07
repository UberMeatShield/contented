package actions

import (
	"contented/internals"
	"contented/utils"
	"log"
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
	log.Printf("App being created()\n")
	if app == nil {
		app = internals.CreateBuffaloApp(UseDatabase, ENV)
		app.Use(forceSSL())

		// Need to move all the things under /api
		// TODO: Clean this up to always use content_id
		app.GET("/preview/{mcID}", PreviewHandler)
		app.GET("/view/{mcID}", FullHandler)
		app.GET("/download/{mcID}", DownloadHandler)
		app.GET("/api/search/contents", SearchHandler)
		app.GET("/api/search/containers", SearchContainersHandler)
		app.GET("/splash", SplashHandler)

		// Allow for manipulation of content already on the server
		app.POST("/editing_queue/{contentID}/screens/{count}/{startTimeSeconds}", ContentTaskScreensHandler)
		app.POST("/editing_queue/{contentID}/encoding", VideoEncodingHandler)
		app.POST("/editing_queue/{contentID}/webp", WebpFromScreensHandler)
		app.POST("/editing_queue/{contentID}/tagging", TaggingHandler)

		// TODO: Check that we can still kick off a duplicates task for the container.
		app.POST("/editing_container_queue/{containerID}/screens/{count}/{startTimeSeconds}", ContainerScreensHandler)
		app.POST("/editing_container_queue/{containerID}/encoding", ContainerVideoEncodingHandler)
		//app.POST("/editing_container_queue/{containerID}/webp", ContainerWebpHandler)
		app.POST("/editing_container_queue/{containerID}/tagging", ContainerTaggingHandler)

		// Dupes is a special case where we do not need to create a task per content type right now
		// the checks are fast enough that we can get a summary pretty quickly and probably faster with
		// single lookups (for now).
		app.POST("/editing_queue/{contentID}/duplicates", DupesHandler)
		app.POST("/editing_container_queue/{containerID}/duplicates", DupesHandler)

		// Allow for the creation of new content
		// app.POST("/uploading/contents/", TaskContentUploadHandler)
		// app.POST("/uploading/previews/{mcID}", TaskPreviewUploadHandler)
		// app.POST("/uploading/screen/{mcID}", TaskScreenUploadHandler)

		// The DIR env environment is then served under /static (see actions.SetupContented)
		cr := app.Resource("/containers", ContainersResource{})
		cr.Resource("/contents", ContentsResource{})

		mc_r := app.Resource("/contents", ContentsResource{})
		mc_r.Resource("/screens", ScreensResource{})
		app.Resource("/screens", ScreensResource{})
		app.Resource("/tags", TagsResource{})
		app.Resource("/task_requests", TaskRequestResource{})

		// Host the index.html, also assume that all angular UI routes are going to be under contented
		// Cannot figure out how to just let AngularIndex handle EVERYTHING under ui/*/*
		app.GET("/", AngularIndex)
		app.GET("/ui/browse/{path}", AngularIndex)
		app.GET("/ui/browse/{path}/{idx}", AngularIndex)
		app.GET("/ui/content/{id}", AngularIndex)
		app.GET("/ui/search", AngularIndex)
		app.GET("/ui/video", AngularIndex)
		app.GET("/ui/splash", AngularIndex)
		app.GET("/admin_ui/editor_content/{id}", AngularIndex)
		app.GET("/admin_ui/tasks", AngularIndex)
		app.GET("/admin_ui/containers", AngularIndex)
		app.GET("/admin_ui/search", AngularIndex)

		// Need to make the file serving location smarter (serve the dir + serve static?)
		cfg := utils.GetCfg()
		app.ServeFiles("/public/build", http.Dir(cfg.StaticResourcePath))
		app.ServeFiles("/public/css", http.Dir(cfg.StaticResourcePath))
		app.ServeFiles("/public/static", http.Dir(cfg.StaticLibraryPath))
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

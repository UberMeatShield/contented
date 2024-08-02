package actions

import (
	"contented/internals"
	"contented/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

		/*
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

		*/
	}
	return app
}

func GinApp(r *gin.Engine) {

	cfg := utils.GetCfg()
	// Provide the ability to load static resources
	r.LoadHTMLGlob(fmt.Sprintf("%s/*.html", cfg.StaticResourcePath))
	r.StaticFS("/public/build", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/css", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/static", http.Dir(cfg.StaticLibraryPath))

	// For LB pings etc.
	r.GET("/status", StatusHandler)

	// Search endpoints
	r.GET("/api/search/contents", SearchHandler)
	r.GET("/api/search/containers", SearchContainersHandler)

	// Downloads and
	r.GET("/api/preview/{mcID}", PreviewHandler)
	r.GET("/api/view/{mcID}", FullHandler)
	r.GET("/api/download/{mcID}", DownloadHandler)
	r.GET("/api/splash", SplashHandler)

	// Host the index.html, also assume that all angular UI routes are going to be under contented
	// Cannot figure out how to just let AngularIndex handle EVERYTHING under ui/*/*
	r.GET("/", AngularIndex)
	r.GET("/ui/browse/{path}", AngularIndex)
	r.GET("/ui/browse/{path}/{idx}", AngularIndex)
	r.GET("/ui/content/{id}", AngularIndex)
	r.GET("/ui/search", AngularIndex)
	r.GET("/ui/video", AngularIndex)
	r.GET("/ui/splash", AngularIndex)
	r.GET("/admin_ui/editor_content/{id}", AngularIndex)
	r.GET("/admin_ui/tasks", AngularIndex)
	r.GET("/admin_ui/containers", AngularIndex)
	r.GET("/admin_ui/search", AngularIndex)

	// CRUD
	// Containers
	r.GET("/api/containers", ContainersResourceList)
	r.GET("/api/containers/:container_id", ContainersResourceShow)
	r.GET("/api/containers/:container_id/contents", ContentsResourceList)
	r.POST("/api/containers", ContainersResourceCreate)
	r.PUT("/api/containers/:container_id", ContainersResourceUpdate)
	r.DELETE("/api/containers/:container_id", ContainersResourceDestroy)

	// Content API
	r.GET("/api/contents", ContentsResourceList)
	r.GET("/api/contents/:content_id", ContentsResourceShow)
	r.GET("/api/contents/:content_id/screens", ScreensResourceList)
	r.POST("/api/contents", ContentsResourceCreate)
	r.PUT("/api/contents/:content_id", ContentsResourceUpdate)
	r.DELETE("/api/contents/:content_id", ContentsResourceDestroy)

	// Screens
	r.GET("/api/screens", ScreensResourceList)
	r.GET("/api/screens/:screen_id", ScreensResourceShow)
	r.POST("/api/screens", ScreensResourceCreate)
	r.PUT("/api/screens/:screen_id", ScreensResourceUpdate)
	r.DELETE("/api/screens/:screen_id", ScreensResourceDestroy)

	// Tasks
	r.GET("/api/task_requests", TaskRequestsResourceList)
	r.GET("/api/task_requests/:task_request_id", TaskRequestsResourceShow)
	r.POST("/api/task_requests", TaskRequestsResourceCreate)
	r.PUT("/api/task_requests/:task_request_id", TaskRequestsResourceUpdate)
	r.DELETE("/api/task_requests/:screen_id", TaskRequestsResourceDestroy)

	// Tags
	r.GET("/api/tags", TagsResourceList)
	r.GET("/api/tags/:tag_id", TagsResourceShow)
	r.POST("/api/tags", TagsResourceCreate)
	r.PUT("/api/tags/:tag_id", TagsResourceUpdate)
	r.DELETE("/api/tags/:tag_id", TagsResourceDestroy)
	/*
		app.Resource("/tags", TagsResource{})
	*/
	// The DIR env environment is then served under /static (see actions.SetupContented)
	//cr := app.Resource("/containers", ContainersResource{})
	//cr.Resource("/contents", ContentsResource{})
	//mc_r := app.Resource("/contents", ContentsResource{})
	//mc_r.Resource("/screens", ScreensResource{})
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

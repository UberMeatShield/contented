package actions

import (
	"contented/pkg/config"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = config.GetEnvString("GO_ENV", "development")

// Setup the application
func GinApp(r *gin.Engine) {
	SetupRoutes(r)
	SetupStatic(r)
	SetupWorkers()
}

func SetupStatic(r *gin.Engine) {
	// Provide the ability to load static resources, this can be a little annoying as the serve
	// path in tests needs to be configured as ../ and running in prod needs just ./
	cfg := config.GetCfg()
	r.LoadHTMLGlob(fmt.Sprintf("%s/*.html", cfg.StaticResourcePath))
	r.StaticFS("/public/build", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/css", http.Dir(cfg.StaticResourcePath))
	r.StaticFS("/public/static", http.Dir(cfg.StaticLibraryPath))
}

func SetupRoutes(r *gin.Engine) {
	// For LB pings etc.
	r.GET("/status", StatusHandler)

	// Host the index.html, also assume that all angular UI routes are going to be under contented
	// Cannot figure out how to just let AngularIndex handle EVERYTHING under ui/*/*
	r.GET("/", AngularIndex)
	r.GET("/ui/browse/:path", AngularIndex)
	r.GET("/ui/browse/:path/:idx", AngularIndex)
	r.GET("/ui/content/:id", AngularIndex)
	r.GET("/ui/search", AngularIndex)
	r.GET("/ui/video", AngularIndex)
	r.GET("/ui/splash", AngularIndex)
	r.GET("/admin_ui/editor_content/:id", AngularIndex)
	r.GET("/admin_ui/tasks", AngularIndex)
	r.GET("/admin_ui/containers", AngularIndex)
	r.GET("/admin_ui/search", AngularIndex)

	// Search endpoints
	r.GET("/api/search/contents", SearchHandler)
	r.GET("/api/search/containers", SearchContainersHandler)

	// Downloads and basic rendering of content
	r.GET("/api/preview/:id", PreviewHandler)
	r.GET("/api/view/:id", FullHandler)
	r.GET("/api/download/:id", DownloadHandler)
	r.GET("/api/splash", SplashHandler)

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
	//r.GET("/api/contents/:content_id/tags", TagsResourceList) Needs updates in the ListAllTagsContext
	r.POST("/api/contents", ContentsResourceCreate)
	r.PUT("/api/contents/:content_id", ContentsResourceUpdate)
	r.DELETE("/api/contents/:content_id", ContentsResourceDestroy)
	r.DELETE("/api/contents/:content_id/screens", ContentScreensDestroy)

	// Screens
	r.GET("/api/screens", ScreensResourceList)
	r.GET("/api/screens/:screen_id", ScreensResourceShow)
	r.POST("/api/screens", ScreensResourceCreate)
	r.PUT("/api/screens/:screen_id", ScreensResourceUpdate)
	r.DELETE("/api/screens/:screen_id", ScreensResourceDestroy)

	// Tags
	r.GET("/api/tags", TagsResourceList)
	r.GET("/api/tags/:tag_id", TagsResourceShow)
	r.POST("/api/tags", TagsResourceCreate)
	r.PUT("/api/tags/:tag_id", TagsResourceUpdate)
	r.DELETE("/api/tags/:tag_id", TagsResourceDestroy)

	// Tasks
	r.GET("/api/task_requests", TaskRequestsResourceList)
	r.GET("/api/task_requests/:task_request_id", TaskRequestsResourceShow)
	r.POST("/api/task_requests", TaskRequestsResourceCreate)
	r.PUT("/api/task_requests/:task_request_id", TaskRequestsResourceUpdate)
	r.DELETE("/api/task_requests/:screen_id", TaskRequestsResourceDestroy)

	// Available tasks that can be added ot the system
	r.POST("/api/editing_queue/:content_id/screens/:count/:startTimeSeconds", ContentTaskScreensHandler)
	r.POST("/api/editing_queue/:content_id/encoding", VideoEncodingHandler)
	r.POST("/api/editing_queue/:content_id/webp", WebpFromScreensHandler)
	r.POST("/api/editing_queue/:content_id/tagging", TaggingHandler)
	r.POST("/api/editing_queue/:content_id/duplicates", DupesHandler)

	// TODO: Check that we can still kick off a duplicates task for the container.
	r.POST("/api/editing_container_queue/:container_id/screens/:count/:startTimeSeconds", ContainerScreensHandler)
	r.POST("/api/editing_container_queue/:container_id/encoding", ContainerVideoEncodingHandler)
	r.POST("/api/editing_container_queue/:container_id/tagging", ContainerTaggingHandler)
	r.POST("/api/editing_container_queue/:container_id/duplicates", DupesHandler)
	//TODO: app.POST("/editing_container_queue/{containerID}/webp", ContainerWebpHandler)
}

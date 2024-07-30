package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

type HttpError struct {
	Error string `json:"error"`
	Debug string `json:"debug"`
}

type SearchContentsResult struct {
	Total   int              `json:"total"`
	Results *models.Contents `json:"results"`
}

type SearchContainersResult struct {
	Total   int                `json:"total"`
	Results *models.Containers `json:"results"`
}

// Builds out information given the application and the content directory
func SetupContented(r *gin.Engine, contentDir string, numToPreview int, limit int) {
	cfg := utils.GetCfg()

	// Initialize workers that will listen for encoding tasks (GoBuffalo has some Gin does not)
	log.Printf("TODO: The job processors are busted without GoBuffalo")
	// TODO: SetupWorkers(app)

	// If we are not using databases load up the memory view
	if !cfg.UseDatabase {
		SetupMemory(cfg.Dir)
	}
}

func SetupMemory(dir string) {
	// Database we should assume that it should start loading memory
	if testing.Testing() {
		utils.InitializeMemory(dir)
	} else {
		go utils.InitializeMemory(dir)
	}
}

// TODO: Determine if these should be registered by config (don't use normal workers basically)
func SetupWorkers(app *buffalo.App) {
	cfg := utils.GetCfg()

	if cfg.StartQueueWorkers {
		w := app.Worker
		w.Register(models.TaskOperation.SCREENS.String(), ScreenCaptureWrapper)
		w.Register(models.TaskOperation.ENCODING.String(), VideoEncodingWrapper)
		w.Register(models.TaskOperation.WEBP.String(), WebpFromScreensWrapper)
		w.Register(models.TaskOperation.TAGGING.String(), TaggingContentWrapper)
		w.Register(models.TaskOperation.DUPES.String(), DuplicatesWrapper)
	}
}

func FullHandler(c *gin.Context) {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		c.AbortWithError(400, bad_uuid)
		return
	}
	man := managers.GetManager(c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	fq_path, fq_err := man.FindActualFile(mc)
	if fq_err != nil {
		log.Printf("File to full view not found on disk %s with err %s", fq_path, fq_err)
		c.AbortWithError(http.StatusUnprocessableEntity, fq_err)
		return
	}
	log.Printf("Full preview: %s for %s", fq_path, mc.ID.String())
	c.File(fq_path)
}

func SearchHandler(c *gin.Context) {
	man := managers.GetManager(c)
	mcs, count, err := man.SearchContentContext()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	// log.Printf("Search content returned %s", mcs)
	sr := SearchContentsResult{
		Results: mcs,
		Total:   count,
	}
	c.JSON(200, r.JSON(sr))
}

func SearchContainersHandler(c *gin.Context) {
	man := managers.GetManager(c)
	mcs, count, err := man.SearchContainersContext()
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	// log.Printf("Search content returned %s", mcs)
	// TODO: Hmmm, maybe it should always load the screens in a sane fashion?
	sr := SearchContainersResult{
		Results: mcs,
		Total:   count,
	}
	c.JSON(200, r.JSON(sr))
}

type SplashResponse struct {
	SplashTitle   string            `json:"splashTitle"`
	SplashContent string            `json:"splashContent"`
	RendererType  string            `json:"rendererType"`
	Content       *models.Content   `json:"content"`
	Container     *models.Container `json:"container"`
}

func SplashHandler(c *gin.Context) {
	cfg := utils.GetCfg()
	man := managers.GetManager(c)

	sr := SplashResponse{}
	if cfg.SplashContainerName != "" {
		cs := managers.ContainerQuery{Search: cfg.SplashContainerName, PerPage: 1, Page: 1, IncludeHidden: true}
		if cnts, _, err := man.SearchContainers(cs); err == nil {
			if cnts != nil && len(*cnts) == 1 {
				refs := *cnts // Ok, seriously why is the de-ref so annoying
				cnt := refs[0]

				// Limit the amount loaded for splash, could make it search based on render
				// type but that is pretty over optimized.
				contents, _, load_err := man.ListContent(managers.ContentQuery{ContainerID: cnt.ID.String(), PerPage: 100})
				if load_err == nil {
					cnt.Contents = *contents
				}
				sr.Container = &cnt
			}
		}
	}
	if cfg.SplashContentID != "" {
		log.Printf("It should look up %s", cfg.SplashContentID)
		mc_id, _ := uuid.FromString(cfg.SplashContentID)
		mc, err := man.GetContent(mc_id)
		if err == nil {
			sr.Content = mc
		}
	}
	if cfg.SplashHtmlFile != "" {
		splash, f_err := os.ReadFile(cfg.SplashHtmlFile)
		if f_err == nil {
			sr.SplashContent = string(splash)
		}
	}
	if cfg.SplashTitle != "" {
		sr.SplashTitle = cfg.SplashTitle
	}
	sr.RendererType = cfg.SplashRendererType
	c.JSON(200, r.JSON(sr))
}

// Find the preview of a file (if applicable currently it is just returning the full path)
func PreviewHandler(c *gin.Context) {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		c.AbortWithError(400, bad_uuid)
		return
	}

	man := managers.GetManager(c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	fq_path, fq_err := man.GetPreviewForMC(mc)
	if fq_err != nil {
		log.Printf("File to preview not found on disk %s with err %s", fq_path, fq_err)
		c.AbortWithError(http.StatusUnprocessableEntity, fq_err)
		return
	}
	log.Printf("Found this preview filename to view: %s for %s", fq_path, mc.ID.String())
	c.File(fq_path)
}

// Provides a download handler by directory id and file id
func DownloadHandler(c *gin.Context) {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		c.AbortWithError(400, bad_uuid)
		return
	}
	man := managers.GetManager(c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	// Some content is not actually real
	if mc.NoFile == true {
		c.JSON(200, r.JSON(mc))
		return
	}
	fq_path, fq_err := man.FindActualFile(mc)
	if fq_err != nil {
		log.Printf("Cannot download file not on disk %s with err %s", fq_path, fq_err)
		c.AbortWithError(http.StatusUnprocessableEntity, fq_err)
		return
	}
	finfo, _ := os.Stat(fq_path)
	c.FileAttachment(fq_path, finfo.Name())
}

// This was the code provided to look up params... this seems cumbersome but "eh?"
func GetKeyVal(c *gin.Context, key string, defaultVal string) string {
	val := c.Request.URL.Query().Get(key)
	if val != "" {
		return val
	}
	return defaultVal
}

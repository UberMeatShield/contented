package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

type HttpError struct {
	Error string `json:"error"`
	Debug string `json:"debug"`
}

type SearchResult struct {
	Total   int              `json:"total"`
	Results *models.Contents `json:"results"`
}

// Builds out information given the application and the content directory
func SetupContented(app *buffalo.App, contentDir string, numToPreview int, limit int) {
	cfg := utils.GetCfg()

	// TODO: Check DIR exists
	// TODO: Somehow need to move the dir into App, but first we want to validate the dir...
	app.ServeFiles("/static", http.Dir(cfg.Dir))

	// TODO: When should this get setup
	SetupWorkers(app)
}

// TODO: Determine if these should be registered by config (don't use normal workers basically)
func SetupWorkers(app *buffalo.App) {
	w := app.Worker
	w.Register(models.TaskOperation.SCREENS.String(), ScreenCaptureWrapper)
	w.Register(models.TaskOperation.ENCODING.String(), VideoEncodingWrapper)
	w.Register(models.TaskOperation.WEBP.String(), WebpFromScreensWrapper)
}

func FullHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	man := managers.GetManager(&c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		return c.Error(404, err)
	}
	fq_path, fq_err := man.FindActualFile(mc)
	if fq_err != nil {
		log.Printf("File to full view not found on disk %s with err %s", fq_path, fq_err)
		return c.Error(http.StatusUnprocessableEntity, fq_err)
	}
	log.Printf("Full preview: %s for %s", fq_path, mc.ID.String())
	http.ServeFile(c.Response(), c.Request(), fq_path)
	return nil
}

func SearchHandler(c buffalo.Context) error {
	man := managers.GetManager(&c)
	mcs, count, err := man.SearchContentContext()
	if err != nil {
		return c.Error(400, err)
	}

	// log.Printf("Search content returned %s", mcs)
	// TODO: Hmmm, maybe it should always load the screens in a sane fashion?
	sr := SearchResult{
		Results: mcs,
		Total:   count,
	}
	return c.Render(200, r.JSON(sr))
}

type SplashResponse struct {
	SplashTitle   string            `json:"splashTitle"`
	SplashContent string            `json:"splashContent"`
	RendererType  string            `json:"rendererType"`
	Content       *models.Content   `json:"content"`
	Container     *models.Container `json:"container"`
}

func SplashHandler(c buffalo.Context) error {
	cfg := utils.GetCfg()
	man := managers.GetManager(&c)

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

	return c.Render(200, r.JSON(sr))
}

// Find the preview of a file (if applicable currently it is just returning the full path)
func PreviewHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}

	man := managers.GetManager(&c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		return c.Error(404, err)
	}

	fq_path, fq_err := man.GetPreviewForMC(mc)
	if fq_err != nil {
		log.Printf("File to preview not found on disk %s with err %s", fq_path, fq_err)
		return c.Error(http.StatusUnprocessableEntity, fq_err)
	}
	log.Printf("Found this preview filename to view: %s for %s", fq_path, mc.ID.String())
	http.ServeFile(c.Response(), c.Request(), fq_path)
	return nil
}

// Provides a download handler by directory id and file id
func DownloadHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	man := managers.GetManager(&c)
	mc, err := man.GetContent(mcID)
	if err != nil {
		return c.Error(404, err)
	}
	// Some content is not actually real
	if mc.NoFile == true {
		return c.Render(200, r.JSON(mc))
	}
	fq_path, fq_err := man.FindActualFile(mc)
	if fq_err != nil {
		log.Printf("Cannot download file not on disk %s with err %s", fq_path, fq_err)
		return c.Error(http.StatusUnprocessableEntity, fq_err)
	}
	finfo, _ := os.Stat(fq_path)
	file_contents := utils.GetFileContentsByFqName(fq_path)
	return c.Render(200, r.Download(c, finfo.Name(), file_contents))
}

// This was the code provided to look up params... this seems cumbersome but "eh?"
func GetKeyVal(c buffalo.Context, key string, defaultVal string) string {
	if m, ok := c.Params().(url.Values); ok {
		for k, v := range m {
			if k == key && v != nil {
				return v[0]
			}
		}
	}
	return defaultVal
}

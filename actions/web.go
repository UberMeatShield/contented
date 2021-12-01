package actions

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"contented/utils"
	"contented/managers"
	"contented/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

type HttpError struct {
	Error string `json:"error"`
	Debug string `json:"debug"`
}

type SearchResult struct {
    Total int `json:"total"`
    Media *models.MediaContainers `json:"media"`
}

// Builds out information given the application and the content directory
func SetupContented(app *buffalo.App, contentDir string, numToPreview int, limit int) {
    cfg := utils.GetCfg()

    // TODO: Check DIR exists
	// TODO: Somehow need to move the dir into App, but first we want to validate the dir...
	app.ServeFiles("/static", http.Dir(cfg.Dir))
}

func FullHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
    man := managers.GetManager(&c)
	mc, err := man.FindFileRef(mcID)
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
    mcs, count, err := man.SearchMediaContext()
    if err != nil {
		return c.Error(400, err)
    }
    sr := SearchResult{
        Media: mcs,
        Total: count,
    }
    return c.Render(200, r.JSON(sr))
}

// Find the preview of a file (if applicable currently it is just returning the full path)
func PreviewHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
    man := managers.GetManager(&c)
	mc, err := man.FindFileRef(mcID)
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
    mc, err := man.FindFileRef(mcID)
    if err != nil {
        return c.Error(404, err)
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


package actions

import (
	"contented/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

type PreviewResults struct {
	Success bool                `json:"success"`
	Results []utils.DirContents `json:"results"`
}

type HttpError struct {
	Error string `json:"error"`
	Debug string `json:"debug"`
}

// Builds out information given the application and the content directory
func SetupContented(app *buffalo.App, contentDir string, numToPreview int, limit int) {
	if !strings.HasSuffix(contentDir, "/") {
		contentDir = contentDir + "/"
	}
	log.Printf("Setting up the content directory with %s", contentDir)

	utils.InitConfig(contentDir, &appCfg)

    SetupManager(&appCfg)
    
	appCfg.PreviewCount = numToPreview
	appCfg.Limit = limit

	// TODO: Somehow need to move the dir into App, but first we want to validate the dir...
	app.ServeFiles("/static", http.Dir(appCfg.Dir))
}


func SetupManager(cfg *utils.DirConfigEntry) ContentManager {
    if cfg.UseDatabase {
	    log.Printf("Setting up the DB Manager")
        db_man := ContentManagerDB{cfg: cfg}
        SetManager(db_man)
    } else {
	    log.Printf("Setting up the memory Manager")
        mem_man := ContentManagerMemory{cfg: cfg}
        mem_man.Initialize()
        SetManager(mem_man)
    }
    return GetManager()
}

func FullHandler(c buffalo.Context) error {
	file_id, bad_uuid := uuid.FromString(c.Param("file_id"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
    SetContext(c)
    man := GetManager()
	mc, err := man.FindFileRef(file_id)
	if err != nil {
		return c.Error(404, err)
	}
	fq_path, fq_err := FindActualFile(mc)
	if fq_err != nil {
		log.Printf("File to full view not found on disk %s with err %s", fq_path, fq_err)
		return c.Error(http.StatusUnprocessableEntity, fq_err)
	}
	log.Printf("Full preview: %s for %s", fq_path, mc.ID.String())
	http.ServeFile(c.Response(), c.Request(), fq_path)
	return nil
}

// Find the preview of a file (if applicable currently it is just returning the full path)
func PreviewHandler(c buffalo.Context) error {
    SetContext(c)
	file_id, bad_uuid := uuid.FromString(c.Param("file_id"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
    man := GetManager()
	mc, err := man.FindFileRef(file_id)
	if err != nil {
		return c.Error(404, err)
	}

	fq_path, fq_err := GetPreviewForMC(mc)
	if fq_err != nil {
		log.Printf("File to preview not found on disk %s with err %s", fq_path, fq_err)
		return c.Error(http.StatusUnprocessableEntity, fq_err)
	}
	log.Printf("Found this pReview filename to view: %s for %s", fq_path, mc.ID.String())
	http.ServeFile(c.Response(), c.Request(), fq_path)
	return nil
}

// Provides a download handler by directory id and file id
func DownloadHandler(c buffalo.Context) error {
    file_id, bad_uuid := uuid.FromString(c.Param("file_id"))
    if bad_uuid != nil {
        return c.Error(400, bad_uuid)
    }

    SetContext(c)
    man := GetManager()
    mc, err := man.FindFileRef(file_id)
    if err != nil {
        return c.Error(404, err)
    }
    fq_path, fq_err := FindActualFile(mc)
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


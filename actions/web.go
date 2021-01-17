package actions

import (
	"contented/models"
	"contented/utils"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
        db_man := ContentManagerDB{cfg: cfg}
        SetManager(db_man)
    } else {
        mem_man := ContentManagerMemory{cfg: cfg}
        mem_man.Initialize()
        SetManager(mem_man)
    }
    return GetManager()
}


func ListDefaultHandler(c buffalo.Context) error {
	path, _ := os.Executable()
	log.Printf("Calling into ListDefault run_dir: %s looking at dir: %s", path, appCfg.Dir)
	response := PreviewResults{
		Success: true,
		Results: utils.ListDirs(appCfg.Dir, appCfg.PreviewCount),
	}
	return c.Render(200, r.JSON(response))
}

// Definitely should just make a hash lookup of dirname => dir Obj and dir_id => dir Obj
func isValidDir(dir_id string) bool {
	if _, ok := appCfg.ValidDirs[dir_id]; ok {
		return true
	}
	return false
}

// TODO: Move all this into utils?
// Only a file info, seemingly there is no way to further list from this (aka look ad dir contents)
func getDir(dir_id string) (os.FileInfo, error) {
	if dir, ok := appCfg.ValidDirs[dir_id]; ok {
		return dir, nil
	}
	return nil, errors.New("Directory not found: " + dir_id)
}

// Should hash the lookup with actual directory objects (but perhaps without contents)
func getDirName(dir_id string) (string, error) {
	dir, err := getDir(dir_id)
	if err == nil {
		return dir.Name(), nil
	}
	return "", err
}

// Helper for getting the current file info
func getFileInfo(dir_id string, file_id string) (os.FileInfo, error) {
	dir_name, err := getDirName(dir_id)
	if err == nil {
		return utils.GetFileRefById(appCfg.Dir+dir_name, file_id)
	}
	return nil, err
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

// Just a temp (nuke asap)
func CacheFile(mc models.MediaContainer) {
	appCfg.ValidFiles[mc.ID] = mc
}

// This seems to be a bit cleaner (Deprecate in favor of container loads)
func getFullFilePath(dir_id string, file_id string) (string, error) {
	log.Printf("Searching dir_id(%s) and file_id(%s)", dir_id, file_id)
	dir_name, d_err := getDirName(dir_id)
	if d_err != nil {
		return "", d_err
	}
	file_ref, err := getFileInfo(dir_id, file_id)
	if err != nil {
		return "", err
	}
	fname := filepath.Join(appCfg.Dir, dir_name, file_ref.Name())
	log.Printf("dir_id(%s) and file_id(%s) this directory name: %s", dir_id, file_id, fname)
	return fname, nil
}

// Provides a view of the file (will not open as an attachment)  TODO: Convert to uuid version
func ViewHandler(c buffalo.Context) error {
	dir_id := c.Param("dir_id")
	file_id := c.Param("file_id")

	fname, err := getFullFilePath(dir_id, file_id)
	if err == nil {
		log.Printf("Found this filename to view: %s", fname)
		http.ServeFile(c.Response(), c.Request(), fname)
		return nil
	}
	log.Printf("Failed to find the file reference  %s", err)
	return c.Error(404, err)
}

// Provide a full listing of a specific directory, not just the preview
// TODO: convert to directory ID or name (Make it smarter)
func ListSpecificHandler(c buffalo.Context) error {
	dir_id := c.Param("dir_id")

	// Pull out the limit and offset queries, provides pagination
	limit := DefaultLimit
	offset := 0

	limit, _ = strconv.Atoi(GetKeyVal(c, "limit", strconv.Itoa(DefaultLimit)))
	if limit <= 0 || limit > DefaultLimit {
		limit = DefaultLimit // Still cannot ask for more than the startup specified
	}
	offset, _ = strconv.Atoi(GetKeyVal(c, "offset", "0"))

	log.Printf("Limit %d with offset %d in dir %s", limit, offset, appCfg.Dir)

	// Now actually return the results for a valid directory
	if isValidDir(dir_id) {
		contents, err := getDirectory(appCfg.Dir, dir_id, limit, offset)
		if err == nil {
			return c.Render(200, r.JSON(contents))
		}
		return c.Error(404, err)
	}
	return c.Render(403, r.JSON(invalidDirMsg(dir_id, "")))
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

/**
 * Get the response for a single specific directory
 */
func getDirectory(rootDir string, dir_id string, limit int, offset int) (utils.DirContents, error) {
	// TODO: Do a lookup based on dir ID?
	dir_name, err := getDirName(dir_id)
	if err == nil {
		fq_dirname := filepath.Join(appCfg.Dir, dir_name)
		log.Printf("Loading up all the contents in %s", fq_dirname)
		return utils.GetDirContents(fq_dirname, limit, offset, dir_id), nil
	}
	return utils.DirContents{}, errors.New("This directory was not find")
}

// TODO: Make this a method that does the writting & just takes debug data
func invalidDirMsg(directory string, filename string) HttpError {
	err := HttpError{
		Error: "This is not a valid directory: " + directory + " " + filename,
		Debug: "Not in valid dirs",
	}
	return err
}

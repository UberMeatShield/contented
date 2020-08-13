package actions

import (
  "os"
  //"fmt"
  "errors"
  "strings"
  "strconv"
  "log"
  "path/filepath"
  "net/url"
  "net/http"
  "contented/utils"
  "github.com/gobuffalo/buffalo"
)

type PreviewResults struct{
    Success bool `json:"success"`
    Results []utils.DirContents `json:"results"`
}

type HttpError struct{
    Error string `json:"error"`
    Debug string `json:"debug"`
}

// TODO: this might be useful to add into the utils
type DirConfigEntry struct{
  Dir string  // The root of our loading (path to top level container directory)
  ValidDirs map[string]os.FileInfo  // List of directories under the main element
  PreviewCount int  // How many files should be listed for a preview
  Limit int // The absolute max you can load in a single operation
}

// HomeHandler is a default handler to serve up
var DefaultLimit int = 10000  // The max limit set by environment variable
var DefaultPreviewCount int = 8
var cfg = DirConfigEntry{
    Dir: "",
    PreviewCount: DefaultPreviewCount,
    Limit: DefaultLimit, 
}

// Builds out information given the application and the content directory
func SetupContented(app *buffalo.App, contentDir string, numToPreview int, limit int) {
    if !strings.HasSuffix(contentDir, "/") {
         contentDir = contentDir + "/"
    }
    log.Printf("Setting up the content directory with %s", contentDir)

    cfg.Dir = contentDir
    cfg.ValidDirs = utils.GetDirectoriesLookup(cfg.Dir)
    cfg.PreviewCount = numToPreview
    cfg.Limit = limit

    // TODO: Somehow need to move the dir into App, but first we want to validate the dir...
    app.ServeFiles("/static", http.Dir(cfg.Dir))
}

func ListDefaultHandler(c buffalo.Context) error {
    path, _ := os.Executable()
    log.Printf("Calling into ListDefault run_dir: %s looking at dir: %s", path, cfg.Dir)
    response := PreviewResults{
        Success: true,
        Results: utils.ListDirs(cfg.Dir, cfg.PreviewCount),
    }
    return c.Render(200, r.JSON(response))
}

// Definitely should just make a hash lookup of dirname => dir Obj and dir_id => dir Obj
func isValidDir(dir_id string) bool {
    if _, ok := cfg.ValidDirs[dir_id]; ok {
        return true
    }
    return false
}

// TODO: Move all this into utils?
// Only a file info, seemingly there is no way to further list from this (aka look ad dir contents)
func getDir(dir_id string) (os.FileInfo, error) {
    if dir, ok := cfg.ValidDirs[dir_id]; ok {
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
        return  utils.GetFileRefById(cfg.Dir + dir_name, file_id)
    }
    return nil, err
}

// This seems to be a bit cleaner
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
    fname := filepath.Join(cfg.Dir, dir_name, file_ref.Name())
    log.Printf("dir_id(%s) and file_id(%s) this directory name: %s", dir_id, file_id, fname)
    return fname, nil
}

// Provides a view of the file (will not open as an attachment)
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

// Provides a download handler by directory id and file id
func DownloadHandler(c buffalo.Context) error {
    dir_id := c.Param("dir_id")  // This can be the current directory or directory name
    file_id := c.Param("file_id")

    fname, err := getFullFilePath(dir_id, file_id)
    if err == nil {
        file_ref, f_err := getFileInfo(dir_id, file_id)
        if f_err == nil {
            log.Printf("Providing a download to this filename  %s", fname)
            file_contents := utils.GetFileContentsByFqName(fname)
            return c.Render(200, r.Download(c, file_ref.Name(), file_contents))
        }
        log.Printf("Failed to find the file reference for dir_id(%s) and file_id(%s) err %s", dir_id, file_id, f_err)
        return c.Error(404, f_err)
    }
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

    log.Printf("Limit %d with offset %d in dir %s", limit, offset, cfg.Dir)

    // Now actually return the results for a valid directory
    if isValidDir(dir_id) {
        contents, err := getDirectory(cfg.Dir, dir_id, limit, offset)
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
    if err == nil{
        fq_dirname := filepath.Join(cfg.Dir, dir_name)
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

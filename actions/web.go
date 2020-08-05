package actions

import (
  "os"
  "fmt"
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

type DirConfigEntry struct{
  Dir string
  ValidDirs map[string]string
  PreviewCount int
  Limit int
}

// HomeHandler is a default handler to serve up
var DefaultLimit int = 10000  // The max limit set by environment variable
var DefaultPreviewCount int = 8
var cfg = DirConfigEntry{
    Dir: "",
    PreviewCount: DefaultPreviewCount,
    Limit: DefaultLimit,
}

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

func getDirName(dir_id string) string {
    if val, ok := cfg.ValidDirs[dir_id]; ok {
        return val
    }
    return ""
}

func ViewHandler(c buffalo.Context) error {
    dir_id := c.Param("dir_id")
    file_id := c.Param("file_id")

    valid_dir := isValidDir(dir_id)
    dir_to_list := ""
    if valid_dir {
        dir_to_list = getDirName(dir_id)
    }

    log.Printf("Calling into view handler with filename %s under %s", dir_to_list, file_id)
    if valid_dir {
        file_ref, err := utils.GetFileRefById(cfg.Dir + dir_to_list, file_id)
        if err != nil {
            return c.Error(404, err)
        }
        if file_ref == nil {
            return c.Error(404, fmt.Errorf("File was not found with id %s", file_id))
        }
        fname := filepath.Join(cfg.Dir, dir_to_list, file_ref.Name())
        log.Printf("Found this filename: %s", file_ref.Name())
        http.ServeFile(c.Response(), c.Request(), fname)
        return nil
    } else {
        return c.Render(404, r.JSON(invalidDirMsg(dir_to_list, file_id)))
    }
}

func DownloadHandler(c buffalo.Context) error {
    dir_to_list := c.Param("dir_to_list")
    filename := c.Param("filename")
    log.Printf("Calling into download handler with filename %s under %s", dir_to_list, filename)

    if isValidDir(dir_to_list) {
        fileref := utils.GetFileContents(cfg.Dir + dir_to_list, filename)
        if fileref != nil {
            return c.Render(200, r.Download(c, filename, fileref))
        } 
    } 
    return c.Render(403, r.JSON(invalidDirMsg(dir_to_list, filename)))
}


// Provide a full listing of a specific directory, not just the preview
func ListSpecificHandler(c buffalo.Context) error {
    dirId := c.Param("dir_to_list")

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
    if isValidDir(dirId) {
        return c.Render(200, r.JSON(getDirectory(cfg.Dir, dirId, limit, offset)))
    } 
    return c.Render(403, r.JSON(invalidDirMsg(dirId, "")))
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
func getDirectory(rootDir string, dirId string, limit int, offset int) utils.DirContents {
    // TODO: Do a lookup based on dir ID?
    fqPathToDir := rootDir + dirId
    return utils.GetDirContents(fqPathToDir, limit, offset, dirId)
}

// TODO: Make this a method that does the writting & just takes debug data
func invalidDirMsg(directory string, filename string) HttpError {
    err := HttpError{
        Error: "This is not a valid directory: " + directory + " " + filename,
        Debug: "Not in valid dirs",
    }
    return err
}

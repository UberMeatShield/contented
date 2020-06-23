package actions

import (
    //"io"
    //"fmt"
//    "url"
    "strconv"
    "log"
//    "encoding/json"
//    "net/http"
    "net/url"
    "net/http"
//    "io/ioutil"
    "contentedutils"
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

// HomeHandler is a default handler to serve up
// a home page.
var dir string
var validDirs map[string]bool
var previewCount int
var DefaultLimit int = 10000

func SetupContented(app *buffalo.App, contentDir string, numToPreview int, limit int) {
    dir = contentDir
    validDirs = utils.GetDirectoriesLookup(dir)
    previewCount = numToPreview
    DefaultLimit = limit

    app.ServeFiles("/static", http.Dir(dir))
}

func ListDefaultHandler(c buffalo.Context) error {
    log.Println("Calling into ListDefault")
    response := PreviewResults{
        Success: true,
        Results: utils.ListDirs(dir, previewCount),
    }
    return c.Render(200, r.JSON(response))
}

func DownloadHandler(c buffalo.Context) error {
    log.Println("Calling into download handler")

    dir_to_list := c.Param("dir_to_list")
    filename := c.Param("filename")

    if validDirs[dir_to_list] {
        fileref := utils.GetFileContents(dir + dir_to_list, filename)
        if fileref != nil {
            return c.Render(200, r.Download(c, filename, fileref))
        } 
    } 
    return c.Render(403, r.JSON(invalidDirMsg(dir_to_list, filename)))
}


// Provide a full listing of a specific directory, not just the preview
func ListSpecificHandler(c buffalo.Context) error {
    argument := c.Param("dir_to_list")

    // Pull out the limit and offset queries, provides pagination
    limit := DefaultLimit
    offset := 0

    limit, _ = strconv.Atoi(GetKeyVal(c, "limit", strconv.Itoa(DefaultLimit)))
    if limit <= 0 || limit > DefaultLimit {
        limit = DefaultLimit // Still cannot ask for more than the startup specified
    }
    offset, _ = strconv.Atoi(GetKeyVal(c, "offset", "0"))

    log.Printf("Limit %d with offset %d in dir %s", limit, offset, dir)

    // Now actually return the results for a valid directory
    if validDirs[argument] {
        return c.Render(200, r.JSON(getDirectory(dir, argument, limit, offset)))
    } 
    return c.Render(403, r.JSON(invalidDirMsg(argument, "")))
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
func getDirectory(dir string, argument string, limit int, offset int) utils.DirContents {
    path := dir + argument
    return utils.GetDirContents(path, limit, offset, dir)
}

// TODO: Make this a method that does the writting & just takes debug data
func invalidDirMsg(directory string, filename string) HttpError {
    err := HttpError{
        Error: "This is not a valid directory: " + directory + " " + filename,
        Debug: "Not in valid dirs",
    }
    return err
}

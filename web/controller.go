package web

import (
    "fmt"
    "log"
    "encoding/json"
    "net/http"
    "io/ioutil"
    "github.com/gorilla/mux"
    "contented/utils"
)

type PreviewResults struct{
    Success bool `json:"success"`
    Results []utils.DirContents `json:"results"`
}

type HttpError struct{
    Error string `json:"error"`
    Debug string `json:"debug"`
}

var dir string
var validDirs map[string]bool

func SetupContented(router *mux.Router, contentDir string) {
    dir = contentDir
    validDirs = utils.GetDirectoriesLookup(dir)

    router.PathPrefix("/contented/").Handler(http.StripPrefix("/contented/", http.FileServer(http.Dir(dir))))
    router.HandleFunc("/content/", ListDefaultHandler)
    router.HandleFunc("/content/{dir_to_list}", ListSpecificHandler)

    // Host the index.html, also assume that all angular UI routes are going to be under contented
    router.HandleFunc("/", Index)
    router.HandleFunc("/ui/{path}", Index)
}


func SetupStatic(router *mux.Router, staticDir string) {
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
}


// Replace this with nginx or something else better at serving static content (probably)
func Index(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadFile("./static/build/index.html")

    // Try to keep the same amount of headers
    w.Header().Set("Server", "gophr")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Content-Type", "text/html")

    if err != nil {
        err_msg := "Could not find index.html" + err.Error()
        w.Header().Set("Content-Length", fmt.Sprint(len(err_msg)))
        fmt.Fprint(w, err_msg)
    } else {
        output := string(body)
        w.Header().Set("Content-Length", fmt.Sprint(len(output)))
        fmt.Fprint(w, output)
    }
}

func ListDefaultHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Calling into ListDefault")
    w.Header().Set("Content-Type", "application/json")
    response := PreviewResults{
		Success: true,
		Results: utils.ListDirs(dir, 4),
    }
    j, _ := json.Marshal(response)
    w.Write(j)
}

// argument := r.URL.Query().Get("dir_to_list")
func ListSpecificHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    vars := mux.Vars(r)
    argument := vars["dir_to_list"]
    if validDirs[argument] {
        j, _ := json.Marshal(getDirectory(dir, argument))
        w.Write(j)
    } else {
		err := HttpError{
			Error: "This is not a valid directory: " + argument,
			Debug: "Not in valid dirs",
		}
        j, _ := json.Marshal(err)
		w.WriteHeader(403)
		w.Write(j)
    }
}

/**
 * Get the response for a single specific directory
 */
func getDirectory(dir string, argument string) utils.DirContents {
	path := dir + argument
    return utils.GetDirContents(path, 1000, dir)
}

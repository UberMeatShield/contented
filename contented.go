package main

import (
    "fmt"
    "time"
    "flag"
    "log"
    "encoding/json"
    "net/http"
    "io/ioutil"
    "gorilla/mux"
)

type HttpStdResults struct{
    Success bool `json:"success"`
    Results []string `json:"results"`
    Path string `json:"path"`
}

type HttpErrResults struct{
    Error string `json:"error"`
    Debug string `json:"debug"`
}


var dir string
var validDirs map[string]bool

func main() {
    flag.StringVar(&dir, "dir", ".", "Directory to serve files from")
    flag.Parse()

    fmt.Println("Using this directory As the static root: ", dir)

    router := mux.NewRouter()
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
    router.HandleFunc("/content/", ListDefaultHandler)
    router.HandleFunc("/content/{dir_to_list}", ListSpecificHandler)

	// Host the index.html, also assume that all angular UI routes are going to be under contented
    router.HandleFunc("/", Index)
    router.HandleFunc("/contented/{path}", Index)

    validDirs = getDirectoriesLookup(dir)
    srv := &http.Server{
        Handler:      router,
        Addr:         "127.0.0.1:8000",
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }
    log.Fatal(srv.ListenAndServe())
    http.Handle("/", router)
}


// Replace this with nginx or something else better at serving static content (probably)
func Index(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadFile("./static/app.html")

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
    j, _ := json.Marshal(ListDirs(dir, 3))
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
        j, _ := json.Marshal(map[string]string{"error": "This is not a valid directory: " + argument})
        w.Write(j)
    }
}

/**
 * Get the response for a single specific directory
 */
func getDirectory(dir string, argument string) HttpStdResults {
    path := dir + argument
    response := HttpStdResults{
        true,
        GetDirContents(path, 1000),
        "static/" + argument + "/",
    }
    return response
}

/**
 *  Check if a directory is a legal thing to view
 */
func getDirectoriesLookup(legal string) map[string]bool {
    var listings = make(map[string]bool)
    files, _ := ioutil.ReadDir(legal)
    for _, f := range files {
        if f.IsDir() {
            listings[f.Name()] = true
        }
    }
    return listings
}

/**
 * Grab a small preview list of all items in the directory.
 */
func ListDirs(dir string, previewCount int) map[string][]string {
    var listings = make(map[string][]string)
    files, _ := ioutil.ReadDir(dir)
    for _, f := range files {
        if f.IsDir() {
            listings[f.Name()] = GetDirContents(dir + f.Name(), previewCount)
        }
    }
    return listings
}

func GetDirContents(dir string, limit int) []string {
    var arr = []string{}
    imgs, _ := ioutil.ReadDir(dir)

    for _, img := range imgs {
        if !img.IsDir() && len(arr) < limit + 1 {
            arr = append(arr, img.Name())
        }
    }
    return arr
}

package main

import (
    "fmt"
    "time"
    "flag"
    "log"
    "net/http"
    "testing"
    "github.com/gorilla/mux"
    "contented/unit"
    "contented/web"
)

func main() {
    var dir string
    var port string = "8000"

    flag.StringVar(&dir, "dir", ".", "Directory to serve files from")
    flag.StringVar(&port, "port", "8000", "Port to run the webserver.")
    previewCount := flag.Int("previewCount", 8, "Number of refrences to return by default")
    limitCount := flag.Int("limit", web.DefaultLimit, "Max default items returned")

    test := flag.Bool("test", false, "Instead of running a server, test the http calls against a running server")
    flag.Parse()

    server_url := "127.0.0.1:" + port
    if *test == true { // Obviously requires a server up and running on another process
        unit_test(server_url)
    } else {
        server(server_url, dir, *previewCount, *limitCount)
    }

}

func unit_test(server_url string) {
    fmt.Println("Run the unit tests instead")
    t := testing.T{}
    unit.IntegrationLoad(&t)
}


func server(server_url string, dir string, previewCount int, limitCount int) {
    fmt.Printf("Using this directory As the static root: %s with directory %s", server_url, dir)

    router := mux.NewRouter()
    web.SetupContented(router, dir, previewCount, limitCount)
    web.SetupStatic(router, "./static")

    srv := &http.Server{
        Handler:      router,
        Addr:         server_url,
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 18 * time.Second,
        ReadTimeout:  18 * time.Second,
    }
    log.Fatal(srv.ListenAndServe())
    http.Handle("/", router)

}

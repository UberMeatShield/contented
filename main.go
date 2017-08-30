package main

import (
    "fmt"
    "time"
    "flag"
    "log"
    "net/http"
    "gorilla/mux"
	"contented/web"
)

var dir string
func main() {
    flag.StringVar(&dir, "dir", ".", "Directory to serve files from")
    flag.Parse()

    fmt.Println("Using this directory As the static root: ", dir)

    router := mux.NewRouter()
	web.SetupContented(router, dir)
	web.SetupStatic(router, "./static")

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


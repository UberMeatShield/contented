package main

import (
    "fmt"
    "time"
    "flag"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "contented/web"
)

var dir string
var port string = "8000"
func main() {
    flag.StringVar(&dir, "dir", ".", "Directory to serve files from")
    flag.StringVar(&port, "port", "8000", "Port to run the webserver.")
    flag.Parse()

    fmt.Println("Using this directory As the static root: ", dir, port, "WAT")

    router := mux.NewRouter()
	web.SetupContented(router, dir)
	web.SetupStatic(router, "./static")

    srv := &http.Server{
        Handler:      router,
		Addr:         "127.0.0.1:" + port,
        // Good practice: enforce timeouts for servers you create!
		WriteTimeout: 18 * time.Second,
		ReadTimeout:  18 * time.Second,
    }
    log.Fatal(srv.ListenAndServe())
    http.Handle("/", router)
}


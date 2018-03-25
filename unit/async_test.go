package unit

import (
    "fmt"
    "net/http"
)

func Prime() {
    fmt.Println("Attempting to call content link.")
    res, err := http.Get("http://localhost:8000/content/")
    fmt.Println("Call has been sent out.")

    if err != nil {
        fmt.Println("Failed to load up the content")
    } else {
        fmt.Println("Success loading up results %s", res)
    }
}


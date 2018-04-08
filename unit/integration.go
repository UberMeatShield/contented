package unit

import (
    "fmt"
    "testing"
    "io/ioutil"
    "encoding/json"
    "contented/web"
    "net/http"
)

func IntegrationLoad(t *testing.T) {
    fmt.Println("Attempting to call content link.")
    res, err := http.Get("http://localhost:8000/content/")
    fmt.Println("Call has been sent out.")

    if err != nil {
        t.Error("Failed to load data from the contented server.", err)
    }
    body, readError := ioutil.ReadAll(res.Body)
    if readError != nil {
        t.Error("Could not read back a json response from the server.")
    }
    parsed := web.PreviewResults{}
    json.Unmarshal(body, parsed)

    if len(parsed.Results) > 0 {
        t.Log("Success loading content results from the server %d", len(parsed.Results))
    } else {
        t.Error("Failed to load a valid result set.", parsed)
    }
    fmt.Println("Sucess")
}


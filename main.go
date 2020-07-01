package main

import (
    "os"
	"log"
    "strconv"
	"contented/actions"
    "github.com/gobuffalo/envy"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	app := actions.App()

    dir, err := envy.MustGet("DIR")
    limitCount, limErr := strconv.Atoi(envy.Get("LIMIT", strconv.Itoa(actions.DefaultLimit)))
    previewCount, previewErr := strconv.Atoi(envy.Get("PREVIEW", strconv.Itoa(actions.DefaultPreviewCount)))

    if err != nil {
        panic(err)
    } else if limErr != nil {
        panic(limErr)
    } else if previewErr != nil {
        panic(previewErr)
    } else if  _, noDirErr := os.Stat(dir); os.IsNotExist(noDirErr) {
        panic(noDirErr)
    }

    log.Printf("Parsed Env. Dir %s Limit %d with preview count %d", dir, limitCount, previewCount)
    actions.SetupContented(app, dir, previewCount, limitCount)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/

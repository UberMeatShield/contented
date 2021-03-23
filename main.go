package main

import (
	"contented/actions"
	"log"
	"os"
	"strconv"
	"github.com/gobuffalo/envy"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {

    // Should I move this into the config itself?
    var err error
	dir := envy.Get("DIR", "")
    if dir == "" {
        dir, err = envy.MustGet("CONTENT_DIR")  // From the .env file
    }
	limitCount, limErr := strconv.Atoi(envy.Get("LIMIT", strconv.Itoa(actions.DefaultLimit)))

    // We need to get that actually get a default load somehow
	previewCount, previewErr := strconv.Atoi(envy.Get("PREVIEW", strconv.Itoa(actions.DefaultPreviewCount)))
    useDatabase, connErr := strconv.ParseBool(envy.Get("USE_DATABASE", strconv.FormatBool(actions.DefaultUseDatabase)))

	if err != nil {
		panic(err)
	} else if limErr != nil {
		panic(limErr)
	} else if previewErr != nil {
		panic(previewErr)
	} else if _, noDirErr := os.Stat(dir); os.IsNotExist(noDirErr) {
		panic(noDirErr)
    } else if connErr != nil {
        panic(connErr)
    }

    appCfg := actions.GetCfg()
    appCfg.UseDatabase = useDatabase

	log.Printf("Parsed Env. Dir %s Limit %d with preview count %d\n", dir, limitCount, previewCount)
    log.Printf("Use connection type of database %t\n", appCfg.UseDatabase)

	app := actions.App(appCfg.UseDatabase)
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

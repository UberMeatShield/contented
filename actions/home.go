package actions

import (
	"contented/utils"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(map[string]string{"message": "Contented Is Up"}))
}

// Replace this with nginx or something else better at serving static content (probably)
func AngularIndex(c buffalo.Context) error {
	cfg := utils.GetCfg()
	index := filepath.Join(cfg.StaticResourcePath, "index.html")

	// TODO:  I guess this is dumb if I have to read the thing anyway...
	body, err_read := ioutil.ReadFile(index)
	if err_read != nil {
		err_msg := "Could not find index.html: " + err_read.Error()
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": err_msg}))
	} else {
		return c.Render(http.StatusOK, r.Func("text/html", func(w io.Writer, d render.Data) error {
			_, err := w.Write([]byte(body))
			return err
		}))
	}
}

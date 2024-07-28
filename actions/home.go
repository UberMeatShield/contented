package actions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	//"github.com/gobuffalo/buffalo"
	//"github.com/gobuffalo/buffalo/render"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c *gin.Context) {
	obj := gin.H{"message": "Contented Is Up"}
	c.JSON(http.StatusOK, obj)
}

// Replace this with nginx or something else better at serving static content (probably)
func AngularIndex(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", gin.H{})
	//index := filepath.Join(cfg.StaticResourcePath, "index.html")

	/*
		cfg := utils.GetCfg()
		// TODO:  I guess this is dumb if I have to read the thing anyway...
		//body, err_read := ioutil.ReadFile(index)
		if err_read != nil {
			err_msg := "Could not find index.html: " + err_read.Error()
			c.JSON(http.StatusNotFound, map[string]string{"error": err_msg})
		} else {
			c.HTML(http.StatusOK, )
					r.Func("text/html", func(w io.Writer, d render.Data) error {
					_, err := w.Write([]byte(body))
					return err
				}))
		}
	*/
}

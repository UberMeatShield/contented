package actions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HomeHandler is a default handler to serve up a home page.
func StatusHandler(c *gin.Context) {
	obj := gin.H{"message": "Contented Is Up"}
	c.JSON(http.StatusOK, obj)
}

// Replace this with nginx or something else better at serving static content (probably)
func AngularIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

package grifts

import (
    "strconv"
    "fmt"
    "contented/models"
    "contented/actions"
	"github.com/markbates/grift/grift"
    "github.com/pkg/errors"
    "github.com/gobuffalo/envy"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Populate the DB with a set of directory content.")
	grift.Add("seed", func(c *grift.Context) error {
        err := models.DB.TruncateAll()
        if err != nil {
            return errors.WithStack(err)
        }

        // Grab the directory which we want to process
        dir_name, d_err := envy.MustGet("DIR")
        if d_err != nil {
            return errors.WithStack(d_err)
        }
        // Clean out the current DB setup then builds a new one
        return actions.CreateInitialStructure(dir_name)
	})
    // Then add the content for the entire directory structure

	grift.Add("preview", func(c *grift.Context) error {
        // TODO: Strip this out into a different function
        var size int64 = 1024 * 1000 * 2
        psize := envy.Get("PREVIEW_IF_GREATER", "")
        if psize != "" {
            setSize, i_err := strconv.ParseInt(psize, 10, 64)
            if i_err == nil {
                size = setSize
            }
        }
        fmt.Printf("Using size %d for preview creation", size)
        return actions.CreateAllPreviews(size)
	})
    // Then add in linkage to the related models.
})

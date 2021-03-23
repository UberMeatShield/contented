package grifts

import (
    "strconv"
    "fmt"
    "contented/models"
    "contented/utils"
	"github.com/markbates/grift/grift"
    "github.com/pkg/errors"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/nulls"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Populate the DB with a set of directory content.")
	grift.Add("seed", func(c *grift.Context) error {
        // Clean out the current DB setup
        err := models.DB.TruncateAll()
        if err != nil {
            return errors.WithStack(err)
        }

        // Grab the directory which we want to process
        dir_name, d_err := envy.MustGet("DIR")
        if d_err != nil {
            return errors.WithStack(d_err)
        }

        var size int64 = 1024 * 1000 * 2
        psize := envy.Get("PREVIEW_IF_GREATER", "")
        if psize != "" {
            setSize, i_err := strconv.ParseInt(psize, 10, 64)
            if i_err == nil {
                size = setSize
            }
        }
        fmt.Printf("Using size %d for preview creation", size)

        // Process all the directories and get a valid setup
        dirs := utils.FindContainers(dir_name)
        fmt.Printf("Found %d directories.\n", len(dirs))

        // TODO: Need to do this in a single transaction
        for _, dir := range dirs {
            fmt.Printf("Adding directory %s with id %s\n", dir.Name, dir.ID)

            media := utils.FindMedia(dir, 90001, 0) // A more sensible limit?
            fmt.Printf("Adding Media to %s with total media %d \n", dir.Name, len(media))

            // Use the database version of uuid generation (minimize the miniscule conflict)
            unset_uuid, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
            dir.ID = unset_uuid
            dir.Total = len(media)
            models.DB.Create(&dir)
            fmt.Printf("Created %s with id %s\n", dir.Name, dir.ID)

            for _, mc := range media {
                mc.ContainerID = nulls.NewUUID(dir.ID) 
                models.DB.Create(&mc)
            }
        }
		// Add DB seeding stuff here
		return nil
	})
    // Then add the content for the entire directory structure

    // Then add in linkage to the related models.

})

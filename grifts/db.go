package grifts

import (
    "contented/models"
    "contented/utils"
	"github.com/markbates/grift/grift"
    "github.com/pkg/errors"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Populate the DB with a set of directory content.")
	grift.Add("seed", func(c *grift.Context) error {
        err := models.DB.TruncateAll()
        if err != nil {
            return errors.WithStack(err)
        }

        env_dir := "/Users/justincarlson/go/src/contented/mocks/content"

        dirs := utils.ListDirs(env_dir, 4)
        for _, dir := range dirs {
            dirObj := &models.Container{
              Path: dir.Path,
              Name: dir.Name,
            }
            models.DB.Create(dirObj)
        }

		// Add DB seeding stuff here
		return nil
	})

    // Need to do a stanard lookup

    // Then add the content for the entire directory structure

    // Then add in linkage to the related models.

})

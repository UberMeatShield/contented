package actions

import (
	"contented/models"
	//"contented/utils"
	//"os"
	//"testing"
	//"github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/envy"
	// "github.com/gobuffalo/suite"
)

func (as *ActionSuite) TestInitialCreation() {
	dir, _ := envy.MustGet("DIR")
    as.NotEmpty(dir, "The test must specify a directory to run on")

    err := CreateInitialStructure(dir)
    as.NoError(err, "It should successfully create the full DB setup")

    cnts := models.Containers{}
    as.DB.All(&cnts)

    media := models.MediaContainers{}
    as.DB.All(&media)
    as.Equal(len(media), 23, "The mocks have a specific expected number of items")
}


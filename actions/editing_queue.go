package actions

import (
	"contented/managers"
	"contented/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

// Should deny quickly if the media content type is incorrect for the action
func TaskScreensHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	numberOfScreens, countErr := strconv.Atoi(c.Param("numberOfScreens"))
	log.Printf("Attempting to request a screens be built out")
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	cfg := utils.GetCfg()
	if countErr != nil {
		numberOfScreens = cfg.PreviewCount
	}
	man := managers.GetManager(&c)
	content, err := man.GetContent(mcID)
	if err != nil {
		return c.Error(404, err)
	}
	log.Printf("Not implemented mcID %s with number of screens %d", content.ID.String(), numberOfScreens)
	return c.Render(http.StatusBadRequest, r.JSON(content))
}

func TaskScreenHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	timeForScreen, timeErr := strconv.Atoi(c.Param("timeSeconds"))
	log.Printf("Attempting to request a screen at a particular time")
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	if timeErr != nil {
		msg := fmt.Sprintf("Bad timestamp %s", c.Param("timeSeconds"))
		return c.Error(400, errors.New(msg))
	}
	log.Printf("Media id %s with time %d", mcID, timeForScreen)
	return c.Error(400, errors.New("Not implemented"))
}

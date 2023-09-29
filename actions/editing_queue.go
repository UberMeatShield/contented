package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gofrs/uuid"
)

// Should deny quickly if the media content type is incorrect for the action
func TaskScreensHandler(c buffalo.Context) error {
	cfg := utils.GetCfg()
	contentID, bad_uuid := uuid.FromString(c.Param("contentID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	startTimeSeconds, startErr := strconv.Atoi(c.Param("startTimeSeconds"))
	if startErr != nil || startTimeSeconds < 0 {
		startTimeSeconds = cfg.PreviewFirstScreenOffset
	}
	numberOfScreens, countErr := strconv.Atoi(c.Param("count"))
	if countErr != nil {
		numberOfScreens = cfg.PreviewCount
	}
	if numberOfScreens <= 0 || numberOfScreens > 50 {
		return c.Error(http.StatusBadRequest, errors.New("Too many or few screens requested"))
	}

	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return c.Error(404, err)
	}
	log.Printf("Requesting screens be built out %s start %d count %d", content.Src, startTimeSeconds, numberOfScreens)
	tr := models.TaskRequest{
		ContentID:        content.ID,
		Operation:        models.TaskOperation.SCREENS,
		NumberOfScreens:  numberOfScreens,
		StartTimeSeconds: startTimeSeconds,
	}
	createdTR, tErr := man.CreateTask(&tr)
	if tErr != nil {
		c.Error(http.StatusInternalServerError, tErr)
	}
	return c.Render(http.StatusCreated, r.JSON(createdTR))
}

func ScreenCapture(args worker.Args) error {
	cfg := utils.GetCfg()
	app := App(cfg.UseDatabase)
	taskId := ""
	for k, v := range args {
		if k == "id" {
			taskId = v.(string)
		}
	}
	id, err := uuid.FromString(taskId)
	if err != nil {
		log.Printf("Failed to load task bad id %s", err)
		return err
	}
	log.Printf("Async Task being called %s have to figure out a DB connection %s", args, taskId)

	man := managers.GetAppManager(app)
	task, tErr := man.GetTask(id)
	if tErr != nil {
		log.Printf("Could not look up the task successfully %s", tErr)
		return tErr
	}
	log.Printf("Found the correct task %s", task)
	// managers.GetManager()
	return nil
}

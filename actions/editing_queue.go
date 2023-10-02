package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/pop/v6"
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
	if !strings.Contains(content.ContentType, "video") && content.NoFile == false {
		return c.Error(http.StatusBadRequest, errors.New("Content was not a video %s"))
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

/*
 * For all the transaction middleware to play nice you have to ensure that everything
 * is wrapped by a transaction
 */
func ScreenCaptureWrapper(args worker.Args) error {
	cfg := utils.GetCfg()
	getConnection := func() *pop.Connection {
		return nil
	}
	app := App(cfg.UseDatabase)

	// Note this is extra complicated by the fact it SHOULD be able to run with NO connections
	// or DB sessions made.
	if cfg.UseDatabase == true {
		// There has to be a good way to have all transaction middleware commit and work
		// without exploding and being fully wrapping the scope.
		return models.DB.Transaction(func(tx *pop.Connection) error {
			getConnection = func() *pop.Connection {
				return tx
			}
			man := managers.GetAppManager(app, getConnection)
			return ScreenCapture(man, args)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return ScreenCapture(man, args)
}

/**
 * Awkward to test.
 */
func ScreenCapture(man managers.ContentManager, args worker.Args) error {
	log.Printf("Trying to do a screen Capture but failing DB connections %s", args)
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

	task, tErr := man.GetTask(id)
	if tErr != nil {
		log.Printf("Could not look up the task successfully %s", tErr)
		return tErr
	}
	upTask, _ := ChangeTaskState(man, task, models.TaskStatus.PENDING, "Starting to execute task")
	task = upTask
	content, cErr := man.GetContent(task.ContentID)
	if cErr != nil {
		FailTask(man, task, fmt.Sprintf("Content not found %s %s", task.ContentID, cErr))
		return cErr
	}

	task, upErr := ChangeTaskState(man, task, models.TaskStatus.IN_PROGRESS, fmt.Sprintf("Content was found %s", content.Src))
	if upErr != nil {
		log.Printf("Failed to update task state to in progress %s", upErr)
		FailTask(man, task, fmt.Sprintf("Failed task intentionally %s", upErr))
		return upErr
	}

	screens, sErr, pattern := managers.CreateScreensForContent(man, task.ContentID, task.NumberOfScreens, task.StartTimeSeconds)
	if sErr != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", sErr)
		FailTask(man, task, failMsg)
	}
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("Successfully created screens %s", screens))
	log.Printf("Screens %s and the pattern %s", screens, pattern)
	return sErr
}

func ChangeTaskState(man managers.ContentManager, task *models.TaskRequest, newStatus models.TaskStatusType, msg string) (*models.TaskRequest, error) {
	status := task.Status
	task.Status = newStatus
	task.Message = msg
	log.Printf("Changing Task State %s", task)
	return man.UpdateTask(task, status)
}

func FailTask(man managers.ContentManager, task *models.TaskRequest, errMsg string) (*models.TaskRequest, error) {
	status := task.Status
	task.Status = models.TaskStatus.ERROR
	task.ErrMsg = errMsg
	log.Printf("Failing task becasue %s", task)
	return man.UpdateTask(task, status)
}

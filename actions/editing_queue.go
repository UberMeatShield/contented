package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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

func ScreenCapture(args worker.Args) error {
	cfg := utils.GetCfg()
	app := App(cfg.UseDatabase)
	log.Printf("Trying to do a screen Capture %s", args)

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
	upTask, _ := ChangeTaskState(man, task, models.TaskStatus.PENDING, "Starting to execute task")
	task = upTask
	content, cErr := man.GetContent(task.ContentID)
	if cErr != nil {
		FailTask(man, task, fmt.Sprintf("Failed to load content %s %s", task.ContentID, cErr))
		return cErr
	}
	cnt, cntErr := man.GetContainer(content.ContainerID.UUID)
	if cntErr != nil {
		FailTask(man, task, fmt.Sprintf("Failed to load container %s %s", task.ContentID, cntErr))
		return cntErr
	}
	task, upErr := ChangeTaskState(man, task, models.TaskStatus.IN_PROGRESS, fmt.Sprintf("Content was found %s", content.Src))
	if upErr != nil {
		log.Printf("Couldn't update task state %s", upErr)
		return upErr
	}

	path := cnt.GetFqPath()
	srcFile := filepath.Join(path, content.Src)
	dstPath := utils.GetPreviewDst(path)
	log.Printf("What is going on %s dst %s\n", srcFile, dstPath)

	filename := utils.GetPreviewPathDestination(content.Src, dstPath, "video")
	log.Printf("Src File %s and Destination %s", srcFile, filename)
	utils.MakePreviewPath(dstPath)
	fn := utils.GetScreensOutputPattern(filename)
	log.Printf("Final name %s", fn)
	/*
		dstFile := filepath.Join(dstDir, content.Src)
		dstFile := utils.GetScreensOutputPattern()
	*/

	// Now attempt to get a screen
	return nil
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
	log.Printf("Failing task %s", task)
	return man.UpdateTask(task, status)
}

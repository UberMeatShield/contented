package actions

import (
	"contented/managers"
	"contented/models"
	"contented/utils"
	"errors"
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

	// It should probably not kick off the job task inside the manager
	job := worker.Job{
		Queue:   "default",
		Handler: tr.Operation.String(),
		Args: worker.Args{
			"id": tr.ID.String(),
		},
	}
	App(cfg.UseDatabase).Worker.Perform(job)
	return c.Render(http.StatusCreated, r.JSON(createdTR))
}

/*
 * For all the transaction middleware to play nice you have to ensure that everything
 * is wrapped by a transaction
 */
func ScreenCaptureWrapper(args worker.Args) error {
	log.Printf("ScreenCaptureWrapper() Starting Task args %s", args)
	cfg := utils.GetCfg()
	getConnection := func() *pop.Connection {
		return nil
	}
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
			return managers.ScreenCapture(man, id)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.ScreenCapture(man, id)
}

package actions

/***
 * Wrapper functions are used by the buffalo queuing system to actually execute the various tasks.
 * The handler functions are used to actually add something in the API and validate the inputs.
 */
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
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

/**
 * Execute the task within the transaction middleware scope.
 * TODO: Can this work in a full unit test?
 */
func VideoEncodingWrapper(args worker.Args) error {
	log.Printf("VideoEncodingWrapper () Starting Task args %s", args)
	cfg := utils.GetCfg()
	getConnection := func() *pop.Connection {
		return nil
	}
	app := App(cfg.UseDatabase)
	taskId, err := GetTaskId(args)
	if err != nil {
		return err
	}
	// Note this is extra complicated by the fact it SHOULD be able to run with NO connections
	// or DB sessions made.
	if cfg.UseDatabase {
		return models.DB.Transaction(func(tx *pop.Connection) error {
			getConnection = func() *pop.Connection {
				return tx
			}
			man := managers.GetAppManager(app, getConnection)
			return managers.EncodingVideoTask(man, taskId)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.EncodingVideoTask(man, taskId)
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
	taskId, err := GetTaskId(args)
	if err != nil {
		return err
	}

	// Note this is extra complicated by the fact it SHOULD be able to run with NO connections
	// or DB sessions made.
	if cfg.UseDatabase {
		// There has to be a good way to have all transaction middleware commit and work
		// without exploding and being fully wrapping the scope.
		return models.DB.Transaction(func(tx *pop.Connection) error {
			getConnection = func() *pop.Connection {
				return tx
			}
			man := managers.GetAppManager(app, getConnection)
			return managers.ScreenCaptureTask(man, taskId)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.ScreenCaptureTask(man, taskId)
}

func WebpFromScreensWrapper(args worker.Args) error {
	log.Printf("Web From Screens () Starting Task args %s", args)
	cfg := utils.GetCfg()
	getConnection := func() *pop.Connection {
		return nil
	}
	app := App(cfg.UseDatabase)
	taskId, err := GetTaskId(args)
	if err != nil {
		return err
	}
	// Note this is extra complicated by the fact it SHOULD be able to run with NO connections
	// or DB sessions made.
	if cfg.UseDatabase == true {
		return models.DB.Transaction(func(tx *pop.Connection) error {
			getConnection = func() *pop.Connection {
				return tx
			}
			man := managers.GetAppManager(app, getConnection)
			return managers.WebpFromScreensTask(man, taskId)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.WebpFromScreensTask(man, taskId)
}

/*
 * Attempt to tag a piece of content
 */
func TaggingContentWrapper(args worker.Args) error {
	log.Printf("Tagging content element () Starting Task args %s", args)
	cfg := utils.GetCfg()
	getConnection := func() *pop.Connection {
		return nil
	}
	app := App(cfg.UseDatabase)
	taskId, err := GetTaskId(args)
	if err != nil {
		return err
	}
	// Note this is extra complicated by the fact it SHOULD be able to run with NO connections
	// or DB sessions made.
	if cfg.UseDatabase == true {
		return models.DB.Transaction(func(tx *pop.Connection) error {
			getConnection = func() *pop.Connection {
				return tx
			}
			man := managers.GetAppManager(app, getConnection)
			return managers.TaggingContentTask(man, taskId)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.TaggingContentTask(man, taskId)
}

func GetTaskId(args worker.Args) (uuid.UUID, error) {
	taskId := ""
	for k, v := range args {
		if k == "id" {
			taskId = v.(string)
		}
	}
	id, err := uuid.FromString(taskId)
	if err != nil {
		log.Printf("Failed to load task bad id %s", err)
		bad, _ := uuid.NewV4()
		return bad, err
	}
	return id, err
}

func WebpFromScreensHandler(c buffalo.Context) error {
	contentID, bad_uuid := uuid.FromString(c.Param("contentID"))
	if bad_uuid != nil {
		return c.Error(http.StatusBadRequest, bad_uuid)
	}
	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return nil
	}
	tr := models.TaskRequest{
		ContentID: content.ID,
		Operation: models.TaskOperation.WEBP,
	}
	return QueueTaskRequest(c, man, &tr)
}

// Should deny quickly if the media content type is incorrect for the action
func VideoEncodingHandler(c buffalo.Context) error {
	contentID, bad_uuid := uuid.FromString(c.Param("contentID"))
	if bad_uuid != nil {
		return c.Error(http.StatusBadRequest, bad_uuid)
	}
	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	if !strings.Contains(content.ContentType, "video") && !content.NoFile {
		return c.Error(http.StatusBadRequest, fmt.Errorf("content was not a video %s", content.ContentType))
	}
	// Probably should at least sanity check the codecs
	cfg := utils.GetCfg()
	codec := managers.StringDefault(c.Param("codec"), cfg.CodecForConversion)
	log.Printf("Requesting a re-encode %s with codec %s for contentID %s", content.Src, codec, content.ID.String())
	tr := models.TaskRequest{
		ContentID:        content.ID,
		Operation:        models.TaskOperation.ENCODING,
		NumberOfScreens:  0,
		StartTimeSeconds: 0,
		Codec:            codec,
	}
	return QueueTaskRequest(c, man, &tr)
}

// Should deny quickly if the media content type is incorrect for the action
func TaggingHandler(c buffalo.Context) error {
	contentID, bad_uuid := uuid.FromString(c.Param("contentID"))
	if bad_uuid != nil {
		return c.Error(http.StatusBadRequest, bad_uuid)
	}
	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	// TODO: Make it so it can tag everything under a container
	// TODO: Make it bail if there are no tags in the system

	// Probably should at least sanity check the codecs
	tr := models.TaskRequest{
		ContentID: content.ID,
		Operation: models.TaskOperation.TAGGING,
	}
	return QueueTaskRequest(c, man, &tr)
}

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
	if numberOfScreens <= 0 || numberOfScreens > 300 {
		return c.Error(http.StatusBadRequest, errors.New("too many or few screens requested"))
	}

	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return c.Error(404, err)
	}
	if !strings.Contains(content.ContentType, "video") && !content.NoFile {
		return c.Error(http.StatusBadRequest, errors.New("content was not a video %s"))
	}
	log.Printf("Requesting screens be built out %s start %d count %d", content.Src, startTimeSeconds, numberOfScreens)
	tr := models.TaskRequest{
		ContentID:        content.ID,
		Operation:        models.TaskOperation.SCREENS,
		NumberOfScreens:  numberOfScreens,
		StartTimeSeconds: startTimeSeconds,
	}
	return QueueTaskRequest(c, man, &tr)
}

func QueueTaskRequest(c buffalo.Context, man managers.ContentManager, tr *models.TaskRequest) error {
	createdTR, tErr := man.CreateTask(tr)
	if tErr != nil {
		c.Error(http.StatusInternalServerError, tErr)
	}
	// This needs to delay a little before it starts
	// It should probably not kick off the job task inside the manager
	cfg := man.GetCfg()
	job := worker.Job{
		Queue:   "default",
		Handler: tr.Operation.String(),
		Args: worker.Args{
			"id": tr.ID.String(),
		},
	}
	err := App(cfg.UseDatabase).Worker.PerformIn(job, 2*time.Second)
	if err != nil {
		msg := fmt.Sprintf("Failed to enqueue task in the work queue %s", err)
		log.Print(msg)
		return c.Error(http.StatusInternalServerError, errors.New(msg))
	}
	return c.Render(http.StatusCreated, r.JSON(createdTR))
}

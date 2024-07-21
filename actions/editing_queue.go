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
	"net/url"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

type TasksQueuedResponse struct {
	Message string              `json:"message" default:""`
	Results models.TaskRequests `json:"results" default:"[]"`
}

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
	if cfg.UseDatabase {
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
 * Attempt to tag a piece of content (tempting to just make this a switch)
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
	if cfg.UseDatabase {
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

func DuplicatesWrapper(args worker.Args) error {
	log.Printf("Finding Duplicates %s", args)
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
			return managers.DetectDuplicatesTask(man, taskId)
		})
	}
	// Memory manager version
	man := managers.GetAppManager(app, getConnection)
	return managers.DetectDuplicatesTask(man, taskId)
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
	tr, tErr := CreateWebpTask(content)
	if tErr != nil {
		c.Error(http.StatusBadRequest, tErr)
	}
	return QueueTaskRequest(c, man, tr)
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

	task, badTask := CreateVideoEncodingTask(content, c.Param("codec"))
	if badTask != nil {
		return c.Error(http.StatusBadRequest, badTask)
	}
	return QueueTaskRequest(c, man, task)
}

func ContainerVideoEncodingHandler(c buffalo.Context) error {
	containerID, bad_uuid := uuid.FromString(c.Param("containerID"))
	if bad_uuid != nil {
		return c.Error(http.StatusBadRequest, bad_uuid)
	}

	// A lot of these will follow a pretty simple pattern of load all the container content
	// and then attempt to act on them.  Unify it?
	man := managers.GetManager(&c)
	contentQuery := managers.ContentQuery{
		ContainerID: containerID.String(),
		ContentType: "video",
		PerPage:     man.GetCfg().Limit,
	}
	contents, total, err := man.ListContent(contentQuery)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}
	if total == 0 {
		queueResponse := TasksQueuedResponse{
			Message: "No video content found to re-encode",
			Results: models.TaskRequests{}, // hate
		}
		return c.Render(http.StatusOK, r.JSON(queueResponse))
	}

	// TODO: Need to make it so that we get all the tasks created.
	tasks := models.TaskRequests{}
	for _, content := range *contents {
		task, taskErr := CreateVideoEncodingTask(&content, c.Param("codec"))
		if taskErr != nil {
			return taskErr
		}
		tasks = append(tasks, *task)
	}
	return QueueTaskRequests(c, man, tasks)
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
	tr := models.TaskRequest{
		ContentID: nulls.NewUUID(content.ID),
		Operation: models.TaskOperation.TAGGING,
	}
	return QueueTaskRequest(c, man, &tr)
}

// Container tagging would probably be better as another single task or the managers just need to cache the tasks in redis.
func ContainerTaggingHandler(c buffalo.Context) error {
	containerID, bad_uuid := uuid.FromString(c.Param("containerID"))
	if bad_uuid != nil {
		return c.Error(http.StatusBadRequest, bad_uuid)
	}

	// A lot of these will follow a pretty simple pattern of load all the container content
	// and then attempt to act on them.  Unify it?
	man := managers.GetManager(&c)
	contentQuery := managers.ContentQuery{
		ContainerID: containerID.String(),
		PerPage:     man.GetCfg().Limit,
	}
	_, total, tagErr := man.ListAllTags(managers.TagQuery{PerPage: 1})
	if total == 0 || tagErr != nil {
		return c.Error(http.StatusBadRequest, fmt.Errorf("no tags currently found in the system"))
	}
	contents, total, err := man.ListContent(contentQuery)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}
	if total == 0 {
		queueResponse := TasksQueuedResponse{
			Message: "No content found to tag",
			Results: models.TaskRequests{},
		}
		return c.Render(http.StatusOK, r.JSON(queueResponse))
	}

	// TODO: Need to make it so that we get all the tasks created.
	tasks := models.TaskRequests{}
	for _, content := range *contents {
		task := models.TaskRequest{
			ContentID: nulls.NewUUID(content.ID),
			Operation: models.TaskOperation.TAGGING,
		}
		tasks = append(tasks, task)
	}
	return QueueTaskRequests(c, man, tasks)
}

// Should deny quickly if the media content type is incorrect for the action
func DupesHandler(c buffalo.Context) error {
	// Get content search from params
	man := managers.GetManager(&c)

	params := c.Params()
	cId := managers.StringDefault(params.Get("containerID"), "")
	id := managers.StringDefault(params.Get("contentID"), "")

	// It could just take 'nothing' and run against ALL video I guess.
	tr := models.TaskRequest{
		Operation: models.TaskOperation.DUPES,
	}
	query := managers.ContentQuery{
		ContentType: "video",
		PerPage:     1,
	}

	// This is kinda ugly, might want to make it just two handlers
	if cId != "" {
		if containerID, err := uuid.FromString(cId); err == nil {
			tr.ContainerID = nulls.NewUUID(containerID)
			query.ContainerID = cId
		} else {
			return c.Error(http.StatusBadRequest, fmt.Errorf("invalid containerID %s", cId))
		}
	} else if id != "" {
		if contentID, err := uuid.FromString(id); err == nil {
			tr.ContentID = nulls.NewUUID(contentID)
			query.ContentID = id
		} else {
			return c.Error(http.StatusBadRequest, fmt.Errorf("invalid containerID %s", cId))
		}
	} else {
		return c.Error(http.StatusBadRequest, errors.New("containerID or contentID are required"))
	}

	_, total, err := man.SearchContent(query)
	if err != nil {
		log.Printf("Cannot queue dupe task %s err: %s", query, err)
		return c.Error(http.StatusInternalServerError, err)
	}
	// TODO: Always return this format from the task kick off (kind of a pain for tests but 'eh')
	if total == 0 {
		res := TasksQueuedResponse{
			Message: "No duplicate videos found in this contianer",
			Results: models.TaskRequests{},
		}
		return c.Render(http.StatusOK, r.JSON(res))
	}
	if total < 1 {
		return c.Error(http.StatusBadRequest, fmt.Errorf("could not find content to check %s", query))
	}
	log.Printf("Attempting to queue task %s", tr)
	return QueueTaskRequest(c, man, &tr)
}

// Should deny quickly if the media content type is incorrect for the action
func ContentTaskScreensHandler(c buffalo.Context) error {
	contentID, bad_uuid := uuid.FromString(c.Param("contentID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	startTimeSeconds, numberOfScreens, err := ValidateScreensParams(c.Params().(url.Values))
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	man := managers.GetManager(&c)
	content, err := man.GetContent(contentID)
	if err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	tr, tErr := CreateScreensTask(content, numberOfScreens, startTimeSeconds)
	if tErr != nil {
		return c.Error(http.StatusBadRequest, tErr)
	}
	log.Printf("Requesting screens be built out %s start %d count %d", content.Src, startTimeSeconds, numberOfScreens)
	return QueueTaskRequest(c, man, tr)
}

func ValidateScreensParams(params url.Values) (int, int, error) {
	cfg := utils.GetCfg()

	startTimeSeconds, startErr := strconv.Atoi(params.Get("startTimeSeconds"))
	if startErr != nil || startTimeSeconds < 0 {
		startTimeSeconds = cfg.PreviewFirstScreenOffset
	}
	numberOfScreens, countErr := strconv.Atoi(params.Get("count"))
	if countErr != nil {
		numberOfScreens = cfg.PreviewCount
	}
	if numberOfScreens <= 0 || numberOfScreens > 300 {
		return startTimeSeconds, numberOfScreens, errors.New("too many or few screens requested")
	}
	return startTimeSeconds, numberOfScreens, nil
}

func ContainerScreensHandler(c buffalo.Context) error {
	cID, badUuid := uuid.FromString(c.Param("containerID"))
	if badUuid != nil {
		return c.Error(http.StatusBadRequest, badUuid)
	}
	cfg := utils.GetCfg()
	startTimeSeconds, numberOfScreens, err := ValidateScreensParams(c.Params().(url.Values))
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	cQ := managers.ContentQuery{
		ContainerID: cID.String(),
		ContentType: "video",
		PerPage:     cfg.Limit,
	}
	man := managers.GetManager(&c)
	contents, total, err := man.SearchContent(cQ)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}
	if total == 0 {
		res := TasksQueuedResponse{
			Message: "No videos found to create screens for",
			Results: models.TaskRequests{},
		}
		return c.Render(http.StatusOK, r.JSON(res))
	}

	tasks := models.TaskRequests{}
	for _, content := range *contents {
		task, taskErr := CreateScreensTask(&content, numberOfScreens, startTimeSeconds)
		if taskErr != nil {
			return taskErr
		}
		tasks = append(tasks, *task)
	}
	return QueueTaskRequests(c, man, tasks)
}

func CreateScreensTask(content *models.Content, numberOfScreens int, startTimeSeconds int) (*models.TaskRequest, error) {
	if !content.IsVideo() {
		return nil, fmt.Errorf("content was not a video %s", content.ContentType)
	}
	tr := models.TaskRequest{
		ContentID:        nulls.NewUUID(content.ID),
		Operation:        models.TaskOperation.SCREENS,
		NumberOfScreens:  numberOfScreens,
		StartTimeSeconds: startTimeSeconds,
	}
	return &tr, nil
}

func CreateVideoEncodingTask(content *models.Content, codecChoice string) (*models.TaskRequest, error) {
	// Probably should at least sanity check the codecs
	if !content.IsVideo() {
		return nil, fmt.Errorf("content %s was not a video %s", content.Src, content.ContentType)
	}
	cfg := utils.GetCfg()
	codec := managers.StringDefault(codecChoice, cfg.CodecForConversion)

	// Check codec seems valid?
	log.Printf("Requesting a re-encode %s with codec %s for contentID %s", content.Src, codec, content.ID.String())
	tr := models.TaskRequest{
		ContentID:        nulls.NewUUID(content.ID),
		Operation:        models.TaskOperation.ENCODING,
		NumberOfScreens:  0,
		StartTimeSeconds: 0,
		Codec:            codec,
	}
	return &tr, nil
}

func CreateWebpTask(content *models.Content) (*models.TaskRequest, error) {
	// Check required since it was not a search
	if !content.IsVideo() {
		return nil, fmt.Errorf("cannot create screens content was not video %s", content.ContentType)
	}

	// TODO: The actual task processing should check if the entry has actual screens
	tr := models.TaskRequest{
		ContentID: nulls.NewUUID(content.ID),
		Operation: models.TaskOperation.WEBP,
	}
	return &tr, nil
}

func QueueTaskRequest(c buffalo.Context, man managers.ContentManager, tr *models.TaskRequest) error {
	taskCreated, queueErr := AddTaskRequest(man, tr)
	if queueErr != nil {
		return c.Error(http.StatusInternalServerError, queueErr)
	}
	return c.Render(http.StatusCreated, r.JSON(taskCreated))
}

// Hande a partial failure
func QueueTaskRequests(c buffalo.Context, man managers.ContentManager, tasks models.TaskRequests) error {
	tasksOk := models.TaskRequests{}
	for _, task := range tasks {
		taskCreated, queueErr := AddTaskRequest(man, &task)
		if queueErr != nil {
			return c.Error(http.StatusInternalServerError, queueErr)
		}
		tasksOk = append(tasksOk, *taskCreated)
	}

	queueResponse := TasksQueuedResponse{
		Message: fmt.Sprintf("Queued %d tasks for", len(tasksOk)),
		Results: tasksOk,
	}
	return c.Render(http.StatusCreated, r.JSON(queueResponse))
}

func AddTaskRequest(man managers.ContentManager, tr *models.TaskRequest) (*models.TaskRequest, error) {
	createdTask, tErr := man.CreateTask(tr)
	if tErr != nil {
		return nil, tErr
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

	// This would be a good place to add in support for other task queues
	err := App(cfg.UseDatabase).Worker.PerformIn(job, 2*time.Second)
	if err != nil {
		msg := fmt.Sprintf("Failed to enqueue task in the work queue %s", err)
		log.Print(msg)
		return nil, err
	}
	return createdTask, nil
}

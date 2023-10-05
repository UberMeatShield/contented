/**
 * The manager setup supports connecting to a database or just using an in memory representation that can be
 * hosted.  The Managers load either from the utility/MemStorage or provides wrappers around Pop connections
 * to data loaded into a DB by using the db:seed grift.
 *
 * This is also an example of dealing with some of the annoying gunk GoLang hits you with if you want to
 * use an interface that is also semi configurable.  You know... prevent code duplication and that kind of
 * thing.
 */
package managers

import (
	"contented/models"
	"contented/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

type GetConnType func() *pop.Connection
type GetParamsType func() *url.Values
type GetAppWorker func() worker.Worker

type SearchRequest struct {
	Text        string   `json:"text" default:""`
	Page        int      `json:"page" default:"1"`
	PerPage     int      `json:"per_page" default:"10"`
	ContainerID string   `json:"container_id" default:""`
	ContentType string   `json:"content_type" default:""`
	Hidden      bool     `json:"hidden" default:"false"`
	Tags        []string `json:"tags" default:"[]"`
}

type TaskQuery struct {
	Page      int    `json:"page" default:"1"`
	PerPage   int    `json:"per_page" default:"100"`
	ContentID string `json:"content_id" default:""`
	Status    string `json:"status" default:""`
}

func (sr SearchRequest) String() string {
	s, _ := json.MarshalIndent(sr, "", "  ")
	return string(s)
}

// This is the primary interface used by the Buffalo actions.
type ContentManager interface {
	GetCfg() *utils.DirConfigEntry
	CanEdit() bool // Do we support CRUD or just R

	// Utility
	GetParams() *url.Values
	FindActualFile(mc *models.Content) (string, error)

	// Container Management
	GetContainer(cID uuid.UUID) (*models.Container, error)
	ListContainers(page int, per_page int) (*models.Containers, error)
	ListContainersFiltered(page int, per_page int, includeHidden bool) (*models.Containers, error)
	ListContainersContext() (*models.Containers, error)
	UpdateContainer(c *models.Container) (*models.Container, error)
	CreateContainer(c *models.Container) error
	DestroyContainer(id string) (*models.Container, error)

	// Content listing (why did I name it Content vs Media?)
	GetContent(content_id uuid.UUID) (*models.Content, error)
	ListContent(ContainerID uuid.UUID, page int, per_page int) (*models.Contents, error)
	ListContentContext(ContainerID uuid.UUID) (*models.Contents, error)
	ListAllContent(page int, per_page int) (*models.Contents, error)
	SearchContentContext() (*models.Contents, int, error)
	SearchContent(sr SearchRequest) (*models.Contents, int, error)
	SearchContainers(search string, page int, per_page int, includeHidden bool) (*models.Containers, error)
	UpdateContent(content *models.Content) error
	DestroyContent(id string) (*models.Content, error)
	CreateContent(mc *models.Content) error
	GetPreviewForMC(mc *models.Content) (string, error)

	// Functions that help with viewing movie screens if found.
	ListAllScreens(page int, per_page int) (*models.Screens, error)
	ListAllScreensContext() (*models.Screens, error)
	ListScreensContext(mcID uuid.UUID) (*models.Screens, error)
	ListScreens(mcID uuid.UUID, page int, per_page int) (*models.Screens, error)
	GetScreen(psID uuid.UUID) (*models.Screen, error)
	CreateScreen(s *models.Screen) error
	UpdateScreen(s *models.Screen) error
	DestroyScreen(id string) (*models.Screen, error)

	// Tags listing
	GetTag(id string) (*models.Tag, error)
	ListAllTags(page int, perPage int) (*models.Tags, error)
	ListAllTagsContext() (*models.Tags, error)
	CreateTag(tag *models.Tag) error
	UpdateTag(tag *models.Tag) error
	DestroyTag(id string) (*models.Tag, error)
	AssociateTag(tag *models.Tag, c *models.Content) error
	AssociateTagByID(tagID string, mcID uuid.UUID) error

	// For processing encoding requests
	CreateTask(task *models.TaskRequest) (*models.TaskRequest, error)
	UpdateTask(task *models.TaskRequest, currentStatus models.TaskStatusType) (*models.TaskRequest, error)
	NextTask() (*models.TaskRequest, error) // Assigns it (not really required yet)

	// For the API exposed
	ListTasksContext() (*models.TaskRequests, error)
	ListTasks(query TaskQuery) (*models.TaskRequests, error)
	GetTask(id uuid.UUID) (*models.TaskRequest, error)
}

// Dealing with buffalo.Context vs grift.Context is kinda annoying, this handles the
// buffalo context which handles tests or runtime but doesn't work in grifts.
func GetManager(c *buffalo.Context) ContentManager {
	cfg := utils.GetCfg()
	ctx := *c

	// The get connection might need an async channel or it potentially locks
	// the dev server :(.   Need to only do this if use database is setup and connects
	var get_connection GetConnType
	if cfg.UseDatabase {
		var conn *pop.Connection
		get_connection = func() *pop.Connection {
			if conn == nil {
				tx, ok := ctx.Value("tx").(*pop.Connection)
				if !ok {
					log.Fatalf("Failed to get a connection")
					return nil
				}
				conn = tx
			}
			return conn
		}
	} else {
		// Just required for the memory version create statement
		get_connection = func() *pop.Connection {
			return nil
		}
	}
	get_params := func() *url.Values {
		params := ctx.Params().(url.Values)
		return &params
	}
	return CreateManager(cfg, get_connection, get_params)
}

// this is sketchy because of the connection scope closing on us
func GetAppManager(app *buffalo.App, getConnection GetConnType) ContentManager {
	cfg := utils.GetCfg()
	getParams := func() *url.Values {
		return &url.Values{}
	}
	return CreateManager(cfg, getConnection, getParams)
}

// can this manager create, update or destroy
func ManagerCanCUD(c *buffalo.Context) (*ContentManager, *pop.Connection, error) {
	man := GetManager(c)
	ctx := *c
	if man.CanEdit() == false {
		return &man, nil, ctx.Error(
			http.StatusNotImplemented,
			errors.New("Edit not supported by this manager"),
		)
	}
	if man.GetCfg().UseDatabase {
		tx, ok := ctx.Value("tx").(*pop.Connection)
		if !ok {
			return &man, nil, fmt.Errorf("No transaction found")
		}
		return &man, tx, nil
	}
	return &man, nil, nil
}

// Provides the ability to pass a connection function and get params function to the manager so we can handle
// a request.  We set this up so that the interface can still use the buffalo connection and param management.
func CreateManager(cfg *utils.DirConfigEntry, get_conn GetConnType, get_params GetParamsType) ContentManager {
	if cfg.UseDatabase {
		// Not really important for a DB manager, just need to look at it
		log.Printf("Setting up the DB Manager")
		db_man := ContentManagerDB{cfg: cfg}
		db_man.GetConnection = get_conn
		db_man.Params = get_params
		return db_man
	} else {
		// This should now be used to build the filesystem into memory one time.
		log.Printf("Setting up the memory Manager")
		mem_man := ContentManagerMemory{cfg: cfg}
		mem_man.Params = get_params
		mem_man.Initialize() // Break this into a sensible initialization
		return mem_man
	}
}

// How is it that GoLang doesn't have a more sensible default fallback?
func StringDefault(s1 string, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
}

// Used when doing pagination on the arrays of memory manager
func GetOffsetEnd(page int, per_page int, max int) (int, int) {
	offset := (page - 1) * per_page
	if offset < 0 {
		offset = 0
	}

	end := offset + per_page
	if end > max {
		end = max
	}
	// Maybe just toss a value error when asking for out of bounds
	if offset >= max {
		offset = end - 1
	}
	return offset, end
}

// Returns the offest, limit, page from pagination params (page indexing starts at 1)
func GetPagination(params pop.PaginationParams, DefaultLimit int) (int, int, int) {
	p := StringDefault(params.Get("page"), "1")
	page, err := strconv.Atoi(p)
	if err != nil || page < 1 {
		page = 1
	}

	perPage := StringDefault(params.Get("per_page"), strconv.Itoa(DefaultLimit))
	limit, err := strconv.Atoi(perPage)
	if err != nil || limit < 1 {
		limit = DefaultLimit
	}
	offset := (page - 1) * limit
	return offset, limit, page
}

// TODO: Fill in tags if they are provided.
func ContextToSearchRequest(params pop.PaginationParams, cfg *utils.DirConfigEntry) SearchRequest {
	_, per_page, page := GetPagination(params, cfg.Limit)
	sReq := SearchRequest{
		Text:        StringDefault(params.Get("text"), ""),
		ContainerID: StringDefault(params.Get("cId"), ""),
		ContentType: StringDefault(params.Get("contentType"), ""),
		PerPage:     per_page,
		Page:        page,
		Hidden:      false,
	}

	tagStr := StringDefault(params.Get("tags"), "")
	if tagStr != "" {
		sReq.Tags = strings.Split(tagStr, ",")
	}
	return sReq
}

func GetContentAndContainer(cm ContentManager, contentID uuid.UUID) (*models.Content, *models.Container, error) {
	content, cErr := cm.GetContent(contentID)
	if cErr != nil {
		return nil, nil, cErr
	}
	cnt, cntErr := cm.GetContainer(content.ContainerID.UUID)
	if cntErr != nil {
		return nil, nil, cntErr
	}
	return content, cnt, nil
}

func CreateScreensForContent(cm ContentManager, contentID uuid.UUID, count int, offset int) ([]string, error, string) {
	// It would be good to have the screens element take a few more params and have a wrapper on the
	// Content manager level.
	content, cnt, err := GetContentAndContainer(cm, contentID)
	if err != nil {
		return nil, err, ""
	}
	path := cnt.GetFqPath()
	srcFile := filepath.Join(path, content.Src)
	dstPath := utils.GetPreviewDst(path)
	dstFile := utils.GetPreviewPathDestination(content.Src, dstPath, "video")

	log.Printf("Src file %s and Destination %s", srcFile, dstFile)
	utils.MakePreviewPath(dstPath)
	screens, err, ptrn := utils.CreateSeekScreens(srcFile, dstFile, count, offset)

	for idx, sFile := range screens {
		src := strings.ReplaceAll(sFile, dstPath, "")
		s := models.Screen{
			Src:       src,
			Path:      dstPath,
			Idx:       idx,
			ContentID: contentID,
			SizeBytes: 0,
		}
		sErr := cm.CreateScreen(&s)
		if sErr != nil {
			log.Printf("Failed to create a screen %s", sErr)
		} else {
			log.Printf("Screen not actually in the DB? %s", s)
		}
	}
	return screens, err, ptrn
}

/**
 * Capture a set of screens given a task
 */
func ScreenCaptureTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers Screen Capture for taskID %s", id)
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

	screens, sErr, pattern := CreateScreensForContent(man, task.ContentID, task.NumberOfScreens, task.StartTimeSeconds)
	if sErr != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", sErr)
		FailTask(man, task, failMsg)
		return sErr
	}
	ChangeTaskState(man, task, models.TaskStatus.DONE, fmt.Sprintf("Successfully created screens %s", pattern))
	log.Printf("Screens %s and the pattern %s", screens, pattern)
	return sErr
}

/**
 * Could definitely make this a method assuming the next task uses the same logic.
 */
func EncodingVideoTask(man ContentManager, id uuid.UUID) error {
	log.Printf("Managers Encoding taskID %s", id)
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

	msg, encodeErr, shouldEncode := EncodeVideoContent(man, content, task.Codec)
	log.Printf("Video Encode video %s %s %t", msg, encodeErr, shouldEncode)
	if encodeErr != nil {
		failMsg := fmt.Sprintf("Failing to create screen %s", encodeErr)
		FailTask(man, task, failMsg)
		return encodeErr
	}
	taskMsg := fmt.Sprintf("Completed video encoding %s and had to encode %t", msg, shouldEncode)
	_, doneErr := ChangeTaskState(man, task, models.TaskStatus.DONE, taskMsg)
	return doneErr
}

// Should get a bunch of crap here
func EncodeVideoContent(man ContentManager, content *models.Content, codec string) (string, error, bool) {
	content, cnt, err := GetContentAndContainer(man, content.ID)
	if err != nil {
		return "No content to encode", err, false
	}
	path := cnt.GetFqPath()
	srcFile := filepath.Join(path, content.Src)
	dstFile := utils.GetVideoConversionName(srcFile)
	return utils.ConvertVideoToH256(srcFile, dstFile)
}

func ChangeTaskState(man ContentManager, task *models.TaskRequest, newStatus models.TaskStatusType, msg string) (*models.TaskRequest, error) {
	status := task.Status
	task.Status = newStatus
	task.Message = msg
	log.Printf("Changing Task State %s", task)
	return man.UpdateTask(task, status)
}

func FailTask(man ContentManager, task *models.TaskRequest, errMsg string) (*models.TaskRequest, error) {
	status := task.Status
	task.Status = models.TaskStatus.ERROR
	task.ErrMsg = errMsg
	log.Printf("Failing task becasue %s", task)
	return man.UpdateTask(task, status)
}

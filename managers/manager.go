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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"golang.org/x/exp/maps"
)

type GetConnType func() *pop.Connection
type GetParamsType func() *url.Values
type GetAppWorker func() worker.Worker

type TaskQuery struct {
	Page      int    `json:"page" default:"1"`
	PerPage   int    `json:"per_page" default:"100"`
	ContentID string `json:"content_id" default:""`
	Order     string `json:"order" default:"created_at"`
	Status    string `json:"status" default:""`
	Direction string `json:"direction" default:"desc"`
	Search    string `json:"search" default:""`
}

type ScreensQuery struct {
	Text      string `json:"text" default:""`
	Page      int    `json:"page" default:"1"`
	PerPage   int    `json:"per_page" default:"100"`
	ContentID string `json:"content_id" default:""`
	Order     string `json:"order" default:"created_at"`
	Direction string `json:"direction" default:"desc"`
}

type ContainerQuery struct {
	Name          string `json:"name" default:""`
	Search        string `json:"search" default:""`
	Page          int    `json:"page" default:"1"`
	PerPage       int    `json:"per_page" default:"100"`
	IncludeHidden bool   `json:"hidden" default:"false"`
	Order         string `json:"order" default:"created_at"`
	Direction     string `json:"direction" default:"desc"`
}

type ContentQuery struct {
	Search        string   `json:"search" default:""`
	Text          string   `json:"text" default:""`
	Page          int      `json:"page" default:"1"`
	PerPage       int      `json:"per_page" default:"1000"`
	ContentType   string   `json:"content_type" default:""`
	ContainerID   string   `json:"container_id" default:""`
	IncludeHidden bool     `json:"hidden" default:"false"`
	Order         string   `json:"order" default:"created_at"`
	Tags          []string `json:"tags" default:"[]"`
	Direction     string   `json:"direction" default:"desc"`
}

type TagQuery struct {
	Search  string `json:"search" default:""`
	Text    string `json:"text" default:""`
	Page    int    `json:"page" default:"1"`
	PerPage int    `json:"per_page" default:"1000"` // Doesn't work on create?
	TagType string `json:"tag_type" default:""`
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
	ListContainers(cq ContainerQuery) (*models.Containers, int, error)
	ListContainersFiltered(cq ContainerQuery) (*models.Containers, int, error)
	ListContainersContext() (*models.Containers, int, error)
	UpdateContainer(c *models.Container) (*models.Container, error)
	CreateContainer(c *models.Container) error
	DestroyContainer(id string) (*models.Container, error)

	// Content listing (why did I name it Content vs Media?)
	GetContent(content_id uuid.UUID) (*models.Content, error)
	ListContent(cs ContentQuery) (*models.Contents, int, error)
	ListContentContext() (*models.Contents, int, error)

	SearchContentContext() (*models.Contents, int, error)
	SearchContent(cq ContentQuery) (*models.Contents, int, error)
	SearchContainers(cs ContainerQuery) (*models.Containers, int, error)

	UpdateContent(content *models.Content) error
	UpdateContents(content models.Contents) error
	DestroyContent(id string) (*models.Content, error)
	CreateContent(mc *models.Content) error
	GetPreviewForMC(mc *models.Content) (string, error)

	// Functions that help with viewing movie screens if found.
	ListScreensContext() (*models.Screens, int, error)
	ListScreens(sr ScreensQuery) (*models.Screens, int, error)

	GetScreen(psID uuid.UUID) (*models.Screen, error)
	CreateScreen(s *models.Screen) error
	UpdateScreen(s *models.Screen) error
	DestroyScreen(id string) (*models.Screen, error)

	// Tags listing (oy do I need to deal with this?)
	GetTag(id string) (*models.Tag, error)
	ListAllTags(tq TagQuery) (*models.Tags, int, error)
	ListAllTagsContext() (*models.Tags, int, error)
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
	ListTasksContext() (*models.TaskRequests, int, error)
	ListTasks(query TaskQuery) (*models.TaskRequests, int, error)
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
	if !man.CanEdit() {
		return &man, nil, ctx.Error(
			http.StatusNotImplemented,
			errors.New("edit not supported by this manager"),
		)
	}
	if man.GetCfg().UseDatabase {
		tx, ok := ctx.Value("tx").(*pop.Connection)
		if !ok {
			return &man, nil, fmt.Errorf("no transaction found")
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
	if per_page <= 0 {
		per_page = utils.DefaultLimit
	}
	if page <= 0 {
		page = 1
	}

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
func ContextToContentQuery(params pop.PaginationParams, cfg *utils.DirConfigEntry) ContentQuery {
	_, per_page, page := GetPagination(params, cfg.Limit)
	sReq := ContentQuery{
		Text:          StringDefault(params.Get("text"), ""),
		Search:        StringDefault(params.Get("search"), ""),
		ContainerID:   StringDefault(params.Get("cId"), ""),
		ContentType:   StringDefault(params.Get("contentType"), ""),
		PerPage:       per_page,
		Page:          page,
		IncludeHidden: false,
		Order:         StringDefault(params.Get("order"), ""),
	}
	tags, err := GetTagsFromParam(params.Get("tags"))
	if err == nil {
		// BAIL / reject
		sReq.Tags = tags
	} else {
		log.Printf("Failed to parse query tags, ignoring %s", err)
	}
	return sReq
}

// Should this check if it is single non-array thing and then make it an array?
func GetTagsFromParam(val string) ([]string, error) {
	tags := []string{}
	if val != "" {
		err := json.Unmarshal([]byte(val), &tags)
		return tags, err
	}
	return tags, nil
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

// Should get a bunch of crap here (TODO: Error should always come last)
func EncodeVideoContent(man ContentManager, content *models.Content, codec string) (string, error, bool, string) {
	content, cnt, err := GetContentAndContainer(man, content.ID)
	if err != nil {
		return "No content to encode", err, false, ""
	}
	path := cnt.GetFqPath()
	srcFile := filepath.Join(path, content.Src)
	dstFile := utils.GetVideoConversionName(srcFile)
	msg, eErr, shouldEncode := utils.ConvertVideoToH265(srcFile, dstFile)
	return msg, eErr, shouldEncode, dstFile
}

/**
 * Do it ugly.  TODO: Make it less ugly.
 */
func AssignTagsAndUpdate(man ContentManager, tags models.Tags) error {
	if len(tags) == 0 {
		log.Printf("No tags in the system, nothing to do")
	}

	// Get the total count of content
	cfg := man.GetCfg()
	cs := ContentQuery{PerPage: 0}
	_, total, err := man.SearchContent(cs)
	if err != nil {
		return err
	}

	// Do a loop if total > limit
	offset := 0
	page := 0
	for offset < total {
		offset += cfg.Limit
		page += 1
		cs.Page = page
		cs.PerPage = cfg.Limit
		contents, _, pageErr := man.SearchContent(cs)
		if pageErr != nil {
			log.Printf("Failed to page over content %s", pageErr)
			return pageErr
		}
		tagErr := AssignTagsToContents(man, contents, &tags)
		if tagErr != nil {
			return tagErr
		}
	}
	return nil
}

/**
 * Used to assign a bunch of tags and then do an update to that content.
 */
func AssignTagsToContents(man ContentManager, contents *models.Contents, tags *models.Tags) error {
	if contents == nil || len(*contents) == 0 {
		log.Printf("No content to tag")
		return nil
	}
	if tags == nil || len(*tags) == 0 {
		log.Printf("No tags to match against")
		return nil
	}

	tagMap := models.TagsMap{}
	for _, tag := range *tags {
		tagMap[tag.ID] = tag
	}

	contentMap := models.ContentMap{}
	for _, content := range *contents {
		contentMap[content.ID] = content
	}
	// Ugly but at least allows for a batch update statement later
	updatedMap := utils.AssignTagsToContent(contentMap, tagMap)
	updatedContent := maps.Values(updatedMap)
	upErr := man.UpdateContents(updatedContent)
	if upErr != nil {
		return upErr
	}
	return nil
}

// HMMMM, should this be smarter?
func WebpFromContent(man ContentManager, content *models.Content) (string, error) {
	sr := ScreensQuery{ContentID: content.ID.String()}
	screens, count, err := man.ListScreens(sr)
	if err != nil {
		return "", err
	}

	if screens == nil || count <= 0 {
		return "", errors.New("not enough screens to create a preview")
	}
	_, cnt, err := GetContentAndContainer(man, content.ID)
	if err != nil {
		return "No content to encode", err
	}

	// It would be nice if the screens path could actually be a list of files
	// but I never managed to get that working.
	path := cnt.GetFqPath()
	dstPath := utils.GetPreviewDst(path)
	dstFile := utils.GetPreviewPathDestination(content.Src, dstPath, "video")
	globMatch := utils.GetScreensOutputGlob(dstFile)

	webp, err := utils.CreateWebpFromScreens(globMatch, dstFile)
	if err != nil {
		return webp, err
	}
	// log.Printf("What is the webp? %s", webp)
	content.Preview = utils.GetRelativePreviewPath(webp, cnt.GetFqPath())
	upErr := man.UpdateContent(content)
	log.Printf("What is the webp preview? %s", content.Preview)
	return webp, upErr
}

/*
* This function will check if we already have content for a new file after an encoding request.
 */
func CreateContentAfterEncoding(man ContentManager, originalContent *models.Content, newFile string) (*models.Content, error) {
	// First we check if the file ACTUALLY exists.
	path := filepath.Dir(newFile)
	if f, ok := os.Stat(newFile); ok == nil {

		// Check if we already have a content object for this.
		sr := ContentQuery{Text: f.Name(), ContainerID: originalContent.ContainerID.UUID.String()}
		contents, _, err := man.SearchContent(sr)
		if err != nil {
			return nil, err
		}
		if contents != nil && len(*contents) == 1 {
			cnts := *contents
			return &cnts[0], nil
		}

		newId, _ := uuid.NewV4()
		newContent := utils.GetContent(newId, f, path)
		newContent.Description = originalContent.Description
		newContent.Tags = originalContent.Tags
		newContent.ContainerID = originalContent.ContainerID
		createErr := man.CreateContent(&newContent)
		if createErr != nil {
			msg := fmt.Sprintf("Failed to create a newly encoded piece of content (re-encode might have worked). %s", createErr)
			return nil, errors.New(msg)
		}
		log.Printf("Created a new content element after encoding %s", newContent)
		return &newContent, nil
	}
	return nil, fmt.Errorf("%s file did not exist", newFile)
}

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
	"contented/pkg/config"
	"contented/pkg/models"
	"contented/pkg/utils"
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

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/maps"
	"gorm.io/gorm"
)

type GetConnType func() *gorm.DB
type GetParamsType func() *url.Values

type TaskQuery struct {
	Page        int    `json:"page" default:"1"`
	Offset      int    `json:"-" default:"0"`
	PerPage     int    `json:"per_page" default:"100"`
	ContentID   string `json:"content_id" default:""`
	ContainerID string `json:"container_id" default:""`
	Order       string `json:"order" default:"created_at"`
	Status      string `json:"status" default:""`
	Direction   string `json:"direction" default:"desc"`
	Search      string `json:"search" default:""`
}

func (t TaskQuery) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

type ScreensQuery struct {
	Text      string `json:"text" default:""`
	Page      int    `json:"page" default:"1"`
	Offset    int    `json:"-" default:"0"`
	PerPage   int    `json:"per_page" default:"100"`
	ContentID string `json:"content_id" default:""`
	Order     string `json:"order" default:"created_at"`
	Direction string `json:"direction" default:"desc"`
}

func (t ScreensQuery) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

type ContainerQuery struct {
	Name          string `json:"name" default:""`
	Search        string `json:"search" default:""`
	Page          int    `json:"page" default:"1"`
	Offset        int    `json:"-" default:"0"`
	PerPage       int    `json:"per_page" default:"100"`
	IncludeHidden bool   `json:"hidden" default:"false"`
	Order         string `json:"order" default:"created_at"`
	Direction     string `json:"direction" default:"desc"`
}

func (t ContainerQuery) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

type ContentQuery struct {
	Search        string   `json:"search" default:""`
	Text          string   `json:"text" default:""`
	Page          int      `json:"page" default:"1"`
	Offset        int      `json:"-" default:"0"`
	PerPage       int      `json:"per_page" default:"1000"`
	ContentType   string   `json:"content_type" default:""`
	ContainerID   string   `json:"container_id" default:""`
	ContentID     string   `json:"content_id" default:""`
	IncludeHidden bool     `json:"hidden" default:"false"`
	Order         string   `json:"order" default:"created_at"`
	Tags          []string `json:"tags" default:"[]"`
	Direction     string   `json:"direction" default:"desc"`
	Duplicate     bool     `json:"duplicate" default:"false"`
}

type TagQuery struct {
	Search  string `json:"search" default:""`
	Text    string `json:"text" default:""`
	Page    int    `json:"page" default:"1"`
	Offset  int    `json:"-" default:"0"`
	PerPage int    `json:"per_page" default:"1000"` // Doesn't work on create?
	TagType string `json:"tag_type" default:""`
}

func (t TagQuery) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// This is the primary interface used to interact with the Content and related models.
type ContentManager interface {
	GetCfg() *config.DirConfigEntry
	CanEdit() bool // Do we support CRUD or just R

	// Utility
	GetParams() *url.Values
	FindActualFile(mc *models.Content) (string, error)

	// Container Management
	GetContainer(cID int64) (*models.Container, error)
	ListContainers(cq ContainerQuery) (*models.Containers, int64, error)
	ListContainersFiltered(cq ContainerQuery) (*models.Containers, int64, error)
	ListContainersContext() (*models.Containers, int64, error)
	UpdateContainer(c *models.Container) (*models.Container, error)
	CreateContainer(c *models.Container) error
	DestroyContainer(id int64) (*models.Container, error)

	// Content listing (why did I name it Content vs Media?)
	GetContent(content_id int64) (*models.Content, error)
	ListContent(cs ContentQuery) (*models.Contents, int64, error)
	ListContentContext() (*models.Contents, int64, error)

	SearchContentContext() (*models.Contents, int64, error)
	SearchContent(cq ContentQuery) (*models.Contents, int64, error)
	SearchContainersContext() (*models.Containers, int64, error)
	SearchContainers(cs ContainerQuery) (*models.Containers, int64, error)

	UpdateContent(content *models.Content) error
	UpdateContents(content models.Contents) error
	DestroyContent(id int64) (*models.Content, error)
	CreateContent(mc *models.Content) error
	GetPreviewForMC(mc *models.Content) (string, error)

	// Functions that help with viewing movie screens if found.
	ListScreensContext() (*models.Screens, int64, error)
	ListScreens(sr ScreensQuery) (*models.Screens, int64, error)
	ClearScreens(content *models.Content) error

	GetScreen(psID int64) (*models.Screen, error)
	CreateScreen(s *models.Screen) error
	UpdateScreen(s *models.Screen) error
	DestroyScreen(id int64) (*models.Screen, error)

	// Tags listing (oy do I need to deal with this?)
	GetTag(id string) (*models.Tag, error)
	ListAllTags(tq TagQuery) (*models.Tags, int64, error)
	ListAllTagsContext() (*models.Tags, int64, error)
	CreateTag(tag *models.Tag) error
	UpdateTag(tag *models.Tag) error
	DestroyTag(id string) (*models.Tag, error)
	AssociateTag(tag *models.Tag, c *models.Content) error
	AssociateTagByID(tagID string, mcID int64) error

	// For processing encoding requests
	CreateTask(task *models.TaskRequest) (*models.TaskRequest, error)
	UpdateTask(task *models.TaskRequest, currentStatus models.TaskStatusType) (*models.TaskRequest, error)
	NextTask() (*models.TaskRequest, error) // Assigns it (not really required yet)

	// For the API exposed
	ListTasksContext() (*models.TaskRequests, int64, error)
	ListTasks(query TaskQuery) (*models.TaskRequests, int64, error)
	GetTask(id int64) (*models.TaskRequest, error)
}

// Grab a manager based on the gin context. This should probably be mo
func GetManager(c *gin.Context) ContentManager {
	cfg := config.GetCfg()

	// The connection should probably using both pooling and actual transactions when this is allo cleaned up.
	getConnection := GetConnection(cfg)
	// Annoying differences between url values and the param option.  Another GoLang bit where it isn't
	// handling ?id=1&id=2 and instead just allows for a single param (ie: The tags hack is still required in Gin)
	getParams := func() *url.Values {
		return GinParamsToUrlValues(c.Params, c.Request.URL.Query()) //c.Request.URL.Query())
	}
	return CreateManager(cfg, getConnection, getParams)
}

func GetManagerNoContext() ContentManager {
	cfg := config.GetCfg()
	get_params := func() *url.Values {
		return &url.Values{}
	}
	getConnection := GetConnection(cfg)
	return CreateManager(cfg, getConnection, get_params)
}

func GetConnection(cfg *config.DirConfigEntry) GetConnType {
	if cfg.UseDatabase {
		var conn *gorm.DB
		return func() *gorm.DB {
			if conn == nil {
				conn = models.InitGorm(false)
				if conn == nil || conn.Error != nil {
					log.Fatalf("Failed to get a db connection %s", conn.Error)
					return nil
				}
			}
			return conn
		}
	}
	// Just required for the memory version create statement
	return func() *gorm.DB {
		return nil
	}
}

func GinParamsToUrlValues(params gin.Params, queryValues url.Values) *url.Values {
	vals := url.Values{}
	for _, param := range params {
		vals[param.Key] = []string{param.Value}
	}
	for key, val := range queryValues {
		vals[key] = val
	}
	// Probably have to check that post body gets in here somehow as well.
	return &vals

}

// this is sketchy because of the connection scope closing on us
func GetAppManager(getConnection GetConnType) ContentManager {
	cfg := config.GetCfg()
	getParams := func() *url.Values {

		// Need to fix this up as well or change the worker model.
		return &url.Values{}
	}
	return CreateManager(cfg, getConnection, getParams)
}

// can this manager create, update or destroy
func ManagerCanCUD(c *gin.Context) (ContentManager, *gorm.DB, error) {
	man := GetManager(c)
	if !man.CanEdit() {
		err := errors.New("edit not supported by this manager")
		c.AbortWithError(http.StatusNotImplemented, err)
		return man, nil, err
	}
	if man.GetCfg().UseDatabase {
		db := models.InitGorm(false)
		if db.Error == nil {
			return man, db, nil
		}
		return man, nil, fmt.Errorf("DB Connection error %s", db.Error)
	}
	return man, nil, nil
}

// Create a manager based on the config, connection and params.
func CreateManager(cfg *config.DirConfigEntry, get_conn GetConnType, get_params GetParamsType) ContentManager {
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

func BoolDefault(s1 string, s2 bool) bool {
	if s1 == "true" {
		return true
	}
	return s2
}

// Used when doing pagination on the arrays of memory manager
func GetOffsetEnd(page int, per_page int, max int) (int, int) {
	if max == 0 {
		return 0, 0
	}

	if per_page <= 0 {
		per_page = config.DefaultLimit
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
func GetPagination(params *url.Values, DefaultLimit int) (int, int, int) {
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

func GetPerPage(perPage int) int {
	cfg := config.GetCfg()
	if perPage > cfg.Limit || perPage <= 0 {
		return cfg.Limit
	}
	return perPage
}

// TODO: Fill in tags if they are provided.
func ContextToContentQuery(params *url.Values, cfg *config.DirConfigEntry) ContentQuery {
	offset, per_page, page := GetPagination(params, cfg.Limit)
	sReq := ContentQuery{
		Text:          StringDefault(params.Get("text"), ""),
		Search:        StringDefault(params.Get("search"), ""),
		ContainerID:   StringDefault(params.Get("cId"), ""),
		ContentType:   StringDefault(params.Get("contentType"), ""),
		PerPage:       per_page,
		Page:          page,
		IncludeHidden: false,
		Order:         StringDefault(params.Get("order"), ""),
		Offset:        offset,
		Duplicate:     BoolDefault(params.Get("duplicate"), false),
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

// TODO: Fill in tags if they are provided.
func ContextToContainerQuery(params *url.Values, cfg *config.DirConfigEntry) ContainerQuery {
	offset, per_page, page := GetPagination(params, cfg.Limit)
	sReq := ContainerQuery{
		Name:          StringDefault(params.Get("name"), ""),
		Search:        StringDefault(params.Get("search"), ""),
		PerPage:       per_page,
		Page:          page,
		Offset:        offset,
		IncludeHidden: false,
		Order:         StringDefault(params.Get("order"), ""),
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

func GetContentAndContainer(cm ContentManager, contentID int64) (*models.Content, *models.Container, error) {
	content, cErr := cm.GetContent(contentID)
	if cErr != nil {
		return nil, nil, cErr
	}
	cnt, cntErr := cm.GetContainer(*content.ContainerID)
	if cntErr != nil {
		return nil, nil, cntErr
	}
	return content, cnt, nil
}

func CreateScreensForContent(cm ContentManager, contentID int64, count int, offset int) ([]string, string, error) {
	// It would be good to have the screens element take a few more params and have a wrapper on the
	// Content manager level.
	content, cnt, err := GetContentAndContainer(cm, contentID)
	if err != nil {
		return nil, "", err
	}
	path := cnt.GetFqPath()
	srcFile := filepath.Join(path, content.Src)
	dstPath := utils.GetPreviewDst(path)
	dstFile := utils.GetPreviewPathDestination(content.Src, dstPath, "video")

	log.Printf("Src file %s and Destination %s", srcFile, dstFile)
	utils.MakePreviewPath(dstPath)
	screens, ptrn, err := utils.CreateSeekScreens(srcFile, dstFile, count, offset)

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
	return screens, ptrn, err
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
	for int64(offset) < total {
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
	sr := ScreensQuery{ContentID: strconv.FormatInt(content.ID, 10)}
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
		sr := ContentQuery{Text: f.Name()}
		if originalContent.ContainerID != nil {
			sr.ContainerID = strconv.FormatInt(*originalContent.ContainerID, 10)
		}
		contents, _, err := man.SearchContent(sr)
		if err != nil {
			return nil, err
		}
		if contents != nil && len(*contents) == 1 {
			cnts := *contents
			return &cnts[0], nil
		}

		newId := utils.AssignNumerical(0, "contents")
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

/**
 * Remove the screens for a content element and also remove the files from disk.
 * Should this be "best effort" and still at least try and remove the screens even if the disk removes fail?
 */
func RemoveScreensForContent(man ContentManager, contentID int64) error {
	sr := ScreensQuery{ContentID: strconv.FormatInt(contentID, 10)}
	screens, _, err := man.ListScreens(sr)
	if err != nil {
		return err
	}
	if screens == nil || len(*screens) == 0 {
		return nil
	}

	// Might need to get the container in these cases... ugly issue
	content, err := man.GetContent(contentID)
	if err != nil {
		return err
	}
	_, err = man.GetContainer(*content.ContainerID)
	if err != nil {
		return err
	}

	/*
	 * DB it is more efficient to use the ClearScreens method but then we wouldn't have the actual
	 * disk removal...
	 */
	for _, screen := range *screens {
		_, err := man.DestroyScreen(screen.ID)
		if err != nil {
			return err
		}
		// It can be removed from disk already and we already check for IsNotExist
		removed, err := utils.RemoveFile(screen.Path, screen.Src, man.GetCfg().Dir)
		if !removed {
			log.Printf("Failed to remove file on disk for screen %d %s %s error %s", screen.ID, screen.Path, screen.Src, err)
		}
	}
	return nil
}

// Load all the contents in the container that are duplicates and destroy them.
func RemoveDuplicateContents(man ContentManager, cnt *models.Container) (int, error) {
	cs := ContentQuery{
		Duplicate:   true,
		PerPage:     9001,
		ContainerID: strconv.FormatInt(cnt.ID, 10),
	}

	// For now this is fine as paging doesn't use a cursor after each loop the search would no longer work, could just do a while
	contents, _, cErr := man.SearchContent(cs)
	if cErr != nil {
		return 0, cErr
	}
	if contents == nil {
		return 0, nil
	}

	// Destroy the content from the database or memory and remove it from the disk
	for _, content := range *contents {

		// The query SHOULD have filtered this but be paranoid in case search bugs out
		if !content.Duplicate {
			continue
		}
		_, dErr := man.DestroyContent(content.ID)
		if dErr != nil {
			return 0, fmt.Errorf("Failed to destroy content %s", dErr)
		}
		// Only move content IF the destroy worked and a removal location was configured
		RemoveContentFromContainer(man, &content, cnt)
	}
	return len(*contents), nil
}

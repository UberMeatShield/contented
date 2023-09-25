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
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

type GetConnType func() *pop.Connection
type GetParamsType func() *url.Values

// This is the primary interface used by the Buffalo actions.
type ContentManager interface {
	GetCfg() *utils.DirConfigEntry
	CanEdit() bool // Do we support CRUD or just R

	// Utility
	GetParams() *url.Values
	FindActualFile(mc *models.Content) (string, error)
	FindFileRef(mcID uuid.UUID) (*models.Content, error)

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
	SearchContent(search string, page int, per_page int, cId string, contentType string, includeHidden bool) (*models.Contents, int, error)
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
	tx, ok := ctx.Value("tx").(*pop.Connection)
	if !ok {
		return &man, nil, fmt.Errorf("No transaction found")
	}
	return &man, tx, nil
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

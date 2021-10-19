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
    "log"
    "errors"
    "sort"
    "net/url"
    "contented/models"
    "contented/utils"
    "strconv"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/pop/v5"
)


type GetConnType func() *pop.Connection
type GetParamsType func() *url.Values

// This is the primary interface used by the Buffalo actions.
type ContentManager interface {
    GetCfg() *utils.DirConfigEntry
    CanEdit() bool  // Do we support CRUD or just R

    GetParams() *url.Values

    FindFileRef(mc_id uuid.UUID) (*models.MediaContainer, error)

    GetContainer(c_id uuid.UUID) (*models.Container, error)
    ListContainers(page int, per_page int) (*models.Containers, error)
    ListContainersContext() (*models.Containers, error)

    GetMedia(media_id uuid.UUID) (*models.MediaContainer, error)
    ListMedia(ContainerID uuid.UUID, page int, per_page int) (*models.MediaContainers, error)
    ListMediaContext(ContainerID uuid.UUID) (*models.MediaContainers, error)
    ListAllMedia(page int, per_page int) (*models.MediaContainers, error)

    UpdateMedia(media *models.MediaContainer) error
    FindActualFile(mc *models.MediaContainer) (string, error)
    GetPreviewForMC(mc *models.MediaContainer) (string, error)
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
        mem_man.Initialize()  // Break this into a sensible initialization
        return mem_man
    }
}

// Provides the support for looking up media by ID while only using memory
type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry

    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    validate string 

    params *url.Values
    Params GetParamsType
}

// We do not allow editing in a memory manager
func (cm ContentManagerMemory) CanEdit() bool {
    return false;
}

// Provide the ability to set the configuration for a memory manager.
func (cm *ContentManagerMemory) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
    log.Printf("It should have a preview %s\n", cm.cfg.Dir)
    log.Printf("Using memory manager %s\n", cm.validate)
}

// Get the currently configuration for this manager.
func (cm ContentManagerMemory) GetCfg() *utils.DirConfigEntry {
    log.Printf("Get the Config validate val %s\n", cm.validate)
    log.Printf("Config is using path %s", cm.cfg.Dir)
    return cm.cfg
}

// On a first time load / use we will pull back content information from dist and from
// then on continue to use already loaded information.
func (cm *ContentManagerMemory) Initialize() {
    
    // TODO: Should we allow for a timeout or rescan option?
    memStorage := utils.GetMemStorage()
    if memStorage.Initialized == false {
        memStorage = utils.InitializeMemory(cm.cfg.Dir)
    }
    cm.ValidContainers = memStorage.ValidContainers
    cm.ValidMedia = memStorage.ValidMedia
    log.Printf("Found %d directories with %d media elements \n", len(cm.ValidContainers), len(cm.ValidMedia))
}

// Kinda strange but it seems hard to assign the type into an interface
// type GetParamsType func() *url.Values
func (cm ContentManagerMemory) GetParams() *url.Values {
    return cm.Params()
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

func (cm ContentManagerMemory) ListMediaContext(c_id uuid.UUID) (*models.MediaContainers, error) {
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListMedia(c_id, page, limit)
}

func (cm ContentManagerMemory) ListAllMedia(page int, per_page int) (*models.MediaContainers, error) {
    m_arr := models.MediaContainers{}
    offset := (page - 1) * per_page
    for _, m := range cm.ValidMedia {
        m_arr = append(m_arr, m)
    }
    sort.SliceStable(m_arr, func(i, j int) bool {
        return m_arr[i].Idx < m_arr[j].Idx
    })
    offset, end := GetOffsetEnd(page, per_page, len(m_arr))
    if end > 0 {  // If it is empty a slice ending in 0 = boom
        m_arr = m_arr[offset : end]
        return &m_arr, nil
    }
    return &m_arr, nil
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListMedia(ContainerID uuid.UUID, page int, per_page int) (*models.MediaContainers, error) {
    m_arr := models.MediaContainers{}
    for _, m := range cm.ValidMedia {
        if m.ContainerID.Valid && m.ContainerID.UUID == ContainerID {
            m_arr = append(m_arr, m)
        }
    }
    sort.SliceStable(m_arr, func(i, j int) bool {
        return m_arr[i].Idx < m_arr[j].Idx
    })
    offset, end := GetOffsetEnd(page, per_page, len(m_arr))
    if end > 0 {  // If it is empty a slice ending in 0 = boom
        m_arr = m_arr[offset : end]
        return &m_arr, nil
    }
    log.Printf("Get a list of media offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
    return &m_arr, nil
}

// Get a media element by the ID
func (cm ContentManagerMemory) GetMedia(mc_id uuid.UUID) (*models.MediaContainer, error) {
    log.Printf("Memory Get a single media %s", mc_id)
    if mc, ok := cm.ValidMedia[mc_id]; ok {
        return &mc, nil
    }
    return nil, errors.New("Media was not found in memory")
}

// No updates should be allowed for memory management.
func (cm ContentManagerMemory) UpdateMedia(media *models.MediaContainer) error {
    return errors.New("Updates are not allowed for in memory management")
}

// Given the current parameters in the buffalo context return a list of matching containers.
func (cm ContentManagerMemory) ListContainersContext() (*models.Containers, error) {
    _, per_page, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListContainers(page, per_page)
}

// Actually list containers using a page and per_page which is consistent with buffalo standard pagination
func (cm ContentManagerMemory) ListContainers(page int, per_page int) (*models.Containers, error) {
    log.Printf("List Containers with page(%d) and per_page(%d)", page, per_page)

    c_arr := models.Containers{}
    for _, c := range cm.ValidContainers {
        c_arr = append(c_arr, c)
    }
    sort.SliceStable(c_arr, func(i, j int) bool {
        return c_arr[i].Idx < c_arr[j].Idx
    })

    offset, end := GetOffsetEnd(page, per_page, len(c_arr))
    c_arr = c_arr[offset : end]
    return &c_arr, nil
}

// Get a single container given the primary key
func (cm ContentManagerMemory) GetContainer(c_id uuid.UUID) (*models.Container, error) {
    log.Printf("Get a single container %s", c_id)
    if c, ok := cm.ValidContainers[c_id]; ok {
        return &c, nil
    }
    return nil, errors.New("Memory manager did not find this container id: " + c_id.String())
}

func (cm ContentManagerMemory) FindFileRef(mc_id uuid.UUID) (*models.MediaContainer, error) {
    if mc, ok := cm.ValidMedia[mc_id]; ok {
        return &mc, nil
    }
    return nil, errors.New("File was not found in the current list of files")
}

func (cm ContentManagerMemory) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
    dir, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        return "Memory Manager Preview no Parent Found", err
    }
    src := mc.Src
    if mc.Preview != "" {
        src = mc.Preview
    }
    log.Printf("Memory Manager loading %s preview %s\n", mc.ID.String(), src)
    return utils.GetFilePathInContainer(src, dir.Name)
}

func (cm ContentManagerMemory) FindActualFile(mc *models.MediaContainer) (string, error) {
    dir, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        return "Memory Manager View no Parent Found", err
    }
    log.Printf("Memory Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
    return utils.GetFilePathInContainer(mc.Src, dir.Name)
}

// If you want to do in memory testing and already manually created previews this will
// then try and use the previews for the in memory manager.
func (cm ContentManagerMemory) SetPreviewIfExists(mc *models.MediaContainer) (string, error) {
    c, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        log.Fatal(err)
        return "", err
    }
    pFile := utils.AssignPreviewIfExists(c, mc)
    return pFile, nil
}

// DB version of content management
type ContentManagerDB struct {
    cfg *utils.DirConfigEntry
    c *buffalo.Context

    /* Is this even useful ? */
    conn *pop.Connection
    params *url.Values

    GetConnection GetConnType  // Returns .conn or context.Value(tx)
    Params GetParamsType    // returns .params or context.Params()
}


// This is a little sketchy that the two are not directly linked
func (cm *ContentManagerDB) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
}
func (cm ContentManagerDB) GetCfg() *utils.DirConfigEntry {
    return cm.cfg
}

func (cm ContentManagerDB) GetParams() *url.Values {
    return cm.Params()
}

func (cm ContentManagerDB) CanEdit() bool {
    return true;
}


func (cm ContentManagerDB) ListMediaContext(c_id uuid.UUID) (*models.MediaContainers, error) {
    // Could add the context here correctly
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListMedia(c_id, page, limit)
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerDB) ListMedia(c_id uuid.UUID, page int, per_page int) (*models.MediaContainers, error) {
    log.Printf("Get a list of media from DB, we should have some %s", c_id.String())
    tx := cm.GetConnection()
    mediaContainers := &models.MediaContainers{}

    // Paginate results. Params "page" and "per_page" control pagination.
    // Default values are "page=1" and "per_page=20".
    // TODO: Make it paginate using the params not the context
    q := tx.Paginate(page, per_page)
    q_conn := q.Where("container_id = ?", c_id)
    if q_err := q_conn.All(mediaContainers); q_err != nil {
        return nil, q_err
    }

    return mediaContainers, nil
}

func (cm ContentManagerDB) GetMedia(mc_id uuid.UUID) (*models.MediaContainer, error) {
    log.Printf("Get a single media %s", mc_id)
    tx := cm.GetConnection()
    container := &models.MediaContainer{}
    if err := tx.Find(container, mc_id); err != nil {
        return nil, err
    }
    return container, nil
}

func (cm ContentManagerDB) UpdateMedia(media *models.MediaContainer) error {
    tx := cm.GetConnection()
    return tx.Update(media)
}

func (cm ContentManagerDB) ListAllMedia(page int, per_page int) (*models.MediaContainers, error) {
    log.Printf("List all media DB manager")
    tx := cm.GetConnection()
    q := tx.Paginate(page, per_page)
    mediaContainers := &models.MediaContainers{}
    if err := q.All(mediaContainers); err != nil {
        return nil, err
    }
    return mediaContainers, nil
}

// The default list using the current manager configuration
func (cm ContentManagerDB) ListContainersContext() (*models.Containers, error) {
    return cm.ListContainers(1, cm.cfg.Limit)
}

// TODO: Add in support for actually doing the query using the current buffalo.Context
func (cm ContentManagerDB) ListContainers(page int, per_page int) (*models.Containers, error) {
    log.Printf("DB List all containers")
    tx := cm.GetConnection()
    q := tx.Paginate(page, per_page)

    // Retrieve all Containers from the DB
    containers := &models.Containers{}
    if err := q.All(containers); err != nil {
        return nil, err
    }
    return containers, nil
}

func (cm ContentManagerDB) GetContainer(mc_id uuid.UUID) (*models.Container, error) {
    log.Printf("Get a single container %s", mc_id)
    tx := cm.GetConnection()
    p := cm.Params()

    // Allocate an empty Container
    container := &models.Container{}
    if err := tx.Find(container, p.Get("container_id")); err != nil {
        return nil, err
    }
    return container, nil
}

func (cm *ContentManagerDB) Initialize() {
    // Connect to the DB using the context or some other option?
    log.Printf("Make a DB connection here")
}

func (cm ContentManagerDB) FindFileRef(mc_id uuid.UUID) (*models.MediaContainer, error) {
    mc_db := models.MediaContainer{}
    err := models.DB.Find(&mc_db, mc_id)
    if err == nil {
        return &mc_db, nil
    }
    return nil, err
}

func (cm ContentManagerDB) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
    dir, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        return "DB Manager Preview no Parent Found", err
    }
    src := mc.Src
    if mc.Preview != "" {
        src = mc.Preview
    }
    log.Printf("DB Manager loading %s preview %s\n", mc.ID.String(), src)
    return utils.GetFilePathInContainer(src, dir.Name)
}

func (cm ContentManagerDB) FindActualFile(mc *models.MediaContainer) (string, error) {
    dir, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        return "DB Manager View no Parent Found", err
    }
    log.Printf("DB Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
    return utils.GetFilePathInContainer(mc.Src, dir.Name)
}

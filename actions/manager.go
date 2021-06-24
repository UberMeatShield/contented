package actions

import (
    "log"
    "os"
    "errors"
    "sort"
    "net/url"
    "path/filepath"
    "contented/models"
    "contented/utils"
    "strconv"
    //"github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/pop/v5"
    //"github.com/gobuffalo/pop/v5/paginator"
    //"github.com/gobuffalo/pop/v5/defaults"
)


type GetConnType func() *pop.Connection
type GetParamsType func() *url.Values

// Dealing with buffalo.Context vs grift.Context is kinda annoying, this handles the
// buffalo context which handles tests or runtime but doesn't work in grifts.
func GetManager(c *buffalo.Context) ContentManager {
    cfg := utils.GetCfg()
    ctx := *c

    // The get connection might need an async channel or it potentially locks
    // the dev server :(.   Need to only do this if use database is set
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

func CreateManager(cfg *utils.DirConfigEntry, get_conn GetConnType, get_params GetParamsType) ContentManager {
    // Not really important for a DB manager, just need to look at it
    if cfg.UseDatabase {
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

type ContentManager interface {
    // SetCfg(c *utils.DirConfigEntry)
    GetCfg() *utils.DirConfigEntry
    CanEdit() bool  // Do we support CRUD or just R

    // TODO
    //GetConnection() *pop.Connection
    GetParams() *url.Values

    FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error)
    FindDirRef(dir_id uuid.UUID) (*models.Container, error)

    GetContainer(c_id uuid.UUID) (*models.Container, error)
    ListContainers(page int, per_page int) (*models.Containers, error)
    ListContainersContext() (*models.Containers, error)

    GetMedia(media_id uuid.UUID) (*models.MediaContainer, error)
    ListMedia(ContainerID uuid.UUID, page int, per_page int) (*models.MediaContainers, error)
    ListMediaContext(ContainerID uuid.UUID) (*models.MediaContainers, error)
    ListAllMedia(page int, per_page int) (*models.MediaContainers, error)

    FindActualFile(mc *models.MediaContainer) (string, error)
    GetPreviewForMC(mc *models.MediaContainer) (string, error)
}


// GoLang is just making this awkward
type MemoryStorage struct {
    Initialized bool
    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
}
var memStorage MemoryStorage = MemoryStorage{Initialized: false}

type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry

    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    validate string 

    params *url.Values
    Params GetParamsType    // returns .params or context.Params()
}

func (cm ContentManagerMemory) CanEdit() bool {
    return false;
}

func (cm *ContentManagerMemory) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
    log.Printf("It should have a preview %s\n", cm.cfg.Dir)
    log.Printf("Using memory manager %s\n", cm.validate)
}

func (cm ContentManagerMemory) GetCfg() *utils.DirConfigEntry {
    log.Printf("Get the Config Validate val %s\n", cm.validate)
    log.Printf("Config is what %s", cm.cfg.Dir)
    return cm.cfg
}



// TODO:  Now THIS one can actually build out a singleton that is shared I guess.
func (cm *ContentManagerMemory) Initialize() {

    // We only want the memory storage initialized one time ?  But allow for re-init?
    // Could toss the object into the manager but then that means even more code change.
    if memStorage.Initialized == false {
        memStorage = InitializeMemory(cm.cfg.Dir)
    }
    // Move some of these into an actual singleton.
    cm.ValidContainers = memStorage.ValidContainers
    cm.ValidMedia = memStorage.ValidMedia
    log.Printf("Found %d directories\n", len(cm.ValidContainers))
}

func InitializeMemory(dir_root string) MemoryStorage {
    log.Printf("Initializing Memory Storage %s\n", dir_root)
    containers, files := PopulateMemoryView(dir_root)

    memStorage.Initialized = true
    memStorage.ValidContainers = containers
    memStorage.ValidMedia = files

    return memStorage
}

// Kinda strange but it seems hard to assign the type into an interface
// type GetParamsType func() *url.Values
func (cm ContentManagerMemory) GetParams() *url.Values {
    return cm.Params()
}

/**
 *  TODO:  Require the number to preview and will be Memory only supported.
 */
func PopulateMemoryView(dir_root string) (models.ContainerMap, models.MediaMap) {
    containers := models.ContainerMap{}
    files := models.MediaMap{}

    log.Printf("Searching in %s", dir_root)
    cfg := utils.GetCfg()

    cnts := utils.FindContainers(dir_root)
    for idx, c := range cnts {
        media := utils.FindMediaMatcher(c, 90001, 0, cfg.IncFiles, cfg.ExcFiles)  // TODO: Config this
        // c.Contents = media
        c.Total = len(media)
        c.Idx = idx
        containers[c.ID] = c

        for _, mc := range media {
            files[mc.ID] = mc
        }
    }
    return containers, files
}

func StringDefault(s1, s2 string) string {
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
    m_arr = m_arr[offset : end]
    return &m_arr, nil
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListMedia(ContainerID uuid.UUID, page int, per_page int) (*models.MediaContainers, error) {

    // TODO: Port to using the containers
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
    m_arr = m_arr[offset : end]
    log.Printf("Get a list of media offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
    return &m_arr, nil
}

func (cm ContentManagerMemory) GetMedia(mc_id uuid.UUID) (*models.MediaContainer, error) {
    log.Printf("Memory Get a single media %s", mc_id)
    if mc, ok := cm.ValidMedia[mc_id]; ok {
        return &mc, nil
    }
    return nil, errors.New("Media was not found in memory")
}

func (cm ContentManagerMemory) ListContainersContext() (*models.Containers, error) {
    _, per_page, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListContainers(page, per_page)
}

func (cm ContentManagerMemory) ListContainers(page int, per_page int) (*models.Containers, error) {
    log.Printf("List Containers with page(%d) and per_page(%d)", page, per_page)

    // TODO: Maybe just actually store the array on this
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

func (cm ContentManagerMemory) GetContainer(c_id uuid.UUID) (*models.Container, error) {
    log.Printf("Get a single container %s", c_id)
    if c, ok := cm.ValidContainers[c_id]; ok {
        return &c, nil
    }
    return nil, errors.New("Memory manager did not find this id" + c_id.String())
}

func (cm ContentManagerMemory) FindDirRef(dir_id uuid.UUID) (*models.Container, error) {
    // TODO: Get a FileInfo reference (get parent dir too)
    if d, ok := cm.ValidContainers[dir_id]; ok {
        return &d, nil
    }
    return nil, errors.New("Directory was not found in the current app")
}

func (cm ContentManagerMemory) FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error) {
    if mc, ok := cm.ValidMedia[file_id]; ok {
        return &mc, nil
    }
    return nil, errors.New("File was not found in the current list of files")
}

func (cm ContentManagerMemory) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
    dir, err := cm.FindDirRef(mc.ContainerID.UUID)
    if err != nil {
        return "DB Manager Preview no Parent Found", err
    }
    src := mc.Src
    if mc.Preview != "" {
        src = mc.Preview
    }
    log.Printf("DB Manager loading %s preview %s\n", mc.ID.String(), src)
    return GetFilePathInContainer(src, dir.Name)
}

func (cm ContentManagerMemory) FindActualFile(mc *models.MediaContainer) (string, error) {
    dir, err := cm.FindDirRef(mc.ContainerID.UUID)
    if err != nil {
        return "Memory Manager View no Parent Found", err
    }
    log.Printf("Memory Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
    return GetFilePathInContainer(mc.Src, dir.Name)
}


// DB version of content management
type ContentManagerDB struct {
    cfg *utils.DirConfigEntry
    c *buffalo.Context

    /* Is this even useful ? */
    conn *pop.Connection
    params *url.Values

    // hate
    GetConnection GetConnType  // Returns .conn or context.Value(tx)
    Params GetParamsType    // returns .params or context.Params()
}

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

func (cm ContentManagerDB) FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error) {
    mc_db := models.MediaContainer{}
    err := models.DB.Find(&mc_db, file_id)
    if err == nil {
        return &mc_db, nil
    }
    return nil, err
}

func (cm ContentManagerDB) FindDirRef(dir_id uuid.UUID) (*models.Container, error) {
    c_db := models.Container{}
    err := models.DB.Find(&c_db, dir_id)
    if err == nil {
        return &c_db, nil
    }
    return nil, err
}

func (cm ContentManagerDB) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
    dir, err := cm.FindDirRef(mc.ContainerID.UUID)
    if err != nil {
        return "DB Manager Preview no Parent Found", err
    }
    src := mc.Src
    if mc.Preview != "" {
        src = mc.Preview
    }
    log.Printf("DB Manager loading %s preview %s\n", mc.ID.String(), src)
    return GetFilePathInContainer(src, dir.Name)
}

func (cm ContentManagerDB) FindActualFile(mc *models.MediaContainer) (string, error) {
    dir, err := cm.FindDirRef(mc.ContainerID.UUID)
    if err != nil {
        return "DB Manager View no Parent Found", err
    }
    log.Printf("DB Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
    return GetFilePathInContainer(mc.Src, dir.Name)
}

// Given a container ID and the src of a file in there, get a path and check if it exists
func GetFilePathInContainer(src string, dir_name string) (string, error) {
    cfg := utils.GetCfg()
    path := filepath.Join(cfg.Dir, dir_name)
    fq_path := filepath.Join(path, src)
    if _, os_err := os.Stat(fq_path); os_err != nil {
        return fq_path, os_err
    }
    return fq_path, nil
}

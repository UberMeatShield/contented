package actions

import (
    "log"
    "os"
    "errors"
    "path/filepath"
    "contented/models"
    "contented/utils"
    // "github.com/gobuffalo/nulls"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/buffalo"
)

// HomeHandler is a default handler to serve up
var DefaultLimit int = 10000 // The max limit set by environment variable
var DefaultPreviewCount int = 8

// https://medium.com/@TobiasSchmidt89/the-singleton-object-oriented-design-pattern-in-golang-9f6ce75c21f7

var appManager ContentManager
func GetManager() ContentManager {
    return appManager
}
func SetManager(manager ContentManager) {
    appManager = manager
}

// TODO: Remove all the old endpoints and endpoint tests
// TODO: Add in a few more tests around data loading / pagination
// TODO: Create an in memory version and have that be setup on app initialization
// TODO: Fix the UI to actually load in the data
// TODO: Add in index code into the data model
// TODO: Make a manager have a config entry vs what I currently do
var appCfg utils.DirConfigEntry = utils.DirConfigEntry{
    Initialized:  false,
    Dir:          "",
    PreviewCount: DefaultPreviewCount,
    Limit:        DefaultLimit,
}
func GetCfg() *utils.DirConfigEntry {
    return &appCfg
}
func SetCfg(c utils.DirConfigEntry) {
    appCfg = c
}

type ContentManager interface {
    // SetCfg(c *utils.DirConfigEntry)
    GetCfg() *utils.DirConfigEntry

    FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error)
    FindDirRef(dir_id uuid.UUID) (*models.Container, error)

    GetContainer(c_id uuid.UUID) *models.Container
    ListContainers(page int, per_page int) *models.Containers
    ListContainersContext() *models.Containers

    GetMedia(media_id uuid.UUID) *models.MediaContainer
    ListMedia(container_id uuid.UUID, page int, per_page int) *models.MediaContainers
    ListMediaContext(container_id uuid.UUID) *models.MediaContainers
}


type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry
    c buffalo.Context

    ValidDirs       map[string]os.FileInfo
    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    validate string 
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

func (cm *ContentManagerMemory) Initialize() {
    dir_root := cm.cfg.Dir
    log.Printf("Initializing Memory manager %s\n", dir_root)

    dir_lookup := utils.GetDirectoriesLookup(dir_root)

    // Move some of these
    containers, files := utils.PopulatePreviews(dir_root, dir_lookup)
    cm.ValidDirs = dir_lookup
    cm.ValidContainers = containers
    cm.ValidMedia = files
    log.Printf("Found %d directories\n", len(cm.ValidDirs))
}

func (cm ContentManagerMemory) ListMediaContext(c_id uuid.UUID) *models.MediaContainers{
    return cm.ListMedia(c_id, int(1), cm.cfg.Limit)
}
// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListMedia(container_id uuid.UUID, page int, per_page int) *models.MediaContainers{

    m_arr := models.MediaContainers{}
    for _, m := range cm.ValidMedia {
        if m.ContainerID.Valid {
            if m.ContainerID.UUID == container_id && len(m_arr) < (per_page + 1) {
                m_arr = append(m_arr, m)
            }
        } else if len(m_arr) < (per_page + 1) {
            m_arr = append(m_arr, m)
        }
    }
    log.Printf("Get a list of media, we should have some %d", len(m_arr))
    return &m_arr
}

func (cm ContentManagerMemory) GetMedia(mc_id uuid.UUID) *models.MediaContainer {
    log.Printf("Get a single media %s", mc_id)
    if mc, ok := cm.ValidMedia[mc_id]; ok {
        return &mc
    }
    return nil
}

func (cm ContentManagerMemory) ListContainersContext() *models.Containers {
    return cm.ListContainers(int(1), cm.cfg.Limit)
}

func (cm ContentManagerMemory) ListContainers(page int, per_page int) *models.Containers {
    log.Printf("List Containers")
    c_arr := models.Containers{}
    for _, c := range cm.ValidContainers {
        c_arr = append(c_arr, c)
    }
    return &c_arr
}

func (cm ContentManagerMemory) GetContainer(c_id uuid.UUID) *models.Container {
    log.Printf("Get a single contaienr %s", c_id)
    if c, ok := cm.ValidContainers[c_id]; ok {
        return &c
    }
    return nil
}

func (cm ContentManagerMemory) FindDirRef(dir_id uuid.UUID) (*models.Container, error) {
    // TODO: Get a FileInfo reference (get parent dir too)
    if d, ok := appCfg.ValidContainers[dir_id]; ok {
        return &d, nil
    }
    return nil, errors.New("Directory was not found in the current app")
}

func (cm ContentManagerMemory) FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error) {
    if mc, ok := appCfg.ValidFiles[file_id]; ok {
        return &mc, nil
    }
    return nil, errors.New("File was not found in the current list of files")
}


// DB version of content management
type ContentManagerDB struct {
    cfg *utils.DirConfigEntry
    c buffalo.Context
}

func (cm *ContentManagerDB) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
}
func (cm ContentManagerDB) GetCfg() *utils.DirConfigEntry {
    return cm.cfg
}

func (cm ContentManagerDB) ListMediaContext(c_id uuid.UUID) *models.MediaContainers {
    return cm.ListMedia(c_id, 1, cm.cfg.Limit)
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerDB) ListMedia(c_id uuid.UUID, page int, per_page int) *models.MediaContainers {
    log.Printf("Get a list of media from DB, we should have some %s", c_id.String())
    mediaContainers := &models.MediaContainers{}
    return mediaContainers
}

func (cm ContentManagerDB) GetMedia(mc_id uuid.UUID) *models.MediaContainer {
    log.Printf("Get a single media %s", mc_id)
    mediaContainer := &models.MediaContainer{}
    return mediaContainer
}

// The default list using the current manager configuration
func (cm ContentManagerDB) ListContainersContext() *models.Containers {
    return cm.ListContainers(1, cm.cfg.Limit)
}

// TODO: Add in support for actually doing the query using the current buffalo.Context
func (cm ContentManagerDB) ListContainers(page int, per_page int) *models.Containers {
    log.Printf("List Containers")
    container := &models.Containers{}
    return container
}

func (cm ContentManagerDB) GetContainer(mc_id uuid.UUID) *models.Container {
    log.Printf("Get a single contaienr %s", mc_id)
    container := &models.Container{}
    return container
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

// Store a list of the various file references
func FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error) {
    // TODO: Get a FileInfo reference (get parent dir too)
    if mc, ok := appCfg.ValidFiles[file_id]; ok {
        return &mc, nil
    }

    // Fallback to the DB
    mc_db := models.MediaContainer{}
    err := models.DB.Find(&mc_db, file_id)
    if err == nil {
        return &mc_db, nil
    }
    return nil, err
}

func FindDirRef(dir_id uuid.UUID) (*models.Container, error) {
    if d, ok := appCfg.ValidContainers[dir_id]; ok {
        return &d, nil
    }

    // Sketchy DB fallback  HATE
    c_db := models.Container{}
    err := models.DB.Find(&c_db, dir_id)
    if err == nil {
        return &c_db, nil
    }
    return nil, err
}

// If a preview is found, return the path to that file otherwise use the actual file
func GetPreviewForMC(mc *models.MediaContainer) (string, error) {
    src := mc.Src
    if mc.Preview != "" {
        src = mc.Preview
    }
    log.Printf("It should have a preview %s\n", mc.Preview)
    return GetFilePathInContainer(mc.ContainerID.UUID, src)
}

// Get the on disk location for the media container.
func FindActualFile(mc *models.MediaContainer) (string, error) {
    return GetFilePathInContainer(mc.ContainerID.UUID, mc.Src)
}

// Given a container ID and the src of a file in there, get a path and check if it exists
func GetFilePathInContainer(cont_id uuid.UUID, src string) (string, error) {
    dir, err := FindDirRef(cont_id)
    if err != nil {
        return "No Parent Found", err
    }
    path := filepath.Join(appCfg.Dir, dir.Name)
    fq_path := filepath.Join(path, src)

    if _, os_err := os.Stat(fq_path); os_err != nil {
        return fq_path, os_err
    }
    return fq_path, nil
}
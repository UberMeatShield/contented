package actions

import (
    "log"
    "os"
    "errors"
    "path/filepath"
    "contented/models"
    "contented/utils"
    "github.com/gofrs/uuid"
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

    // GetOneContainer
    // GetOneMediaContainer
    // ListContainers
    // ListMedia
}


type ContentManagerDB struct {
    cfg *utils.DirConfigEntry
}

type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry
    validate string 
}

func (cm *ContentManagerDB) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
}
func (cm ContentManagerDB) GetCfg() *utils.DirConfigEntry {
    return cm.cfg
}

func (cm *ContentManagerMemory) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
    log.Printf("It should have a preview %s\n", cm.cfg.Dir)
    log.Printf("Using memory manager %s\n", cm.validate)
}
func (cm ContentManagerMemory) GetCfg() *utils.DirConfigEntry {
    log.Printf("Get the Config %s\n", cm.validate)
    log.Printf("Config is what %s", cm.cfg.Dir)
    return cm.cfg
}

// Have this used by the resources?
func (cm ContentManagerDB) FindFileRef(file_id uuid.UUID) (*models.MediaContainer, error) {
    // Fallback to the DB
    mc_db := models.MediaContainer{}
    err := models.DB.Find(&mc_db, file_id)
    if err == nil {
        return &mc_db, nil
    }
    return nil, err
}

func (cm ContentManagerDB) FindDirRef(dir_id uuid.UUID) (*models.Container, error) {
    // Sketchy DB fallback  HATE
    c_db := models.Container{}
    err := models.DB.Find(&c_db, dir_id)
    if err == nil {
        return &c_db, nil
    }
    return nil, err
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

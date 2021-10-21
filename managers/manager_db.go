/**
* Implements the ContentManager interface and stores information in a postgres db.
*/
package managers

import (
    "log"
    "net/url"
    "contented/models"
    "contented/utils"
    "github.com/gofrs/uuid"
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/pop/v5"
)

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

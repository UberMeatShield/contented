/**
* Implements the ContentManager interface and stores information in a postgres db.
 */
package managers

import (
	"contented/models"
	"contented/utils"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"log"
	"net/url"
)

// DB version of content management
type ContentManagerDB struct {
	cfg *utils.DirConfigEntry
	c   *buffalo.Context

	/* Is this even useful ? */
	conn   *pop.Connection
	params *url.Values

	GetConnection GetConnType   // Returns .conn or context.Value(tx)
	Params        GetParamsType // returns .params or context.Params()
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
	return true
}

func (cm ContentManagerDB) ListMediaContext(cID uuid.UUID) (*models.MediaContainers, error) {
	// Could add the context here correctly
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListMedia(cID, page, limit)
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerDB) ListMedia(cID uuid.UUID, page int, per_page int) (*models.MediaContainers, error) {
	log.Printf("Get a list of media from DB, we should have some %s", cID.String())
	tx := cm.GetConnection()
	mediaContainers := &models.MediaContainers{}

	// Paginate results. Params "page" and "per_page" control pagination.
	q := tx.Paginate(page, per_page)
	q_conn := q.Where("container_id = ?", cID)
	if q_err := q_conn.All(mediaContainers); q_err != nil {
		return nil, q_err
	}
	return mediaContainers, nil
}

func (cm ContentManagerDB) GetMedia(mcID uuid.UUID) (*models.MediaContainer, error) {
	log.Printf("Get a single media %s", mcID)
	tx := cm.GetConnection()
	container := &models.MediaContainer{}
	if err := tx.Find(container, mcID); err != nil {
		return nil, err
	}
	return container, nil
}

func (cm ContentManagerDB) UpdateContainer(c *models.Container) error {
	tx := cm.GetConnection()
	return tx.Update(c)
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

// It should probably be able to search the container too?
func (cm ContentManagerDB) SearchMediaContext() (*models.MediaContainers, int, error) {
	params := cm.Params()
	_, per_page, page := GetPagination(params, cm.cfg.Limit)
	searchStr := StringDefault(params.Get("text"), "")
	cId := StringDefault(params.Get("cId"), "")
	return cm.SearchMedia(searchStr, page, per_page, cId)
}

func (cm ContentManagerDB) SearchMedia(search string, page int, per_page int, cId string) (*models.MediaContainers, int, error) {
	mediaContainers := &models.MediaContainers{}
	tx := cm.GetConnection()

	// TODO: We need to do a count first and preserve that :(  hate
	q := tx.Paginate(page, per_page)
	if search != "*" && search != "" {
		search = ("%" + search + "%")
		q = q.Where(`src like ?`, search)
	}
	if cId != "" {
		q = q.Where(`container_id = ?`, cId)
	}
	count, _ := q.Count(&models.MediaContainers{})
	log.Printf("Total count of search media %d", count)

	if q_err := q.All(mediaContainers); q_err != nil {
		return mediaContainers, count, q_err
	}
	return mediaContainers, count, nil
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

// TODO: Need a preview test using the database where we do NOT have a preview created
func (cm ContentManagerDB) GetContainer(cID uuid.UUID) (*models.Container, error) {
	log.Printf("Get a single container %s", cID)
	tx := cm.GetConnection()

	// Allocate an empty Container p := cm.Params()
	container := &models.Container{}
	if err := tx.Find(container, cID); err != nil {
		return nil, err
	}
	return container, nil
}

func (cm *ContentManagerDB) Initialize() {
	// Connect to the DB using the context or some other option?
	log.Printf("Make a DB connection here")
}

func (cm ContentManagerDB) FindFileRef(mcID uuid.UUID) (*models.MediaContainer, error) {
	mc_db := models.MediaContainer{}
	err := models.DB.Find(&mc_db, mcID)
	if err == nil {
		return &mc_db, nil
	}
	return nil, err
}

func (cm ContentManagerDB) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
	cnt, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		return "DB Manager Preview no Parent Found", err
	}
	src := mc.Src
	if mc.Preview != "" {
		src = mc.Preview
	}
	log.Printf("DB Manager loading %s preview %s\n", mc.ID.String(), src)
	return utils.GetFilePathInContainer(src, cnt.GetFqPath())
}

func (cm ContentManagerDB) FindActualFile(mc *models.MediaContainer) (string, error) {
	cnt, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		return "DB Manager View no Parent Found", err
	}
	log.Printf("DB Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
	return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

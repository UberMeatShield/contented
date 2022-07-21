/**
* Implements the ContentManager interface and stores information in a postgres db.
 */
package managers

import (
	"log"
    "strings"
	"net/url"
	"contented/models"
	"contented/utils"
    "github.com/lib/pq"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
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

func (cm ContentManagerDB) UpdateScreen(s *models.PreviewScreen) error {
    tx := cm.GetConnection()
    return tx.Update(s)
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
	contentType := StringDefault(params.Get("contentType"), "")
	return cm.SearchMedia(searchStr, page, per_page, cId, contentType)
}

func (cm ContentManagerDB) SearchMedia(search string, page int, per_page int, cId string, contentType string) (*models.MediaContainers, int, error) {
	mediaContainers := &models.MediaContainers{}
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	if search != "*" && search != "" {
		search = ("%" + search + "%")
		q = q.Where(`src like ?`, search)
	}
	if contentType != "" {
        contentType = ("%" + contentType + "%")
		q = q.Where(`content_type ilike ?`, contentType)
    }
	if cId != "" {
		q = q.Where(`container_id = ?`, cId)
	}
	count, _ := q.Count(&models.MediaContainers{})
	log.Printf("Total count of search media %d using search (%s) and contentType (%s)", count, search, contentType) 

    // TODO: should grab all the screens associated with any media components
	if q_err := q.All(mediaContainers); q_err != nil {
		return mediaContainers, count, q_err
	}
	return mediaContainers, count, nil
}

func (cm ContentManagerDB) LoadRelatedScreens(media *models.MediaContainers) (*models.PreviewScreens, error) {
    if media == nil || len(*media) == 0{
        return nil, nil
    }
    videoIds := []string{}
    for _, mc := range *media {
        if strings.Contains(mc.ContentType, "video") {
            videoIds = append(videoIds, mc.ID.String())
        }
    }
    if len(videoIds) == 0 {
        log.Printf("None of these media were a video, skipping")
        return nil, nil
    }
	q := cm.GetConnection().Q().Where(`media_container_id = any($1)`, pq.Array(videoIds))
    screens := &models.PreviewScreens{}
	if q_err := q.All(screens); q_err != nil {
        log.Printf("Error loading video screens %s", q_err)
        return nil, q_err
    }
	return screens, nil
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


func (cm ContentManagerDB) ListAllScreensContext() (*models.PreviewScreens, error) {
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListAllScreens(page, limit)
}


func (cm ContentManagerDB) ListAllScreens(page int, per_page int) (*models.PreviewScreens, error) {
    tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	previews := &models.PreviewScreens{}
    if err := q.All(previews); err != nil {
        return nil, err
    }
    return previews, nil
}

func (cm ContentManagerDB) ListScreensContext(mcID uuid.UUID) (*models.PreviewScreens, error) {
    // Could add the context here correctly
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListScreens(mcID, page, limit)
}

// TODO: Re-Assign the preview based on screen information
func (cm ContentManagerDB) ListScreens(mcID uuid.UUID, page int, per_page int) (*models.PreviewScreens, error) {
	tx := cm.GetConnection()
	previews := &models.PreviewScreens{}
	q := tx.Paginate(page, per_page)
	q_conn := q.Where("media_container_id = ?", mcID)
	if q_err := q_conn.All(previews); q_err != nil {
		return nil, q_err
	}
	return previews, nil
}

// Need to make it use the manager and just show the file itself
func (cm ContentManagerDB) GetScreen(psID uuid.UUID) (*models.PreviewScreen, error) {
    previewScreen := &models.PreviewScreen{}
    tx := cm.GetConnection()
    err := tx.Find(previewScreen, psID)
    if err != nil {
        return nil, err
    }
    return previewScreen, nil

}

func (cm ContentManagerDB) CreateScreen(screen *models.PreviewScreen) error {
    tx := cm.GetConnection()
    return tx.Create(screen)
}

func (cm ContentManagerDB) CreateMedia(mc *models.MediaContainer) error {
    tx := cm.GetConnection()
    return tx.Create(mc)
}


// TODO: Security vuln need to ensure that you can only create UNDER the directory
// specified by the initial load.
func (cm ContentManagerDB) CreateContainer(c *models.Container) error {
    tx := cm.GetConnection()
    return tx.Create(c)
}

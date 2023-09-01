/**
* Implements the ContentManager interface and stores information in a postgres db.
 */
package managers

import (
	"contented/models"
	"contented/utils"
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
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

func (cm ContentManagerDB) ListContentContext(cID uuid.UUID) (*models.Contents, error) {
	// Could add the context here correctly
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListContent(cID, page, limit)
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerDB) ListContent(cID uuid.UUID, page int, per_page int) (*models.Contents, error) {
	log.Printf("Get a list of content from DB, we should have some %s", cID.String())
	tx := cm.GetConnection()
	contentContainers := &models.Contents{}

	// Paginate results. Params "page" and "per_page" control pagination.
	q := tx.Paginate(page, per_page)
	q_conn := q.Where("container_id = ?", cID)
	if q_err := q_conn.All(contentContainers); q_err != nil {
		return nil, q_err
	}
	return contentContainers, nil
}

func (cm ContentManagerDB) GetContent(mcID uuid.UUID) (*models.Content, error) {
	log.Printf("Get a single content %s", mcID)
	tx := cm.GetConnection()
	mc := &models.Content{}
	if err := tx.Eager().Find(mc, mcID); err != nil {
		return nil, err
	}
	//tx.Load(&mc, "Screens")
	// wat := tx.Load(&mc.Tags, "Tags")
	return mc, nil
}

func (cm ContentManagerDB) UpdateContainer(c *models.Container) error {
	tx := cm.GetConnection()
	return tx.Update(c)
}

func (cm ContentManagerDB) UpdateContent(content *models.Content) error {
	tx := cm.GetConnection()
	return tx.Eager().Update(content)
}

func (cm ContentManagerDB) UpdateScreen(s *models.Screen) error {
	tx := cm.GetConnection()
	return tx.Update(s)
}

func (cm ContentManagerDB) ListAllContent(page int, per_page int) (*models.Contents, error) {
	log.Printf("List all content DB manager")
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	contentContainers := &models.Contents{}
	if err := q.All(contentContainers); err != nil {
		return nil, err
	}
	return contentContainers, nil
}

// It should probably be able to search the container too?
func (cm ContentManagerDB) SearchContentContext() (*models.Contents, int, error) {
	params := cm.Params()
	_, per_page, page := GetPagination(params, cm.cfg.Limit)
	searchStr := StringDefault(params.Get("text"), "")
	cId := StringDefault(params.Get("cId"), "")
	contentType := StringDefault(params.Get("contentType"), "")
	return cm.SearchContent(searchStr, page, per_page, cId, contentType)
}

func (cm ContentManagerDB) SearchContent(search string, page int, per_page int, cId string, contentType string) (*models.Contents, int, error) {
	contentContainers := &models.Contents{}
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
	count, _ := q.Count(&models.Contents{})
	log.Printf("Total count of search content %d using search (%s) and contentType (%s)", count, search, contentType)

	if q_err := q.All(contentContainers); q_err != nil {
		return contentContainers, count, q_err
	}
	// Now need to get any screens and associate them in a lookup
	screenMap, s_err := cm.LoadRelatedScreens(contentContainers)
	contentWithScreens := models.Contents{}
	if s_err == nil {
		for _, mcPt := range *contentContainers {
			mc := mcPt // GoLang... sometimes this just makes me sad.
			if _, ok := screenMap[mc.ID]; ok {
				mc.Screens = screenMap[mc.ID]
			}
			contentWithScreens = append(contentWithScreens, mc)
		}
	}
	return &contentWithScreens, count, nil
}

func (cm ContentManagerDB) SearchContainers(search string, page int, per_page int) (*models.Containers, error) {
	if search == "" || search == "*" {
		return cm.ListContainers(page, per_page)
	}
	containers := &models.Containers{}
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	q = q.Where("name ilike ?", search)
	if q_err := q.All(containers); q_err != nil {
		return containers, q_err
	}
	return containers, nil
}

func (cm ContentManagerDB) LoadRelatedScreens(content *models.Contents) (models.ScreenCollection, error) {
	if content == nil || len(*content) == 0 {
		return nil, nil
	}
	videoIds := []string{}
	for _, mc := range *content {
		if strings.Contains(mc.ContentType, "video") {
			videoIds = append(videoIds, mc.ID.String())
		}
	}
	if len(videoIds) == 0 {
		log.Printf("None of these content were a video, skipping")
		return nil, nil
	}
	q := cm.GetConnection().Q().Where(`content_id = any($1)`, pq.Array(videoIds))
	screens := &models.Screens{}
	if q_err := q.All(screens); q_err != nil {
		log.Printf("Error loading video screens %s", q_err)
		return nil, q_err
	}

	screenMap := models.ScreenCollection{}
	for _, screen := range *screens {
		log.Printf("Found screen for %s", screen.ContentID.String())
		if _, ok := screenMap[screen.ContentID]; ok {
			screenMap[screen.ContentID] = append(screenMap[screen.ContentID], screen)
			log.Printf("Screen count %s %s", screen.ContentID, screenMap[screen.ContentID])
		} else {
			screenMap[screen.ContentID] = models.Screens{screen}
		}
	}
	return screenMap, nil
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

func (cm ContentManagerDB) FindFileRef(mcID uuid.UUID) (*models.Content, error) {
	mc_db := models.Content{}
	err := models.DB.Find(&mc_db, mcID)
	if err == nil {
		return &mc_db, nil
	}
	return nil, err
}

func (cm ContentManagerDB) GetPreviewForMC(mc *models.Content) (string, error) {
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

func (cm ContentManagerDB) FindActualFile(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		return "DB Manager View no Parent Found", err
	}
	log.Printf("DB Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
	return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

func (cm ContentManagerDB) ListAllScreensContext() (*models.Screens, error) {
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListAllScreens(page, limit)
}

func (cm ContentManagerDB) ListAllScreens(page int, per_page int) (*models.Screens, error) {
	previews := &models.Screens{}
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	if err := q.All(previews); err != nil {
		return nil, err
	}
	return previews, nil
}

func (cm ContentManagerDB) ListScreensContext(mcID uuid.UUID) (*models.Screens, error) {
	// Could add the context here correctly
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListScreens(mcID, page, limit)
}

// TODO: Re-Assign the preview based on screen information
func (cm ContentManagerDB) ListScreens(mcID uuid.UUID, page int, per_page int) (*models.Screens, error) {
	tx := cm.GetConnection()
	previews := &models.Screens{}
	q := tx.Paginate(page, per_page)
	q_conn := q.Where("content_id = ?", mcID)
	if q_err := q_conn.All(previews); q_err != nil {
		return nil, q_err
	}
	return previews, nil
}

// Need to make it use the manager and just show the file itself
func (cm ContentManagerDB) GetScreen(psID uuid.UUID) (*models.Screen, error) {
	previewScreen := &models.Screen{}
	tx := cm.GetConnection()
	err := tx.Find(previewScreen, psID)
	if err != nil {
		return nil, err
	}
	return previewScreen, nil

}

func (cm ContentManagerDB) CreateScreen(screen *models.Screen) error {
	tx := cm.GetConnection()
	return tx.Create(screen)
}

func (cm ContentManagerDB) CreateContent(mc *models.Content) error {
	tx := cm.GetConnection()
	err := tx.Create(mc)
	//log.Printf("What is the content %s", mc)
	return err
}

func (cm ContentManagerDB) ListAllTags(page int, perPage int) (*models.Tags, error) {
	tx := cm.GetConnection()
	tags := &models.Tags{}
	q := tx.Paginate(page, perPage)
	if q_err := q.All(tags); q_err != nil {
		return nil, q_err
	}
	return tags, nil
}

func (cm ContentManagerDB) ListAllTagsContext() (*models.Tags, error) {
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListAllTags(page, limit)
}

func (cm ContentManagerDB) CreateTag(tag *models.Tag) error {
	tx := cm.GetConnection()
	return tx.Create(tag)
}

func (cm ContentManagerDB) UpdateTag(tag *models.Tag) error {
	tx := cm.GetConnection()
	return tx.Update(tag)
}

func (cm ContentManagerDB) DeleteTag(tag *models.Tag) error {
	tx := cm.GetConnection()
	return tx.Destroy(tag)
}

func (cm ContentManagerDB) GetTag(tagID string) (*models.Tag, error) {
	log.Printf("DB Get a tag %s", tagID)
	tx := cm.GetConnection()
	t := &models.Tag{}
	if err := tx.Find(t, tagID); err != nil {
		return nil, err
	}
	return t, nil
}

func (cm ContentManagerDB) AssociateTag(t *models.Tag, mc *models.Content) error {
	log.Printf("Found %s with %s what the %s", mc.ID.String(), t.ID, mc.Tags)

	tx := cm.GetConnection()
	tags := append(mc.Tags, *t)
	err := tx.RawQuery("delete from contents_tags where content_id = ?", mc.ID).Exec()
	if err != nil {
		log.Printf("Could not associate tag %s", err)
		return err
	}
	// hate
	for _, t := range(tags) {
		link_err := tx.RawQuery("insert into contents_tags (tag_id, content_id) values (?, ?)", t.ID, mc.ID).Exec()
		if link_err != nil {
			log.Printf("Failed to re-link %s", link_err)
			return link_err
		}
	}
	return nil
}

func (cm ContentManagerDB) AssociateTagByID(tagId string, mcID uuid.UUID) error {
	mc, m_err := cm.GetContent(mcID)
	t, t_err := cm.GetTag(tagId)
	if m_err != nil || t_err != nil {
		return errors.New("DB Tag or content container not found")
	}
	return cm.AssociateTag(t, mc)
}

// TODO: Security vuln need to ensure that you can only create UNDER the directory
// specified by the initial load.
func (cm ContentManagerDB) CreateContainer(c *models.Container) error {
	tx := cm.GetConnection()
	return tx.Create(c)
}

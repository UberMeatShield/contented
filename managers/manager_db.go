/**
* Implements the ContentManager interface and stores information in a postgres db.
 */
package managers

import (
	"contented/models"
	"contented/utils"
	"errors"
	"fmt"
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
	cfg := utils.GetCfg()
	if cfg.ReadOnly == true {
		return false
	}
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

// Note this DOES allow for loading hidden content
func (cm ContentManagerDB) GetContent(mcID uuid.UUID) (*models.Content, error) {
	log.Printf("Get a single content object %s", mcID)
	tx := cm.GetConnection()
	mc := &models.Content{}
	if err := tx.Eager().Find(mc, mcID); err != nil {
		return nil, err
	}
	return mc, nil
}

// Update of the container should check utils.SubPath
func (cm ContentManagerDB) UpdateContainer(c *models.Container) (*models.Container, error) {
	tx := cm.GetConnection()
	err := tx.Update(c)
	if err != nil {
		return c, err
	}
	return cm.GetContainer(c.ID)
}

func (cm ContentManagerDB) UpdateContent(content *models.Content) error {
	// Check if file exists or allow content to be 'empty'?
	tx := cm.GetConnection()
	if content.NoFile == false {
		cnt, cErr := cm.GetContainer(content.ContainerID.UUID)
		if cErr != nil {
			msg := fmt.Sprintf("Parent container %s not found", content.ContainerID.UUID.String())
			return errors.New(msg)
		}

		exists, pErr := utils.HasContent(content.Src, cnt.GetFqPath())
		if exists == false || pErr != nil {
			log.Printf("Content not in container %s", pErr)
			return errors.New(fmt.Sprintf("Invalid content src %s for container %s", content.Src, cnt.Name))
		}
	}
	err := tx.Eager().Update(content)
	if err != nil {
		return err
	}
	// Just trust we will associate all valid tags to the content.
	if content.Tags != nil {
		return cm.AssociateTag(nil, content)
	}
	return nil
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
	return cm.SearchContent(searchStr, page, per_page, cId, contentType, false)
}

func (cm ContentManagerDB) SearchContent(search string, page int, per_page int, cId string, contentType string, includeHidden bool) (*models.Contents, int, error) {
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
	if includeHidden == false {
		q = q.Where(`hidden = ?`, false)
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

func (cm ContentManagerDB) SearchContainers(search string, page int, per_page int, includeHidden bool) (*models.Containers, error) {
	if search == "" || search == "*" {
		return cm.ListContainers(page, per_page)
	}
	containers := &models.Containers{}
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	q = q.Where("name ilike ?", search)
	if includeHidden == false {
		q = q.Where(`hidden = ?`, false)
	}
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

func (cm ContentManagerDB) ListContainers(page int, per_page int) (*models.Containers, error) {
	return cm.ListContainersFiltered(page, per_page, false)
}

// TODO: Add in support for actually doing the query using the current buffalo.Context
func (cm ContentManagerDB) ListContainersFiltered(page int, per_page int, includeHidden bool) (*models.Containers, error) {
	log.Printf("DB List all containers")
	tx := cm.GetConnection()
	q := tx.Paginate(page, per_page)
	if includeHidden == false {
		q = q.Where("hidden = ?", false)
	}

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

func (cm ContentManagerDB) CreateContent(content *models.Content) error {
	tx := cm.GetConnection()

	// Eager doesn't actually create tags but should we create any old tag on
	// submission?
	validTags, t_err := cm.GetValidTags(&content.Tags)
	if t_err != nil {
		log.Printf("Could not determine valid tags %s", t_err)
		return t_err
	}
	content.Tags = *validTags

	_, err := tx.Eager().ValidateAndCreate(content)
	if err != nil {
		return err
	}
	return err
}

// Note we very intentionally are NOT destroying items on disk.
func (cm ContentManagerDB) DestroyContent(id string) (*models.Content, error) {
	tx := cm.GetConnection()
	content := &models.Content{}
	if err := tx.Find(content, id); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not find content with id %s", id))
	}
	if err := tx.Destroy(content); err != nil {
		return content, err
	}
	return content, nil
}

func (cm ContentManagerDB) DestroyContainer(id string) (*models.Container, error) {
	tx := cm.GetConnection()
	cnt := &models.Container{}
	if err := tx.Find(cnt, id); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not find container with id %s", id))
	}
	if err := tx.Destroy(cnt); err != nil {
		return cnt, err
	}
	return cnt, nil
}

func (cm ContentManagerDB) DestroyScreen(id string) (*models.Screen, error) {
	tx := cm.GetConnection()
	screen := &models.Screen{}
	if err := tx.Find(screen, id); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not find screen with id %s", id))
	}
	if err := tx.Destroy(screen); err != nil {
		return screen, err
	}
	return screen, nil
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

func (cm ContentManagerDB) GetValidTags(tags *models.Tags) (*models.Tags, error) {
	tx := cm.GetConnection()
	validTags := models.Tags{}
	ids := []string{}
	for _, tag := range *tags {
		ids = append(ids, tag.ID)
	}
	if len(ids) == 0 {
		return &validTags, nil
	}

	q := tx.Q().Where("id in (?)", ids)
	q_err := q.All(&validTags)
	if q_err != nil {
		log.Printf("Error validating tags %s", q_err)
		return nil, q_err
	}
	return &validTags, nil
}

// You can also use this with a content element with tags to actually associate them.
func (cm ContentManagerDB) AssociateTag(t *models.Tag, mc *models.Content) error {
	// TODO: Could require [Tags] and not do the append with this function
	tx := cm.GetConnection()

	tags := models.Tags{}
	if mc.Tags != nil {
		tags = mc.Tags
	}
	if t != nil {
		tags = append(mc.Tags, *t)
	}
	// Filter these tags to only VALID tags already in the system since Eager is super
	// busted on many to many relations.
	validTags, v_err := cm.GetValidTags(&tags)
	if v_err != nil {
		return v_err
	}
	tags = *validTags

	err := tx.RawQuery("delete from contents_tags where content_id = ?", mc.ID).Exec()
	if err != nil {
		log.Printf("Could not associate tag %s", err)
		return err
	}

	// I really don't love this but Buffalo many_to_many associations do NOT handle updates.  In addition an integer
	// as the join table ID also doesn't seem to do the link even on a create.
	sql_str := "insert into contents_tags (id, tag_id, content_id, created_at, updated_at) values (?, ?, ?, current_timestamp, current_timestamp)"
	for _, t := range tags {
		linkID, _ := uuid.NewV4()
		link_err := tx.RawQuery(sql_str, linkID, t.ID, mc.ID).Exec()
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

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
func (cm ContentManagerDB) UpdateContainer(cnt *models.Container) (*models.Container, error) {
	cfg := cm.GetCfg()
	pathOk, pErr := utils.PathIsOk(cnt.Path, cnt.Name, cfg.Dir)
	if pErr != nil {
		log.Printf("Path does not exist on disk under the config directory err %s", pErr)
		return nil, pErr
	}
	if pathOk == false {
		msg := fmt.Sprintf("The path was not under a valid container %s", pErr)
		return nil, errors.New(msg)
	}
	tx := cm.GetConnection()
	err := tx.Update(cnt)
	if err != nil {
		return cnt, err
	}
	return cm.GetContainer(cnt.ID)
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
	sr := ContextToSearchQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContent(sr)
}

func (cm ContentManagerDB) SearchContent(sr SearchQuery) (*models.Contents, int, error) {
	contentContainers := &models.Contents{}
	tx := cm.GetConnection()
	q := tx.Paginate(sr.Page, sr.PerPage)

	if len(sr.Tags) > 0 {
		q = q.Join("contents_tags as ct", "ct.content_id = contents.id").Where("ct.tag_id IN (?)", sr.Tags)
	}
	// Could also search description
	if sr.Text != "*" && sr.Text != "" {
		search := ("%" + sr.Text + "%")
		q = q.Where(`src like ?`, search)
	}
	if sr.ContentType != "" {
		contentType := ("%" + sr.ContentType + "%")
		q = q.Where(`content_type ilike ?`, contentType)
	}
	if sr.ContainerID != "" {
		q = q.Where(`container_id = ?`, sr.ContainerID)
	}
	if sr.Hidden == false {
		q = q.Where(`hidden = ?`, false)
	}

	count, _ := q.Count(&models.Contents{})
	log.Printf("Total count of search content %d using search (%s) and contentType (%s)", count, sr.Text, sr.ContentType)

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

func (cm ContentManagerDB) ListScreensContext() (*models.Screens, int, error) {
	// Could add the context here correctly
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.cfg.Limit)
	sr := ScreensQuery{
		Page:      page,
		PerPage:   limit,
		ContentID: params.Get("content_id"),
	}
	return cm.ListScreens(sr)
}

// TODO: Re-Assign the preview based on screen information
func (cm ContentManagerDB) ListScreens(sr ScreensQuery) (*models.Screens, int, error) {
	tx := cm.GetConnection()
	previews := &models.Screens{}
	q := tx.Paginate(sr.Page, sr.PerPage)
	if sr.ContentID != "" {
		q = q.Where("content_id = ?", sr.ContentID)
	}
	if q_err := q.All(previews); q_err != nil {
		return nil, -1, q_err
	}
	count, _ := q.Count(&models.Screens{})
	return previews, count, nil
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

func (cm ContentManagerDB) DestroyTag(id string) (*models.Tag, error) {
	tag, err := cm.GetTag(id)
	if err != nil {
		return nil, err
	}
	tx := cm.GetConnection()
	return tag, tx.Destroy(tag)
}

func (cm ContentManagerDB) GetTag(tagID string) (*models.Tag, error) {
	// log.Printf("DB Get a tag %s", tagID)
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

// TODO: Should we be able to CREATE actual directory information under the
// parent container if it does not exist?
func (cm ContentManagerDB) CreateContainer(c *models.Container) error {
	// Prevent some containers unless they are under the Dir path.
	cfg := cm.GetCfg()

	// TODO: Config value for restrict to under root dir?
	ok, err := utils.PathIsOk(c.Path, c.Name, cfg.Dir)
	if err != nil {
		log.Printf("Path does not exist on disk under the config directory err %s", err)
		return err
	}
	tx := cm.GetConnection()
	if ok == true {
		return tx.Create(c)
	}
	msg := fmt.Sprintf("The directory was not under the config path %s", c.Name)
	return errors.New(msg)

}

func (cm ContentManagerDB) CreateTask(t *models.TaskRequest) (*models.TaskRequest, error) {
	if t == nil {
		return nil, errors.New("Requires a valid task")
	}
	tx := cm.GetConnection()
	t.Status = models.TaskStatus.NEW // The defaults do not seem to work right...
	err := tx.Create(t)
	if err != nil {
		return nil, err
	}
	return t, err
}

func (cm ContentManagerDB) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequest, error) {
	if t == nil {
		return t, errors.New("No task to update")
	}
	checkStatus, cErr := cm.GetTask(t.ID)
	if cErr != nil {
		return nil, cErr
	}
	// I would like to do this in the Update query but then I have to do a raw query for the full
	// update (there isn't a conditional add into the sql for some reason)
	if checkStatus.Status == currentState {
		tx := cm.GetConnection()
		upErr := tx.Update(t)
		if upErr != nil {
			return nil, upErr
		}
	} else {
		msg := fmt.Sprintf("The current DB status %s != exec status %s", checkStatus.Status, currentState)
		log.Printf(msg)
		return nil, errors.New(msg)
	}
	return cm.GetTask(t.ID)
}

func (cm ContentManagerDB) GetTask(id uuid.UUID) (*models.TaskRequest, error) {
	task := models.TaskRequest{}
	tx := cm.GetConnection()
	err := tx.Find(&task, id)
	return &task, err
}

// Get the next task for processing (not super thread safe but enough for mem manager)
func (cm ContentManagerDB) NextTask() (*models.TaskRequest, error) {
	tasks := models.TaskRequests{}
	tx := cm.GetConnection()
	q := tx.Paginate(1, 1)
	err := q.Where("status = ?", models.TaskStatus.NEW).All(&tasks)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 1 {
		task := tasks[0]
		task.Status = models.TaskStatus.PENDING
		return cm.UpdateTask(&task, models.TaskStatus.NEW)
	}
	return nil, errors.New("No Tasks to pull off the queue")
}

func (cm ContentManagerDB) ListTasksContext() (*models.TaskRequests, error) {
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.GetCfg().Limit)
	query := TaskQuery{
		Page:      page,
		PerPage:   limit,
		ContentID: StringDefault(params.Get("content_id"), ""),
		Status:    StringDefault(params.Get("status"), ""), // Check it is in the Status values?
	}
	return cm.ListTasks(query)
}

func (cm ContentManagerDB) ListTasks(query TaskQuery) (*models.TaskRequests, error) {
	tasks := models.TaskRequests{}
	tx := cm.GetConnection()
	q := tx.Paginate(query.Page, query.PerPage)
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.ContentID != "" {
		q = q.Where("content_id = ?", query.ContentID)
	}
	err := q.All(&tasks)
	if err != nil {
		return nil, err
	}
	return &tasks, nil
}

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
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// DB version of content management
type ContentManagerDB struct {
	cfg *utils.DirConfigEntry

	/* Is this even useful ? */
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
	return !cfg.ReadOnly
}

func (cm ContentManagerDB) ListContentContext() (*models.Contents, int64, error) {
	// Could add the context here correctly
	params := cm.Params()
	offset, limit, page := GetPagination(params, cm.cfg.Limit)
	cs := ContentQuery{
		ContainerID: StringDefault(params.Get("container_id"), ""),
		Page:        page,
		Offset:      offset,
		PerPage:     limit,
	}
	return cm.ListContent(cs)
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerDB) ListContent(cs ContentQuery) (*models.Contents, int64, error) {
	log.Printf("Get a list of content from DB, we should have some %s", cs.ContainerID)
	tx := cm.GetConnection()
	contents := &models.Contents{}

	// Paginate results. Params "page" and "per_page" control pagination.
	q := tx.Model(&models.Contents{})
	if cs.ContainerID != "" {
		q = q.Where("container_id = ?", cs.ContainerID)
	}
	q = q.Order(models.GetContentOrder(cs.Order, cs.Direction))

	// Oy, have to count
	var count int64
	countRes := q.Count(&count)
	if countRes.Error != nil {
		return nil, count, countRes.Error
	}

	// Throw error if the offset is too large?
	q = q.Offset(cs.Offset).Limit(GetPerPage(cs.PerPage))
	if count > 0 {
		if res := q.Find(contents); res.Error != nil {
			return nil, -1, res.Error
		}
	}
	return contents, count, nil
}

// Note this DOES allow for loading hidden content
func (cm ContentManagerDB) GetContent(mcID int64) (*models.Content, error) {
	log.Printf("get a single content object %d", mcID)
	tx := cm.GetConnection()
	mc := &models.Content{}
	if res := tx.Preload("Screens").Preload("Tags").Find(mc, mcID); res.Error != nil {
		return nil, res.Error
	}
	if mc.ID == 0 {
		return nil, fmt.Errorf("content not found %d", mcID)
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
	if !pathOk {
		msg := fmt.Sprintf("The path was not under a valid container %s", pErr)
		return nil, errors.New(msg)
	}
	tx := cm.GetConnection()
	if res := tx.Save(cnt); res.Error != nil {
		return cnt, res.Error
	}
	return cm.GetContainer(cnt.ID)
}

func (cm ContentManagerDB) UpdateContent(content *models.Content) error {
	// Check if file exists or allow content to be 'empty'?
	tx := cm.GetConnection()
	if !content.NoFile {
		if content.ContainerID != nil {
			cnt, cErr := cm.GetContainer(*content.ContainerID)
			if cErr != nil {
				return fmt.Errorf("parent container %d not found", content.ContainerID)
			}

			exists, pErr := utils.HasContent(content.Src, cnt.GetFqPath())
			if !exists || pErr != nil {
				log.Printf("Content not in container %s", pErr)
				return fmt.Errorf("invalid content src %s for container %s", content.Src, cnt.Name)
			}
		}
	}

	tags := content.Tags
	content.Tags = nil // We don't want to allow the save to create tags not in the system
	if res := tx.Save(content); res.Error != nil {
		return res.Error
	}
	// Just trust we will associate all valid tags to the content.
	if tags != nil {
		content.Tags = tags
		return cm.AssociateTag(nil, content)
	}
	return nil
}

/**
 * TODO: Make this a batch update using the Go DB layer (should work maybe?)
 */
func (cm ContentManagerDB) UpdateContents(contents models.Contents) error {
	if contents == nil {
		return nil
	}
	for _, content := range contents {
		err := cm.UpdateContent(&content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm ContentManagerDB) UpdateScreen(s *models.Screen) error {
	tx := cm.GetConnection()
	res := tx.Save(s)
	return res.Error
}

func (cm ContentManagerDB) ListAllContent(page int64, perPage int) (*models.Contents, error) {
	log.Printf("List all content DB manager")
	tx := cm.GetConnection()

	q := tx.Offset(int(page) - 1).Limit(perPage)
	contentContainers := &models.Contents{}
	if res := q.Find(contentContainers); res.Error != nil {
		return nil, res.Error
	}
	return contentContainers, nil
}

// It should probably be able to search the container too?
func (cm ContentManagerDB) SearchContentContext() (*models.Contents, int64, error) {
	sr := ContextToContentQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContent(sr)
}

// It should probably be able to search the container too?
func (cm ContentManagerDB) SearchContainersContext() (*models.Containers, int64, error) {
	cq := ContextToContainerQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContainers(cq)
}

func (cm ContentManagerDB) SearchContent(sr ContentQuery) (*models.Contents, int64, error) {
	contents := &models.Contents{}
	q := cm.GetConnection().Model(&contents)
	if len(sr.Tags) > 0 {
		// TODO: This is almost certainly the problem
		q = q.Joins("JOIN contents_tags as ct ON ct.content_id = contents.id").Where("ct.tag_id IN (?)", sr.Tags)
	}
	// TODO: Could also search description (expand this)
	if sr.Search != "*" && sr.Search != "" {
		search := ("%" + sr.Search + "%")
		q = q.Where(`src like ?`, search)
	}
	if sr.ContentID != "" {
		q = q.Where("id = ?", sr.ContentID)
	}
	if sr.Text != "" {
		q = q.Where(`src = ?`, sr.Text)
	}
	if sr.ContentType != "" {
		contentType := ("%" + sr.ContentType + "%")
		q = q.Where(`content_type ilike ?`, contentType)
	}
	if sr.ContainerID != "" {
		q = q.Where(`container_id = ?`, sr.ContainerID)
	}
	if sr.IncludeHidden {
		q = q.Where(`hidden = ?`, false)
	}
	q = q.Order(models.GetContentOrder(sr.Order, sr.Direction))

	var count int64
	if res := q.Model(models.Contents{}).Count(&count); res.Error != nil {
		return nil, count, res.Error
	}

	q = q.Offset(sr.Offset).Limit(GetPerPage(sr.PerPage))
	log.Printf("Total count of search content %d using search (%s) and contentType (%s)", count, sr.Text, sr.ContentType)
	if count > 0 {
		if cRes := q.Find(contents); cRes.Error != nil {
			return contents, -1, cRes.Error
		}

		// Now need to get any screens and associate them in a lookup
		screenMap, s_err := cm.LoadRelatedScreens(contents)
		contentWithScreens := models.Contents{}
		if s_err == nil {
			for _, mcPt := range *contents {
				mc := mcPt // GoLang... sometimes this just makes me sad.
				if _, ok := screenMap[mc.ID]; ok {
					mc.Screens = screenMap[mc.ID]
				}
				contentWithScreens = append(contentWithScreens, mc)
			}
		}
		return &contentWithScreens, count, nil
	}
	return contents, count, nil
}

func (cm ContentManagerDB) SearchContainers(cs ContainerQuery) (*models.Containers, int64, error) {
	if cs.Search == "" || cs.Search == "*" {
		return cm.ListContainers(cs)
	}
	containers := &models.Containers{}
	q := cm.GetConnection().Where("name ilike ?", cs.Search)
	if !cs.IncludeHidden {
		q = q.Where(`hidden = ?`, false)
	}
	q = q.Order(models.GetContainerOrder(cs.Order, cs.Direction))

	var count int64
	res := q.Model(&models.Containers{}).Count(&count)
	if res.Error != nil {
		return nil, count, res.Error
	}

	q = q.Offset(cs.Offset).Limit(GetPerPage(cs.PerPage))
	if count > 0 {
		if res := q.Find(containers); res.Error != nil {
			return containers, -1, res.Error
		}
	}
	return containers, count, nil
}

func (cm ContentManagerDB) LoadRelatedScreens(content *models.Contents) (models.ScreenCollection, error) {
	if content == nil || len(*content) == 0 {
		return nil, nil
	}
	videoIds := []string{}
	for _, mc := range *content {
		if strings.Contains(mc.ContentType, "video") {
			videoIds = append(videoIds, strconv.FormatInt(mc.ID, 10))
		}
	}
	if len(videoIds) == 0 {
		log.Printf("None of these content were a video, skipping")
		return nil, nil
	}
	tx := cm.GetConnection()

	screens := &models.Screens{}
	q := tx.Where(`content_id = any($1)`, pq.Array(videoIds)).Find(screens)
	if q.Error != nil {
		log.Printf("Error loading video screens %s", q.Error)
		return nil, q.Error
	}

	screenMap := models.ScreenCollection{}
	for _, screen := range *screens {
		log.Printf("found screen for %d", screen.ContentID)
		if _, ok := screenMap[screen.ContentID]; ok {
			screenMap[screen.ContentID] = append(screenMap[screen.ContentID], screen)
			log.Printf("screen count %d %s", screen.ContentID, screenMap[screen.ContentID])
		} else {
			screenMap[screen.ContentID] = models.Screens{screen}
		}
	}
	return screenMap, nil
}

// The default list using the current manager configuration
func (cm ContentManagerDB) ListContainersContext() (*models.Containers, int64, error) {
	params := cm.Params()
	offset, limit, page := GetPagination(params, GetPerPage(cm.cfg.Limit))
	cs := ContainerQuery{
		Name:    StringDefault(params.Get("name"), ""),
		Page:    page,
		Offset:  offset,
		PerPage: limit,
	}
	return cm.ListContainers(cs)
}

func (cm ContentManagerDB) ListContainers(cs ContainerQuery) (*models.Containers, int64, error) {
	return cm.ListContainersFiltered(cs)
}

// TODO: Add in support for actually doing the query using the current buffalo.Context
func (cm ContentManagerDB) ListContainersFiltered(cs ContainerQuery) (*models.Containers, int64, error) {
	tx := cm.GetConnection()
	q := tx.Offset(cs.Offset).Limit(GetPerPage(cs.PerPage))
	if !cs.IncludeHidden {
		q = q.Where("hidden = ?", false)
	}
	q.Order(models.GetContainerOrder(cs.Order, cs.Direction))

	// Retrieve all Containers from the DB (if there are any)
	var count int64
	cRes := q.Model(&models.Containers{}).Count(&count)
	if cRes.Error != nil {
		return nil, count, cRes.Error
	}
	containers := &models.Containers{}
	if count > 0 {
		if res := q.Find(containers); res.Error != nil {
			return nil, count, res.Error
		}
	}
	return containers, count, nil
}

// TODO: Need a preview test using the database where we do NOT have a preview created
func (cm ContentManagerDB) GetContainer(cID int64) (*models.Container, error) {
	log.Printf("get a single container %d", cID)
	tx := cm.GetConnection()

	// Allocate an empty Container p := cm.Params()
	container := &models.Container{}
	if res := tx.Find(container, cID); res.Error != nil {
		return nil, res.Error
	}
	if container.ID == 0 {
		return nil, fmt.Errorf("container not found %d", cID)
	}
	return container, nil
}

func (cm *ContentManagerDB) Initialize() {
	// Connect to the DB using the context or some other option?
	log.Printf("Make a DB connection here")
}

func (cm ContentManagerDB) GetPreviewForMC(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(*mc.ContainerID)
	if err != nil {
		return "DB Manager Preview no Parent Found", err
	}
	src := mc.Src
	if mc.Preview != "" {
		src = mc.Preview
	}
	log.Printf("DB Manager loading %d preview %s\n", mc.ID, src)
	return utils.GetFilePathInContainer(src, cnt.GetFqPath())
}

func (cm ContentManagerDB) FindActualFile(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(*mc.ContainerID)
	if err != nil {
		return "DB Manager View no Parent Found", err
	}
	log.Printf("DB Manager View %d loading up %s\n", mc.ID, mc.Src)
	return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

func (cm ContentManagerDB) ListScreensContext() (*models.Screens, int64, error) {
	// Could add the context here correctly
	params := cm.Params()
	offset, limit, page := GetPagination(params, cm.cfg.Limit)
	sr := ScreensQuery{
		Page:      page,
		Offset:    offset,
		PerPage:   limit,
		ContentID: params.Get("content_id"),
	}
	return cm.ListScreens(sr)
}

// TODO: Re-Assign the preview based on screen information
func (cm ContentManagerDB) ListScreens(sr ScreensQuery) (*models.Screens, int64, error) {
	tx := cm.GetConnection()
	previews := &models.Screens{}

	q := tx.Model(previews)
	if sr.ContentID != "" {
		q = q.Where("content_id = ?", sr.ContentID)
	}
	var count int64
	if cRes := q.Count(&count); cRes.Error != nil {
		return nil, count, cRes.Error
	}
	if count > int64(0) {
		q = q.Offset(sr.Offset).Limit(GetPerPage(sr.PerPage))
		if res := q.Find(previews); res.Error != nil {
			return nil, -1, res.Error
		}
	}
	return previews, count, nil
}

// Need to make it use the manager and just show the file itself
func (cm ContentManagerDB) GetScreen(psID int64) (*models.Screen, error) {
	previewScreen := &models.Screen{}
	tx := cm.GetConnection()
	res := tx.Find(previewScreen, psID)
	if res.Error != nil {
		return nil, res.Error
	}
	if previewScreen.ID == 0 {
		return nil, fmt.Errorf("screen not found %d", psID)
	}
	return previewScreen, nil

}

func (cm ContentManagerDB) CreateScreen(screen *models.Screen) error {
	if screen == nil || screen.ContentID == 0 {
		return fmt.Errorf("invalid screenshot data %s", screen)
	}
	tx := cm.GetConnection().Create(screen)
	return tx.Error
}

func (cm ContentManagerDB) CreateContent(content *models.Content) error {
	tx := cm.GetConnection()

	if content.Tags == nil {
		content.Tags = models.Tags{}
	}
	validTags, t_err := cm.GetValidTags(&content.Tags)
	if t_err != nil {
		log.Printf("Could not determine valid tags %s", t_err)
		return t_err
	}
	content.Tags = *validTags

	if res := tx.Save(content); res.Error != nil {
		return res.Error
	}
	return nil
}

// Note we very intentionally are NOT destroying items on disk.
func (cm ContentManagerDB) DestroyContent(id int64) (*models.Content, error) {
	tx := cm.GetConnection()
	content := &models.Content{}
	if res := tx.Find(content, id); res.Error != nil {
		return nil, fmt.Errorf("could not find content with id %d", id)
	}

	if res := tx.Delete(content); res.Error != nil {
		return content, res.Error
	}
	return content, nil
}

func (cm ContentManagerDB) DestroyContainer(id string) (*models.Container, error) {
	tx := cm.GetConnection()
	cnt := &models.Container{}
	log.Printf("What the fuck Lookup %s", id)
	if res := tx.Find(cnt, id); res.Error != nil {
		return nil, fmt.Errorf("could not find container with id %s", id)
	}
	if res := tx.Delete(cnt); res.Error != nil {
		return cnt, res.Error
	}
	log.Printf("What the fuck %s", cnt)
	return cnt, nil
}

func (cm ContentManagerDB) DestroyScreen(id string) (*models.Screen, error) {
	tx := cm.GetConnection()
	screen := &models.Screen{}
	if res := tx.Find(screen, id); res.Error != nil {
		return nil, fmt.Errorf("could not find screen with id %s", id)
	}
	if res := tx.Delete(screen); res.Error != nil {
		return screen, res.Error
	}
	return screen, nil
}

func (cm ContentManagerDB) ListAllTags(tq TagQuery) (*models.Tags, int64, error) {
	q := cm.GetConnection()
	if tq.TagType != "" {
		q = q.Where("tag_type = ?", tq.TagType)
	}

	var total int64
	cRes := q.Model(&models.Tags{}).Count(&total)
	if cRes.Error != nil {
		return nil, total, cRes.Error
	}

	q = q.Offset(tq.Offset).Limit(GetPerPage(tq.PerPage))
	tags := &models.Tags{}
	if total > int64(0) {
		if res := q.Find(tags); res.Error != nil {
			return nil, total, res.Error
		}
	}
	return tags, total, nil
}

func (cm ContentManagerDB) ListAllTagsContext() (*models.Tags, int64, error) {
	params := cm.Params()
	offset, limit, page := GetPagination(params, cm.cfg.Limit)
	tq := TagQuery{
		Page:    page,
		PerPage: limit,
		TagType: StringDefault(params.Get("tag_type"), ""),
		Offset:  offset,
	}
	return cm.ListAllTags(tq)
}

func (cm ContentManagerDB) CreateTag(tag *models.Tag) error {
	res := cm.GetConnection().Create(tag)
	return res.Error
}

func (cm ContentManagerDB) UpdateTag(tag *models.Tag) error {
	tx := cm.GetConnection().Save(tag)
	return tx.Error
}

func (cm ContentManagerDB) DestroyTag(id string) (*models.Tag, error) {
	tag, err := cm.GetTag(id)
	if err != nil {
		return nil, err
	}
	tx := cm.GetConnection()
	return tag, tx.Delete(tag).Error
}

func (cm ContentManagerDB) GetTag(tagID string) (*models.Tag, error) {
	// log.Printf("DB Get a tag %s", tagID)
	tx := cm.GetConnection()
	t := &models.Tag{}
	if res := tx.First(t, "id = ?", tagID); res.Error != nil {
		return nil, res.Error
	}
	if t.ID == "" {
		return nil, fmt.Errorf("no tag found with %s", tagID)
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

	q := tx.Where("id in (?)", ids).Find(&validTags)
	if q.Error != nil {
		log.Printf("Error validating tags %s", q.Error)
		return nil, q.Error
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

	res := tx.Exec("delete from contents_tags where content_id = ?", mc.ID)
	if res.Error != nil {
		log.Printf("Could not associate tag %s", res.Error)
		return res.Error
	}

	// I really don't love this but Buffalo many_to_many associations do NOT handle updates.  In addition an integer
	// as the join table ID also doesn't seem to do the link even on a create.
	sql_str := "insert into contents_tags (tag_id, content_id, created_at, updated_at) values (?, ?, current_timestamp, current_timestamp)"
	for _, t := range tags {
		res := tx.Exec(sql_str, t.ID, mc.ID)

		if res.Error != nil {
			log.Printf("Failed to re-link %s", res.Error)
			return res.Error
		}
	}
	return nil
}

func (cm ContentManagerDB) AssociateTagByID(tagId string, mcID int64) error {
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
	if ok {
		tx := cm.GetConnection()
		return tx.Create(c).Error
	}
	return fmt.Errorf("the directory was not under the config path %s", c.Name)

}

func (cm ContentManagerDB) CreateTask(t *models.TaskRequest) (*models.TaskRequest, error) {
	if t == nil {
		return nil, errors.New("cannot create without a valid task")
	}
	tx := cm.GetConnection()
	t.Status = models.TaskStatus.NEW // The defaults do not seem to work right...
	res := tx.Create(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (cm ContentManagerDB) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequest, error) {
	if t == nil {
		return t, errors.New("no task to update")
	}
	checkStatus, cErr := cm.GetTask(t.ID)
	if cErr != nil {
		return nil, cErr
	}
	// I would like to do this in the Update query but then I have to do a raw query for the full
	// update (there isn't a conditional add into the sql for some reason)
	if checkStatus.Status == currentState {
		tx := cm.GetConnection()
		res := tx.Save(t)
		if res.Error != nil {
			return nil, res.Error
		}
	} else {
		msg := fmt.Sprintf("The current DB status %s != exec status %s", checkStatus.Status, currentState)
		log.Print(msg)
		return nil, errors.New(msg)
	}
	return cm.GetTask(t.ID)
}

func (cm ContentManagerDB) GetTask(id int64) (*models.TaskRequest, error) {
	task := models.TaskRequest{}
	tx := cm.GetConnection()
	res := tx.Find(&task, id)
	if task.ID == 0 {
		return nil, fmt.Errorf("no task found %d", id)
	}
	return &task, res.Error
}

// Get the next task for processing (not super thread safe but enough for mem manager)
func (cm ContentManagerDB) NextTask() (*models.TaskRequest, error) {
	tasks := models.TaskRequests{}
	tx := cm.GetConnection()
	res := tx.Limit(1).Where("status = ?", models.TaskStatus.NEW).Find(&tasks)
	if res.Error != nil {
		return nil, res.Error
	}
	if len(tasks) == 1 {
		task := tasks[0]
		task.Status = models.TaskStatus.PENDING
		return cm.UpdateTask(&task, models.TaskStatus.NEW)
	}
	return nil, errors.New("no tasks to pull off the queue")
}

func (cm ContentManagerDB) ListTasksContext() (*models.TaskRequests, int64, error) {
	params := cm.Params()
	offset, limit, page := GetPagination(params, cm.GetCfg().Limit)
	query := TaskQuery{
		Page:        page,
		Offset:      offset,
		PerPage:     limit,
		ContentID:   StringDefault(params.Get("content_id"), ""),
		ContainerID: StringDefault(params.Get("container_id"), ""),
		Status:      StringDefault(params.Get("status"), ""), // Check it is in the Status values?
		Search:      StringDefault(params.Get("search"), ""),
	}
	return cm.ListTasks(query)
}

func (cm ContentManagerDB) ListTasks(query TaskQuery) (*models.TaskRequests, int64, error) {
	tasks := models.TaskRequests{}
	tx := cm.GetConnection()
	q := tx.Offset(query.Offset).Limit(GetPerPage(query.PerPage))
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.ContentID != "" {
		q = q.Where("content_id = ?", query.ContentID)
	}
	if query.ContainerID != "" {
		q = q.Where("container_id = ?", query.ContainerID)
	}
	// TODO: Add in another search for searching errors potentially
	if query.Search != "" {
		search := ("%" + query.Search + "%")
		q = q.Where("message ilike ?", search)
	}

	var total int64
	cRes := q.Model(&models.TaskRequests{}).Count(&total)
	if cRes.Error != nil {
		return nil, total, cRes.Error
	}
	if total > 0 {
		if res := q.Find(&tasks); res.Error != nil {
			return nil, total, res.Error
		}
	}
	return &tasks, total, nil
}

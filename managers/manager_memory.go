/**
 * Implements the ContentManager interface but stores all information in memory.  This
 * is using the utility/MemStorage singleton which will load up the disk information only
 * one time.
 */
package managers

import (
	"contented/models"
	"contented/utils"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"

	"github.com/gofrs/uuid"
)

// Provides the support for looking up content by ID while only using memory
type ContentManagerMemory struct {
	cfg *utils.DirConfigEntry

	// Hmmm, this should use the memory manager probably
	/*
		ValidContent    models.ContentMap
		ValidContainers models.ContainerMap
		ValidScreens    models.ScreenMap
		ValidTags       models.TagsMap
		ValidTasks      models.TaskRequests
	*/
	validate string

	params *url.Values
	Params GetParamsType
}

// We do not allow editing in a memory manager
func (cm ContentManagerMemory) CanEdit() bool {
	return !cm.GetCfg().ReadOnly
}

// Provide the ability to set the configuration for a memory manager.
func (cm *ContentManagerMemory) SetCfg(cfg *utils.DirConfigEntry) {
	cm.cfg = cfg
	log.Printf("Memory Manager SetCfg() validate: %s\n", cm.validate)
}

// Get the currently configuration for this manager.
func (cm ContentManagerMemory) GetCfg() *utils.DirConfigEntry {
	// log.Printf("Memory Config is using path %s", cm.cfg.Dir)
	return cm.cfg
}

// On a first time load / use we will pull back content information from dist and from
// then on continue to use already loaded information.
func (cm *ContentManagerMemory) Initialize() {
	// TODO: Should we allow for a timeout or rescan option?
	memStorage := cm.GetStore()
	if memStorage.Initialized == false {
		memStorage = utils.InitializeMemory(cm.cfg.Dir)
	}
	// Remove this extra reference
	/*
		cm.ValidContainers = memStorage.ValidContainers
		cm.ValidContent = memStorage.ValidContent
		cm.ValidScreens = memStorage.ValidScreens
		cm.ValidTags = memStorage.ValidTags
		cm.ValidTasks = memStorage.ValidTasks
	*/
}

func (cm ContentManagerMemory) GetStore() *utils.MemoryStorage {
	return utils.GetMemStorage()
}

// Kinda strange but it seems hard to assign the type into an interface
// type GetParamsType func() *url.Values
func (cm ContentManagerMemory) GetParams() *url.Values {
	return cm.Params()
}

func (cm ContentManagerMemory) ListContentContext(cID uuid.UUID) (*models.Contents, error) {
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListContent(cID, page, limit)
}

// Listing all content ignoring the containerID still should respect hidden content.
func (cm ContentManagerMemory) ListAllContent(page int, per_page int) (*models.Contents, error) {
	return cm.ListAllContentFiltered(page, per_page, false)
}

func (cm ContentManagerMemory) ListAllContentFiltered(page int, per_page int, includeHidden bool) (*models.Contents, error) {
	m_arr := models.Contents{}
	mem := cm.GetStore()
	for _, m := range mem.ValidContent {
		if includeHidden == false {
			if m.Hidden == false {
				m_arr = append(m_arr, m)
			}
		} else {
			m_arr = append(m_arr, m)
		}
	}
	if len(m_arr) == 0 {
		return &m_arr, nil
	}

	// Did I create this just to sort by Idx across all content?  Kinda strange but required.
	sort.SliceStable(m_arr, func(i, j int) bool {
		return m_arr[i].Idx < m_arr[j].Idx
	})
	offset, end := GetOffsetEnd(page, per_page, len(m_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		m_arr = m_arr[offset:end]
		return &m_arr, nil
	}
	return &m_arr, nil
}

// It should probably be able to search the container too?
func (cm ContentManagerMemory) SearchContentContext() (*models.Contents, int, error) {
	sr := ContextToSearchQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContent(sr)
}

// Memory version is going to be extra annoying to tag search more than one tag on an or, or AND...
func (cm ContentManagerMemory) SearchContent(sr SearchQuery) (*models.Contents, int, error) {
	filteredContent, cErr := cm.getContentFiltered(sr.ContainerID, sr.Text, sr.ContentType, sr.Hidden)
	if cErr != nil {
		return nil, 0, cErr
	}
	if filteredContent == nil {
		empty := models.Contents{}
		return &empty, 0, nil
	}

	if len(sr.Tags) > 0 {
		filteredContent = cm.tagSearch(filteredContent, sr.Tags)
	}

	mc_arr := *filteredContent
	count := len(mc_arr)
	offset, end := GetOffsetEnd(sr.Page, sr.PerPage, count)
	if end > 0 { // If it is empty a slice ending in 0 = boom
		mc_arr = mc_arr[offset:end]
		return &mc_arr, count, nil
	}
	return &mc_arr, count, nil
}

// This is not great but there isn't a lookup of tag => contents
func (cm ContentManagerMemory) tagSearch(contents *models.Contents, tags []string) *models.Contents {
	filteredContents := models.Contents{}

	// Hmmm, unsafe in some ways because the data may not be loaded so know that this works for memory
	// manager because the tags are associated by the API / testing.
	for _, content := range *contents {
		for _, tag := range tags {
			if content.HasTag(tag) {
				filteredContents = append(filteredContents, content)
			}
		}
	}
	return &filteredContents
}

// Search Request may still make more sense.
func (cm ContentManagerMemory) getContentFiltered(containerID string, search string, contentType string, includeHidden bool) (*models.Contents, error) {
	// If a containerID is specified and is totally invalid raise an error, otherwise filter
	var mcArr models.Contents
	cidArr := models.Contents{}
	mem := cm.GetStore()

	if containerID != "" {
		cID, cErr := uuid.FromString(containerID)
		if cErr == nil {
			for _, mc := range mem.ValidContent {
				if mc.ContainerID.Valid && mc.ContainerID.UUID == cID {
					cidArr = append(cidArr, mc)
				}
			}
			mcArr = cidArr
		} else {
			return nil, cErr
		}
	} else {
		// Empty string for containerID is considered match all content
		for _, mc := range mem.ValidContent {
			cidArr = append(cidArr, mc)
		}
		mcArr = cidArr
	}

	if search != "" && search != "*" {
		searchStr := regexp.QuoteMeta(search)
		searcher := regexp.MustCompile("(?i)" + searchStr)
		searchArr := models.Contents{}
		for _, mc := range mcArr {
			if searcher.MatchString(mc.Src) {
				searchArr = append(searchArr, mc)
			}
		}
		mcArr = searchArr
	}

	if contentType != "" && contentType != "*" {
		searcher := regexp.MustCompile(contentType)
		contentArr := models.Contents{}
		for _, mc := range mcArr {
			if searcher.MatchString(mc.ContentType) {
				contentArr = append(contentArr, mc)
			}
		}
		mcArr = contentArr
	}

	if includeHidden == false {
		visibleArr := models.Contents{}
		for _, mc := range mcArr {
			if mc.Hidden != true {
				visibleArr = append(visibleArr, mc)
			}
		}
		mcArr = visibleArr
	}

	// Finally sort any content that is matching so that pagination will work
	sort.SliceStable(mcArr, func(i, j int) bool {
		return mcArr[i].Idx < mcArr[j].Idx
	})
	return &mcArr, nil
}

// TODO: Make it page but right now this will only be used in splash (regex it?)
func (cm ContentManagerMemory) SearchContainers(search string, page int, per_page int, includeHidden bool) (*models.Containers, error) {
	cArr := models.Containers{}
	if search == "" || search == "*" {
		return cm.ListContainersFiltered(page, per_page, includeHidden)
	}

	searcher := regexp.MustCompile("(?i)" + search)
	mem := cm.GetStore()
	for _, c := range mem.ValidContainers {
		if searcher.MatchString(c.Name) {
			if includeHidden == false {
				if c.Hidden != true {
					cArr = append(cArr, c)
				}
			} else {
				cArr = append(cArr, c)
			}
		}
	}
	return &cArr, nil
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListContent(ContainerID uuid.UUID, page int, per_page int) (*models.Contents, error) {
	return cm.ListContentFiltered(ContainerID, page, per_page, false)
}

func (cm ContentManagerMemory) ListContentFiltered(ContainerID uuid.UUID, page int, per_page int, includeHidden bool) (*models.Contents, error) {
	m_arr := models.Contents{}
	mem := cm.GetStore()
	for _, m := range mem.ValidContent {
		if m.ContainerID.Valid && m.ContainerID.UUID == ContainerID {
			if includeHidden == false {
				if m.Hidden == false {
					m_arr = append(m_arr, m)
				}
			} else {
				m_arr = append(m_arr, m)
			}
		}
	}
	sort.SliceStable(m_arr, func(i, j int) bool {
		return m_arr[i].Idx < m_arr[j].Idx
	})
	offset, end := GetOffsetEnd(page, per_page, len(m_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		m_arr = m_arr[offset:end]
		return &m_arr, nil
	}
	log.Printf("Get a list of content offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
	return &m_arr, nil

}

// Get a content element by the ID
func (cm ContentManagerMemory) GetContent(mcID uuid.UUID) (*models.Content, error) {
	// log.Printf("Memory Get a single content %s", mcID)
	mem := cm.GetStore()
	if mc, ok := mem.ValidContent[mcID]; ok {
		return &mc, nil
	}
	return nil, errors.New("Content was not found in memory")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateContainer(cnt *models.Container) (*models.Container, error) {
	// TODO: Validate that this updates the actual reference in mem storage
	cfg := cm.GetCfg()
	pathOk, err := utils.PathIsOk(cnt.Path, cnt.Name, cfg.Dir)
	if err != nil || pathOk == false {
		log.Printf("Path does not exist on disk under the config directory err %s", err)
		return nil, err
	}
	container, err := cm.GetStore().UpdateContainer(cnt)
	return container, err
}

// No updates should be allowed for memory management.
func (cm ContentManagerMemory) UpdateContent(content *models.Content) error {
	// TODO: Should I be able to ignore being in a container if there is no file?
	if content.NoFile == false {
		cnt, cErr := cm.GetContainer(content.ContainerID.UUID)
		if cErr != nil {
			msg := fmt.Sprintf("Parent container %s not found", content.ContainerID.UUID.String())
			log.Printf(msg)
			return errors.New(msg)
		}
		// Check if file exists or allow content to be 'empty'?
		exists, pErr := utils.HasContent(content.Src, cnt.GetFqPath())
		if exists == false || pErr != nil {
			log.Printf("Content not in container %s", pErr)
			return errors.New(fmt.Sprintf("Invalid content src %s for container %s", content.Src, cnt.Name))
		}
	}
	_, err := cm.GetStore().UpdateContent(content)
	return err
}

// Could use some extra validation (ensure there is content for the screen?)
func (cm ContentManagerMemory) UpdateScreen(s *models.Screen) error {
	_, err := cm.GetStore().UpdateScreen(s)
	return err
}

// Given the current parameters in the buffalo context return a list of matching containers.
func (cm ContentManagerMemory) ListContainersContext() (*models.Containers, error) {
	_, per_page, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListContainers(page, per_page)
}

// Actually list containers using a page and per_page which is consistent with buffalo standard pagination
func (cm ContentManagerMemory) ListContainers(page int, per_page int) (*models.Containers, error) {
	return cm.ListContainersFiltered(page, per_page, false)
}

func (cm ContentManagerMemory) ListContainersFiltered(page int, per_page int, includeHidden bool) (*models.Containers, error) {
	log.Printf("List Containers with page(%d) and per_page(%d)", page, per_page)

	c_arr := models.Containers{}
	mem := cm.GetStore()
	for _, c := range mem.ValidContainers {
		if includeHidden == false {
			if c.Hidden != true {
				c_arr = append(c_arr, c)
			}
		} else {
			c_arr = append(c_arr, c)
		}
	}
	sort.SliceStable(c_arr, func(i, j int) bool {
		return c_arr[i].Idx < c_arr[j].Idx
	})

	offset, end := GetOffsetEnd(page, per_page, len(c_arr))
	c_arr = c_arr[offset:end]
	return &c_arr, nil
}

// Get a single container given the primary key
func (cm ContentManagerMemory) GetContainer(cID uuid.UUID) (*models.Container, error) {
	// log.Printf("Get a single container %s", cID)
	mem := cm.GetStore()
	if c, ok := mem.ValidContainers[cID]; ok {
		return &c, nil
	}
	return nil, errors.New("Memory manager did not find this container id: " + cID.String())
}

func (cm ContentManagerMemory) GetPreviewForMC(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		return "Memory Manager Preview no Parent Found", err
	}
	src := mc.Src
	if mc.Preview != "" {
		src = mc.Preview
	}
	log.Printf("Memory Manager loading %s preview %s\n", mc.ID.String(), src)
	return utils.GetFilePathInContainer(src, cnt.GetFqPath())
}

func (cm ContentManagerMemory) FindActualFile(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		return "Memory Manager View no Parent Found", err
	}
	log.Printf("Memory Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
	return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

// If you want to do in memory testing and already manually created previews this will
// then try and use the previews for the in memory manager.
func (cm ContentManagerMemory) SetPreviewIfExists(mc *models.Content) (string, error) {
	c, err := cm.GetContainer(mc.ContainerID.UUID)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	pFile := utils.AssignPreviewIfExists(c, mc)
	return pFile, nil
}

func (cm ContentManagerMemory) ListScreensContext() (*models.Screens, int, error) {
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

// TODO: Get a pattern for each MC, look at a preview Destination, then match against the pattern
// And build out a set of screens.
func (cm ContentManagerMemory) ListScreens(sr ScreensQuery) (*models.Screens, int, error) {

	// Did I create this just to sort by Idx across all content?  Kinda strange
	mem := cm.GetStore()
	s_arr := models.Screens{}
	if sr.ContentID != "" {
		contentID, idErr := uuid.FromString(sr.ContentID)
		if idErr != nil {
			return nil, -1, idErr
		}
		for _, s := range mem.ValidScreens {
			if s.ContentID == contentID {
				s_arr = append(s_arr, s)
			}
		}
	} else {
		for _, s := range mem.ValidScreens {
			s_arr = append(s_arr, s)
		}
	}
	// Potentially text search the screens

	sort.SliceStable(s_arr, func(i, j int) bool {
		return s_arr[i].Idx < s_arr[j].Idx
	})
	count := len(s_arr)
	offset, end := GetOffsetEnd(sr.Page, sr.PerPage, len(s_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		s_arr = s_arr[offset:end]
		return &s_arr, count, nil
	}
	return &s_arr, count, nil
}

func (cm ContentManagerMemory) GetScreen(psID uuid.UUID) (*models.Screen, error) {
	// Need to build out a memory setup and look the damn thing up :(
	mem := cm.GetStore()
	if screen, ok := mem.ValidScreens[psID]; ok {
		return &screen, nil
	}
	return nil, errors.New("Screen not found")
}

// It really seems like it would be nicer to have a base class do this...
func (cm ContentManagerMemory) ListAllTags(page int, perPage int) (*models.Tags, error) {
	log.Printf("Using memory manager for tag page %d perPage %d \n", page, perPage)
	t_arr := models.Tags{}
	mem := cm.GetStore()
	for _, t := range mem.ValidTags {
		t_arr = append(t_arr, t)
	}
	if len(t_arr) == 0 {
		return &t_arr, nil
	}

	sort.SliceStable(t_arr, func(i, j int) bool {
		return t_arr[i].ID < t_arr[j].ID
	})
	offset, end := GetOffsetEnd(page, perPage, len(t_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		t_arr = t_arr[offset:end]
		return &t_arr, nil
	}
	return nil, errors.New("Not implemented")
}

func (cm ContentManagerMemory) GetTag(id string) (*models.Tag, error) {
	mem := cm.GetStore()
	if tag, ok := mem.ValidTags[id]; ok {
		return &tag, nil
	}
	return nil, errors.New(fmt.Sprintf("Tag not found %s", id))
}

func (cm ContentManagerMemory) ListAllTagsContext() (*models.Tags, error) {
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListAllTags(page, limit)
}

func (cm ContentManagerMemory) CreateTag(tag *models.Tag) error {
	if tag != nil {
		_, err := cm.GetStore().CreateTag(tag)
		return err
	}
	return errors.New("ContentManagerMemory no tag provided.")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateTag(t *models.Tag) error {
	// TODO: Validate that this updates the actual reference in mem storage
	if t != nil {
		_, err := cm.GetStore().UpdateTag(t)
		return err
	}
	return errors.New("ContentManagerMemory Update failed tag not provided")
}

func (cm ContentManagerMemory) DestroyTag(id string) (*models.Tag, error) {
	mem := cm.GetStore()
	if t, ok := mem.ValidTags[id]; ok {
		delete(mem.ValidTags, t.ID)
		return &t, nil
	}
	return nil, errors.New("ContentManagerMemory Destroy failed, no tag found.")
}

func (cm ContentManagerMemory) AssociateTag(t *models.Tag, mc *models.Content) error {
	if t == nil || mc == nil {
		return errors.New(fmt.Sprintf("Cannot associate missing tag %s or content %s", t, mc))
	}
	tag, err := cm.GetTag(t.ID)
	content, cErr := cm.GetContent(mc.ID)

	if err == nil && tag != nil && cErr == nil && content != nil {
		tags := mc.Tags
		if tags == nil {
			tags = models.Tags{}
		}
		found := false
		for _, check := range tags {
			if check.ID == t.ID {
				found = true
			}
		}
		if found == false {
			tags = append(tags, *tag)
		}
		content.Tags = tags
		return cm.UpdateContent(content)
	}
	return errors.New(fmt.Sprintf("Tag %s not in the list of valid tags", t))
}

func (cm ContentManagerMemory) AssociateTagByID(tagId string, mcID uuid.UUID) error {
	t, err := cm.GetTag(tagId)
	if err == nil && t != nil {
		content, cErr := cm.GetContent(mcID)
		if cErr == nil {
			return cm.AssociateTag(t, content)
		}
		return errors.New(fmt.Sprintf("Did not find content %s or err %s", content, cErr))
	}
	msg := fmt.Sprintf("Failed to find either the tag %s or error %s", t, err)
	log.Printf(msg)
	return errors.New(msg)
}

func AssignID(id uuid.UUID) uuid.UUID {
	emptyID, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
	if id == emptyID {
		newID, _ := uuid.NewV4()
		return newID
	}
	return id
}

// TODO: Fix this so that the screen must be under the
func (cm ContentManagerMemory) CreateScreen(screen *models.Screen) error {
	// Validate that the content exists for the screen?
	if screen != nil {
		_, err := cm.GetStore().CreateScreen(screen)
		return err
	}
	return errors.New("ContentManagerMemory no screen instance was passed in to CreateScreen")
}

// TODO: Requires security checks like the DB version.
func (cm ContentManagerMemory) CreateContent(content *models.Content) error {
	if content != nil {
		_, err := cm.GetStore().CreateContent(content)
		return err
	}
	return errors.New("ContentManagerMemory no Instance was passed in to CreateContent")
}

/**
* Not a thing in the memory manager
 */
func (cm ContentManagerMemory) DestroyContent(id string) (*models.Content, error) {
	return nil, errors.New("Not Implemented")
}

func (cm ContentManagerMemory) DestroyContainer(id string) (*models.Container, error) {
	return nil, errors.New("Not Implemented")
}

func (cm ContentManagerMemory) DestroyScreen(id string) (*models.Screen, error) {
	return nil, errors.New("Not Implemented")
}

// Note that we need to lock this down so that it cannot just access arbitrary files
func (cm ContentManagerMemory) CreateContainer(c *models.Container) error {
	if c == nil {
		return errors.New("ContentManagerMemory no container was passed in to CreateContainer")
	}
	cfg := cm.GetCfg()
	ok, err := utils.PathIsOk(c.Path, c.Name, cfg.Dir)
	if err != nil {
		log.Printf("Path does not exist on disk under the config directory err %s", err)
		return err
	}
	if ok == true {
		_, err := cm.GetStore().CreateContainer(c)
		return err
	}
	msg := fmt.Sprintf("The directory was not under the config path %s", c.Name)
	return errors.New(msg)
}

func (cm ContentManagerMemory) CreateTask(t *models.TaskRequest) (*models.TaskRequest, error) {
	if t == nil {
		return nil, errors.New("Requires a valid task")
	}
	mem := cm.GetStore()
	task, err := mem.CreateTask(t)
	if err != nil {
		return nil, err
	}
	return cm.GetTask(task.ID)
}

// Updates and creates will need to actually fully refresh things for background tasks to actually work
func (cm ContentManagerMemory) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequest, error) {
	// Probably does NOT properly update the memStorage
	mem := cm.GetStore()
	_, err := mem.UpdateTask(t, currentState)
	if err != nil {
		log.Printf("Couldn't find task to update %s", err)
		return nil, err
	}
	return cm.GetTask(t.ID)
}

func (cm ContentManagerMemory) GetTask(id uuid.UUID) (*models.TaskRequest, error) {
	mem := cm.GetStore()
	for idx, task := range mem.ValidTasks {
		if task.ID == id {
			return &mem.ValidTasks[idx], nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Task not found %s", id))
}

// Get the next task for processing (not super thread safe but enough for mem manager)
// Where we will ensure only 1 reader.
func (cm ContentManagerMemory) NextTask() (*models.TaskRequest, error) {
	mem := cm.GetStore()
	for _, task := range mem.ValidTasks {
		if task.Status == models.TaskStatus.NEW {
			task.Status = models.TaskStatus.PENDING
			updated, err := cm.UpdateTask(&task, task.Status)
			if err != nil {
				return nil, err
			}
			return updated, nil
		}
	}
	return nil, nil
}

/*
*
 */
func (cm ContentManagerMemory) ListTasksContext() (*models.TaskRequests, error) {
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.GetCfg().Limit)
	query := TaskQuery{
		Page:      page,
		PerPage:   limit,
		ContentID: StringDefault(params.Get("content_id"), ""),
		Status:    StringDefault(params.Get("status"), ""),
	}
	return cm.ListTasks(query)
}

func (cm ContentManagerMemory) ListTasks(query TaskQuery) (*models.TaskRequests, error) {
	mem := cm.GetStore()
	task_arr := mem.ValidTasks
	if query.ContentID != "" {
		contentID, err := uuid.FromString(query.ContentID)
		filtered_tasks := models.TaskRequests{}
		if err != nil {
			return nil, err
		}
		for _, task := range task_arr {
			if task.ContentID == contentID {
				filtered_tasks = append(filtered_tasks, task)
			}
		}
		task_arr = filtered_tasks
	}
	if query.Status != "" {
		filtered_tasks := models.TaskRequests{}
		for _, task := range task_arr {
			if task.Status.String() == query.Status {
				filtered_tasks = append(filtered_tasks, task)
			}
		}
		task_arr = filtered_tasks
	}
	offset, end := GetOffsetEnd(query.Page, query.PerPage, len(task_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		task_arr = task_arr[offset:end]
		return &task_arr, nil
	}
	return &task_arr, nil
}

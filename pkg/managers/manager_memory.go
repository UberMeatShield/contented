/**
 * Implements the ContentManager interface but stores all information in memory.  This
 * is using the utility/MemStorage singleton which will load up the disk information only
 * one time.
 */
package managers

import (
	"contented/pkg/models"
	"contented/pkg/utils"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	Params   GetParamsType
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
func (cm *ContentManagerMemory) Initialize() *utils.MemoryStorage {

	// mem_storage.go
	memStorage := cm.GetStore()
	if !memStorage.Initialized {
		// Might need to instead throw if it is not initialized
		memStorage = utils.InitializeMemory(cm.cfg.Dir)
	}
	return memStorage
}

func (cm ContentManagerMemory) GetStore() *utils.MemoryStorage {
	return utils.GetMemStorage()
}

// Kinda strange but it seems hard to assign the type into an interface
// type GetParamsType func() *url.Values
func (cm ContentManagerMemory) GetParams() *url.Values {
	return cm.Params()
}

func (cm ContentManagerMemory) ListContentContext() (*models.Contents, int64, error) {
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.cfg.Limit)

	cID := StringDefault(params.Get("container_id"), "")
	// Note text is an exact match, search is a regex or partial
	cs := ContentQuery{
		Text:        StringDefault(params.Get("text"), ""),
		ContainerID: cID,
		Page:        page,
		PerPage:     limit,
		Order:       StringDefault(params.Get("order"), ""),
	}
	return cm.ListContent(cs)
}

// It should probably be able to search the container too?
func (cm ContentManagerMemory) SearchContentContext() (*models.Contents, int64, error) {
	sr := ContextToContentQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContent(sr)
}

// Memory version is going to be extra annoying to tag search more than one tag on an or, or AND...
func (cm ContentManagerMemory) SearchContent(sr ContentQuery) (*models.Contents, int64, error) {
	filteredContent, cErr := cm.getContentFiltered(sr)
	if cErr != nil {
		return nil, 0, cErr
	}
	if filteredContent == nil {
		empty := models.Contents{}
		return &empty, 0, nil
	}

	if len(sr.Tags) > 0 {
		log.Printf("Searching using tags query %s", sr.Tags)
		filteredContent = cm.tagSearch(filteredContent, sr.Tags)
	}

	mc_arr := *filteredContent
	count := len(mc_arr)
	offset, end := GetOffsetEnd(sr.Page, sr.PerPage, count)

	// Finally sort any content that is matching so that pagination will work
	sort.SliceStable(mc_arr, models.GetContentSort(mc_arr, sr.Order))
	if sr.Direction == "desc" {
		mc_arr = mc_arr.Reverse()
	}
	if end > 0 { // If it is empty a slice ending in 0 = boom
		mc_arr = mc_arr[offset:end]
		return &mc_arr, int64(count), nil
	}
	return &mc_arr, int64(count), nil
}

// This is not great but there isn't a lookup of tag => contents
func (cm ContentManagerMemory) tagSearch(contents *models.Contents, tags []string) *models.Contents {
	filteredContents := models.Contents{}

	// Hmmm, unsafe in some ways because the data may not be loaded so know that this works for memory
	// manager because the tags are associated by the API / testing.
	if contents != nil {
		cArr := *contents
		var content models.Content
		for _, el := range cArr {
			content = el
			for _, tag := range tags {
				if content.HasTag(tag) {
					filteredContents = append(filteredContents, content)
				}
			}
		}
	}
	return &filteredContents
}

// Search Request may still make more sense.
func (cm ContentManagerMemory) getContentFiltered(cs ContentQuery) (*models.Contents, error) {
	// If a containerID is specified and is totally invalid raise an error, otherwise filter
	var mcArr models.Contents
	cidArr := models.Contents{}
	mem := cm.GetStore()

	// Most common initial filtering call
	if cs.ContainerID != "" {
		cID, cErr := strconv.ParseInt(cs.ContainerID, 10, 64)
		if cErr == nil {
			for _, mc := range mem.ValidContent {
				if mc.ContainerID != nil && *mc.ContainerID == cID {
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

	if !cs.IncludeHidden {
		visibleArr := models.Contents{}
		for _, mc := range mcArr {
			if !mc.Hidden {
				visibleArr = append(visibleArr, mc)
			}
		}
		mcArr = visibleArr
	}

	if id, err := strconv.ParseInt(cs.ContentID, 10, 32); err == nil {
		idArr := models.Contents{}
		for _, mc := range mcArr {
			if mc.ID == id {
				idArr = append(idArr, mc)
			}
		}
		mcArr = idArr
	}

	if cs.Search != "" && cs.Search != "*" {
		log.Printf("It should be searching the contents %s", cs.Search)
		searchStr := regexp.QuoteMeta(cs.Search)
		searcher := regexp.MustCompile("(?i)" + searchStr)
		searchArr := models.Contents{}
		for _, mc := range mcArr {
			if searcher.MatchString(mc.Src) {
				searchArr = append(searchArr, mc)
			}
		}
		mcArr = searchArr
	}

	if cs.Text != "" {
		nameArr := models.Contents{}
		for _, content := range mcArr {
			if content.Src == cs.Text {
				nameArr = append(nameArr, content)
			}
		}
		mcArr = nameArr
	}

	if cs.ContentType != "" && cs.ContentType != "*" {
		searcher := regexp.MustCompile(cs.ContentType)
		contentArr := models.Contents{}
		for _, mc := range mcArr {
			if searcher.MatchString(mc.ContentType) {
				contentArr = append(contentArr, mc)
			}
		}
		mcArr = contentArr
	}
	return &mcArr, nil
}

// It should probably be able to search the container too?
func (cm ContentManagerMemory) SearchContainersContext() (*models.Containers, int64, error) {
	cq := ContextToContainerQuery(cm.Params(), cm.GetCfg())
	return cm.SearchContainers(cq)
}

// TODO: Make it page but right now this will only be used in splash (regex it?)
func (cm ContentManagerMemory) SearchContainers(cs ContainerQuery) (*models.Containers, int64, error) {
	limit := cs.PerPage
	cs.PerPage = 90000 // Search everything in the filtered section
	cnts, _, cErr := cm.ListContainersFiltered(cs)
	if cErr != nil {
		return nil, -1, cErr
	}
	if cnts == nil {
		cnts = &models.Containers{}
	}
	cArr := models.Containers{}
	searcher := regexp.MustCompile("(?i)" + cs.Search)
	for _, c := range *cnts {
		if searcher.MatchString(c.Name) {
			cArr = append(cArr, c)
		}
	}

	offset, end := GetOffsetEnd(cs.Page, limit, len(cArr))
	sort.SliceStable(cArr, models.GetContainerSort(cArr, cs.Order))
	if cs.Direction == "desc" {
		cArr = cArr.Reverse()
	}
	count := int64(len(cArr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		cArr = cArr[offset:end]
		return &cArr, count, nil
	}
	return &cArr, count, nil
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListContent(cs ContentQuery) (*models.Contents, int64, error) {
	cs.IncludeHidden = false
	return cm.ListContentFiltered(cs)
}

func (cm ContentManagerMemory) ListContentFiltered(cs ContentQuery) (*models.Contents, int64, error) {
	m_arr := models.Contents{}
	mem := cm.GetStore()

	log.Printf("What the shit %s", cs.ContainerID)
	// Need to test invalid / empty ""
	containerID, invalid := strconv.ParseInt(cs.ContainerID, 10, 64)
	if invalid == nil {
		for _, content := range mem.ValidContent {
			if content.ContainerID != nil && *content.ContainerID == containerID {
				m_arr = append(m_arr, content)
			}
		}
	} else {
		for _, content := range mem.ValidContent {
			m_arr = append(m_arr, content)
		}
	}

	if contentID, badIdErr := strconv.ParseInt(cs.ContentID, 10, 64); badIdErr == nil {
		for _, content := range m_arr {
			if content.ID == contentID {
				m_arr = models.Contents{content}
				break
			}
		}
	}

	if cs.ContentType != "" {
		ct_arr := models.Contents{}
		for _, content := range m_arr {
			if strings.Contains(content.ContentType, cs.ContentType) {
				ct_arr = append(ct_arr, content)
			}
		}
		m_arr = ct_arr
	}

	h_arr := models.Contents{}
	for _, m := range m_arr {
		if !cs.IncludeHidden {
			if !m.Hidden {
				h_arr = append(h_arr, m)
			}
		} else {
			h_arr = append(h_arr, m)
		}
	}
	m_arr = h_arr

	sort.SliceStable(m_arr, models.GetContentSort(m_arr, cs.Order))
	count := len(m_arr)
	offset, end := GetOffsetEnd(cs.Page, cs.PerPage, len(m_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		m_arr = m_arr[offset:end]
		return &m_arr, int64(count), nil
	}
	// log.Printf("Get a list of content offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
	return &m_arr, int64(count), nil

}

// Get a content element by the ID
func (cm ContentManagerMemory) GetContent(mcID int64) (*models.Content, error) {
	// log.Printf("Memory Get a single content %s", mcID)
	mem := cm.GetStore()
	if mc, ok := mem.ValidContent[mcID]; ok {
		return &mc, nil
	}
	return nil, errors.New("content was not found in memory")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateContainer(cnt *models.Container) (*models.Container, error) {
	// TODO: Validate that this updates the actual reference in mem storage
	cfg := cm.GetCfg()
	pathOk, err := utils.PathIsOk(cnt.Path, cnt.Name, cfg.Dir)
	if err != nil || !pathOk {
		log.Printf("Path does not exist on disk under the config directory err %s", err)
		return nil, err
	}
	container, err := cm.GetStore().UpdateContainer(cnt)
	return container, err
}

// No updates should be allowed for memory management.
func (cm ContentManagerMemory) UpdateContent(content *models.Content) error {
	// TODO: Should I be able to ignore being in a container if there is no file?
	if !content.NoFile {
		cnt, cErr := cm.GetContainer(*content.ContainerID)
		if cErr != nil {
			msg := fmt.Sprintf("parent container %d not found", content.ContainerID)
			log.Print(msg)
			return errors.New(msg)
		}
		// Check if file exists or allow content to be 'empty'?
		exists, pErr := utils.HasContent(content.Src, cnt.GetFqPath())
		if !exists || pErr != nil {
			log.Printf("Content not in container %s", pErr)
			return fmt.Errorf("invalid content src %s for container %s", content.Src, cnt.Name)
		}
	}

	tags, tagErr := cm.GetValidTags(&content.Tags)
	if tagErr != nil {
		return tagErr
	}

	content.Tags = *tags
	_, err := cm.GetStore().UpdateContent(content)
	return err
}

func (cm ContentManagerMemory) UpdateContents(contents models.Contents) error {
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

// Could use some extra validation (ensure there is content for the screen?)
func (cm ContentManagerMemory) UpdateScreen(s *models.Screen) error {
	_, err := cm.GetStore().UpdateScreen(s)
	return err
}

// Given the current parameters in the gin context return a list of matching containers.
func (cm ContentManagerMemory) ListContainersContext() (*models.Containers, int64, error) {
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.cfg.Limit)
	cs := ContainerQuery{
		Page:    page,
		PerPage: limit,
		Name:    StringDefault(params.Get("name"), ""),
		Order:   StringDefault(params.Get("order"), ""),
	}
	return cm.ListContainers(cs)
}

// Actually list containers using a page and per_page consistent with pagination.
func (cm ContentManagerMemory) ListContainers(cs ContainerQuery) (*models.Containers, int64, error) {
	return cm.ListContainersFiltered(cs)
}

func (cm ContentManagerMemory) ListContainersFiltered(cs ContainerQuery) (*models.Containers, int64, error) {
	c_arr := models.Containers{}
	mem := cm.GetStore()
	for _, c := range mem.ValidContainers {
		if cs.Name != "" && !strings.Contains(c.Name, cs.Name) {
			continue
		}
		if !cs.IncludeHidden {
			if !c.Hidden {
				c_arr = append(c_arr, c)
			}
		} else {
			c_arr = append(c_arr, c)
		}
	}
	sort.SliceStable(c_arr, func(i, j int) bool {
		return c_arr[i].Idx < c_arr[j].Idx
	})
	count := len(c_arr)
	offset, end := GetOffsetEnd(cs.Page, cs.PerPage, count)
	c_arr = c_arr[offset:end]
	return &c_arr, int64(count), nil
}

// Get a single container given the primary key
func (cm ContentManagerMemory) GetContainer(cID int64) (*models.Container, error) {
	// log.Printf("Get a single container %s", cID)
	mem := cm.GetStore()
	if c, ok := mem.ValidContainers[cID]; ok {
		return &c, nil
	}
	return nil, fmt.Errorf("ContentMemoryManager did not find this container id: %d", cID)
}

func (cm ContentManagerMemory) GetPreviewForMC(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(*mc.ContainerID)
	if err != nil {
		return "Memory Manager Preview no Parent Found", err
	}
	src := mc.Src
	if mc.Preview != "" {
		src = mc.Preview
	}
	log.Printf("ContentMemoryManager loading %d preview %s\n", mc.ID, src)
	return utils.GetFilePathInContainer(src, cnt.GetFqPath())
}

func (cm ContentManagerMemory) FindActualFile(mc *models.Content) (string, error) {
	cnt, err := cm.GetContainer(*mc.ContainerID)
	if err != nil {
		return "Memory Manager View no Parent Found", err
	}
	log.Printf("Memory Manager View %d loading up %s\n", mc.ID, mc.Src)
	return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

// If you want to do in memory testing and already manually created previews this will
// then try and use the previews for the in memory manager.
func (cm ContentManagerMemory) SetPreviewIfExists(mc *models.Content) (string, error) {
	c, err := cm.GetContainer(*mc.ContainerID)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	pFile := utils.AssignPreviewIfExists(c, mc)
	return pFile, nil
}

func (cm ContentManagerMemory) ListScreensContext() (*models.Screens, int64, error) {
	// Could add the context here correctly
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.cfg.Limit)
	sr := ScreensQuery{
		Page:      page,
		PerPage:   limit,
		ContentID: params.Get("content_id"),
		Order:     StringDefault(params.Get("order"), ""),
	}
	return cm.ListScreens(sr)
}

// TODO: Get a pattern for each MC, look at a preview Destination, then match against the pattern
// And build out a set of screens.
func (cm ContentManagerMemory) ListScreens(sq ScreensQuery) (*models.Screens, int64, error) {

	// Did I create this just to sort by Idx across all content?  Kinda strange
	mem := cm.GetStore()
	s_arr := models.Screens{}
	if sq.ContentID != "" {
		contentID, idErr := strconv.ParseInt(sq.ContentID, 10, 32)
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
	sort.SliceStable(s_arr, models.GetScreensSort(s_arr, sq.Order))
	if sq.Direction == "desc" {
		s_arr = s_arr.Reverse()
	}
	count := len(s_arr)
	offset, end := GetOffsetEnd(sq.Page, sq.PerPage, count)
	if end > 0 { // If it is empty a slice ending in 0 = boom
		s_arr = s_arr[offset:end]
		return &s_arr, int64(count), nil
	}
	return &s_arr, int64(count), nil
}

func (cm ContentManagerMemory) ClearScreens(content *models.Content) error {
	if content == nil || content.ID == 0 {
		return fmt.Errorf("cannot clear screens without content or a valid id")
	}
	// Update the screens reference
	content.Screens = models.Screens{}

	screenMap := cm.GetStore().ValidScreens
	screens := []int64{}
	for _, screen := range screenMap {
		if screen.ContentID == content.ID {
			screens = append(screens, screen.ID)
		}
	}
	// Remove the in memory screens
	for _, screenID := range screens {
		delete(screenMap, screenID)
	}
	return nil
}

func (cm ContentManagerMemory) GetScreen(psID int64) (*models.Screen, error) {
	// Need to build out a memory setup and look the damn thing up :(
	mem := cm.GetStore()
	if screen, ok := mem.ValidScreens[psID]; ok {
		return &screen, nil
	}
	return nil, errors.New("screen not found")
}

// It really seems like it would be nicer to have a base class do this...
func (cm ContentManagerMemory) ListAllTags(tq TagQuery) (*models.Tags, int64, error) {
	log.Printf("Using memory manager for tag page %d perPage %d \n", tq.Page, tq.PerPage)
	t_arr := models.Tags{}
	mem := cm.GetStore()
	for _, t := range mem.ValidTags {
		if tq.TagType == "" || tq.TagType == t.TagType {
			t_arr = append(t_arr, t)
		}
	}
	if len(t_arr) == 0 {
		return &t_arr, 0, nil
	}

	sort.SliceStable(t_arr, func(i, j int) bool {
		return t_arr[i].ID < t_arr[j].ID
	})
	total := len(t_arr)
	offset, end := GetOffsetEnd(tq.Page, tq.PerPage, total)

	if end > 0 { // If it is empty a slice ending in 0 = boom
		t_arr = t_arr[offset:end]
		return &t_arr, int64(total), nil
	}
	return &t_arr, int64(total), fmt.Errorf("invalid page end %d per page %d", tq.Page, tq.PerPage)
}

func (cm ContentManagerMemory) GetTag(id string) (*models.Tag, error) {
	mem := cm.GetStore()
	if tag, ok := mem.ValidTags[id]; ok {
		return &tag, nil
	}
	return nil, fmt.Errorf("tag not found %s", id)
}

func (cm ContentManagerMemory) ListAllTagsContext() (*models.Tags, int64, error) {
	params := cm.Params()
	_, limit, page := GetPagination(params, cm.cfg.Limit)
	tq := TagQuery{
		Page:    page,
		PerPage: limit,
		TagType: StringDefault(params.Get("tag_type"), ""),
	}
	return cm.ListAllTags(tq)
}

func (cm ContentManagerMemory) CreateTag(tag *models.Tag) error {
	if tag != nil {
		_, err := cm.GetStore().CreateTag(tag)
		return err
	}
	return errors.New("ContentManagerMemory no tag provided")
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
	return nil, errors.New("ContentManagerMemory destroy failed, no tag found")
}

func (cm ContentManagerMemory) AssociateTag(t *models.Tag, mc *models.Content) error {
	if t == nil || mc == nil {
		return fmt.Errorf("cannot associate missing tag %s or content %s", t, mc)
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
		if !found {
			tags = append(tags, *tag)
		}
		content.Tags = tags
		return cm.UpdateContent(content)
	}
	return fmt.Errorf("tag %s not in the list of valid tags", t)
}

func (cm ContentManagerMemory) AssociateTagByID(tagId string, mcID int64) error {
	t, err := cm.GetTag(tagId)
	if err == nil && t != nil {
		content, cErr := cm.GetContent(mcID)
		if cErr == nil {
			return cm.AssociateTag(t, content)
		}
		return fmt.Errorf("did not find content %s or err %s", content, cErr)
	}
	msg := fmt.Sprintf("Failed to find either the tag %s or error %s", t, err)
	log.Print(msg)
	return errors.New(msg)
}

func (cm ContentManagerMemory) CreateScreen(screen *models.Screen) error {
	// Validate that the content exists for the screen?
	if screen != nil {
		content, notFound := cm.GetContent(screen.ContentID)
		if content == nil || notFound != nil {
			return fmt.Errorf("a screen must be linked to content and %d was not found", screen.ContentID)
		}
		_, err := cm.GetStore().CreateScreen(screen)
		return err
	}
	return errors.New("ContentManagerMemory no screen instance was passed in to CreateScreen")
}

// TODO: Requires security checks like the DB version.
func (cm ContentManagerMemory) CreateContent(content *models.Content) error {
	if content != nil {
		if content.Tags == nil {
			content.Tags = models.Tags{}
		}
		validTags, tErr := cm.GetValidTags(&content.Tags)
		if tErr != nil {
			log.Printf("Failed to find valid tags %s", tErr)
		}
		content.Tags = *validTags
		_, err := cm.GetStore().CreateContent(content)
		return err
	}
	return errors.New("ContentManagerMemory no Instance was passed in to CreateContent")
}

func (cm ContentManagerMemory) GetValidTags(tags *models.Tags) (*models.Tags, error) {

	goodTags := models.Tags{}
	if tags == nil {
		return &goodTags, nil
	}

	validTags := cm.GetStore().ValidTags
	for _, tag := range *tags {
		if _, ok := validTags[tag.ID]; ok {
			goodTags = append(goodTags, tag)
		}
	}
	return &goodTags, nil
}

/**
* Note that these methods are mostly for consistent API but do NOT cleanup references.
 */
func (cm ContentManagerMemory) DestroyContent(id int64) (*models.Content, error) {
	contentMap := cm.GetStore().ValidContent
	if content, ok := contentMap[id]; ok {
		delete(contentMap, id)
		return &content, nil
	}
	return nil, fmt.Errorf("Content not found %d", id)
}

func (cm ContentManagerMemory) DestroyContainer(id int64) (*models.Container, error) {
	containerMap := cm.GetStore().ValidContainers
	if container, ok := containerMap[id]; ok {
		delete(containerMap, id)
		return &container, nil
	}
	return nil, fmt.Errorf("Container not found %d", id)
}

func (cm ContentManagerMemory) DestroyScreen(id int64) (*models.Screen, error) {
	screensMap := cm.GetStore().ValidScreens
	if screen, ok := screensMap[id]; ok {
		delete(screensMap, id)
		return &screen, nil
	}
	return nil, fmt.Errorf("Screen not found %d", id)
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
	if ok {
		_, err := cm.GetStore().CreateContainer(c)
		return err
	}
	msg := fmt.Sprintf("The directory was not under the config path %s", c.Name)
	return errors.New(msg)
}

func (cm ContentManagerMemory) CreateTask(t *models.TaskRequest) (*models.TaskRequest, error) {
	if t == nil {
		return nil, errors.New("requires a valid task")
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

func (cm ContentManagerMemory) GetTask(id int64) (*models.TaskRequest, error) {
	mem := cm.GetStore()
	for idx, task := range mem.ValidTasks {
		if task.ID == id {
			return &mem.ValidTasks[idx], nil
		}
	}
	return nil, fmt.Errorf("task not found %d", id)
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
func (cm ContentManagerMemory) ListTasksContext() (*models.TaskRequests, int64, error) {
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

func (cm ContentManagerMemory) ListTasks(query TaskQuery) (*models.TaskRequests, int64, error) {
	mem := cm.GetStore()
	task_arr := mem.ValidTasks
	if query.ContentID != "" {
		contentID, err := strconv.ParseInt(query.ContentID, 10, 64)
		filtered_tasks := models.TaskRequests{}
		if err != nil {
			return nil, 0, err
		}
		for _, task := range task_arr {
			if task.ContentID != nil && *task.ContentID == contentID {
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
	total := len(task_arr)
	offset, end := GetOffsetEnd(query.Page, query.PerPage, total)
	if end > 0 { // If it is empty a slice ending in 0 = boom
		task_arr = task_arr[offset:end]
		return &task_arr, int64(total), nil
	}
	return &task_arr, int64(total), nil
}

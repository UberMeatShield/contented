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
	ValidContent    models.ContentMap
	ValidContainers models.ContainerMap
	ValidScreens    models.ScreenMap
	ValidTags       models.TagsMap
	validate        string

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
	memStorage := utils.GetMemStorage()
	if memStorage.Initialized == false {
		memStorage = utils.InitializeMemory(cm.cfg.Dir)
	}
	cm.ValidContainers = memStorage.ValidContainers
	cm.ValidContent = memStorage.ValidContent
	cm.ValidScreens = memStorage.ValidScreens
	cm.ValidTags = memStorage.ValidTags
	log.Printf("Found %d directories with %d content elements \n", len(cm.ValidContainers), len(cm.ValidContent))
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
	for _, m := range cm.ValidContent {
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
	sr := ContextToSearchRequest(cm.Params(), cm.GetCfg())
	return cm.SearchContent(sr)
}

// Memory version is going to be extra annoying to tag search more than one tag on an or, or AND...
func (cm ContentManagerMemory) SearchContent(sr SearchRequest) (*models.Contents, int, error) {
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

func (cm ContentManagerMemory) getContentFiltered(containerID string, search string, contentType string, includeHidden bool) (*models.Contents, error) {
	// If a containerID is specified and is totally invalid raise an error, otherwise filter
	var mcArr models.Contents
	cidArr := models.Contents{}

	if containerID != "" {
		cID, cErr := uuid.FromString(containerID)
		if cErr == nil {
			for _, mc := range cm.ValidContent {
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
		for _, mc := range cm.ValidContent {
			cidArr = append(cidArr, mc)
		}
		mcArr = cidArr
	}

	if search != "" && search != "*" {
		searcher := regexp.MustCompile("(?i)" + search)
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
	for _, c := range cm.ValidContainers {
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
	for _, m := range cm.ValidContent {
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
	if mc, ok := cm.ValidContent[mcID]; ok {
		return &mc, nil
	}
	return nil, errors.New("Content was not found in memory")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateContainer(cnt *models.Container) (*models.Container, error) {
	// TODO: Validate that this updates the actual reference in mem storage
	cfg := cm.GetCfg()
	pathOk, err := utils.PathIsOk(cnt.Path, cnt.Name, cfg.Dir)
	if err != nil {
		log.Printf("Path does not exist on disk under the config directory err %s", err)
		return nil, err
	}
	if _, ok := cm.ValidContainers[cnt.ID]; ok && pathOk {
		cm.ValidContainers[cnt.ID] = *cnt
		return cnt, nil
	}
	return nil, errors.New("Container was not found to update or path illegal")
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
	if _, ok := cm.ValidContent[content.ID]; ok {
		cm.ValidContent[content.ID] = *content
		return nil
	}
	return errors.New("Content was not found to update")
}

func (cm ContentManagerMemory) UpdateScreen(s *models.Screen) error {
	if _, ok := cm.ValidScreens[s.ID]; ok {
		cm.ValidScreens[s.ID] = *s
		return nil
	}
	return errors.New("Content was not found to update")
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
	for _, c := range cm.ValidContainers {
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
	log.Printf("Get a single container %s", cID)
	if c, ok := cm.ValidContainers[cID]; ok {
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

func (cm ContentManagerMemory) ListScreensContext(mcID uuid.UUID) (*models.Screens, error) {
	// Could add the context here correctly
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListScreens(mcID, page, limit)
}

// TODO: Get a pattern for each MC, look at a preview Destination, then match against the pattern
// And build out a set of screens.
func (cm ContentManagerMemory) ListScreens(mcID uuid.UUID, page int, per_page int) (*models.Screens, error) {

	// Did I create this just to sort by Idx across all content?  Kinda strange
	s_arr := models.Screens{}
	for _, s := range cm.ValidScreens {
		if s.ContentID == mcID {
			s_arr = append(s_arr, s)
		}
	}
	sort.SliceStable(s_arr, func(i, j int) bool {
		return s_arr[i].Idx < s_arr[j].Idx
	})
	offset, end := GetOffsetEnd(page, per_page, len(s_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		s_arr = s_arr[offset:end]
		return &s_arr, nil
	}
	return &s_arr, nil
}

func (cm ContentManagerMemory) ListAllScreensContext() (*models.Screens, error) {
	_, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
	return cm.ListAllScreens(page, limit)
}

func (cm ContentManagerMemory) ListAllScreens(page int, per_page int) (*models.Screens, error) {

	log.Printf("Using memory manager for screen page %d per_page %d \n", page, per_page)
	// Did I create this just to sort by Idx across all content?  Kinda strange
	s_arr := models.Screens{}
	for _, s := range cm.ValidScreens {
		s_arr = append(s_arr, s)
	}
	sort.SliceStable(s_arr, func(i, j int) bool {
		return s_arr[i].Idx < s_arr[j].Idx
	})
	offset, end := GetOffsetEnd(page, per_page, len(s_arr))
	if end > 0 { // If it is empty a slice ending in 0 = boom
		s_arr = s_arr[offset:end]
		return &s_arr, nil
	}
	return &s_arr, nil
}

func (cm ContentManagerMemory) GetScreen(psID uuid.UUID) (*models.Screen, error) {
	// Need to build out a memory setup and look the damn thing up :(
	memStorage := utils.GetMemStorage()
	if screen, ok := memStorage.ValidScreens[psID]; ok {
		return &screen, nil
	}
	return nil, errors.New("Screen not found")
}

// It really seems like it would be nicer to have a base class do this...
func (cm ContentManagerMemory) ListAllTags(page int, perPage int) (*models.Tags, error) {
	log.Printf("Using memory manager for tag page %d perPage %d \n", page, perPage)
	t_arr := models.Tags{}
	for _, t := range cm.ValidTags {
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
	if tag, ok := cm.ValidTags[id]; ok {
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
		cm.ValidTags[tag.ID] = *tag
		log.Printf("Created tag %s", tag)
		return nil
	}
	return errors.New("ContentManagerMemory no tag provided.")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateTag(t *models.Tag) error {
	// TODO: Validate that this updates the actual reference in mem storage
	if _, ok := cm.ValidTags[t.ID]; ok {
		cm.ValidTags[t.ID] = *t
		return nil
	}
	return errors.New("ContentManagerMemory Update failed, not found.")
}

func (cm ContentManagerMemory) DestroyTag(id string) (*models.Tag, error) {
	if t, ok := cm.ValidTags[id]; ok {
		delete(cm.ValidTags, t.ID)
		return &t, nil
	}
	return nil, errors.New("ContentManagerMemory Destroy failed, not tag found.")
}

func (cm ContentManagerMemory) AssociateTag(t *models.Tag, mc *models.Content) error {
	if t == nil || mc == nil {
		return errors.New(fmt.Sprintf("Cannot associate missing tag %s or content %s", t, mc))
	}
	if tag, ok := cm.ValidTags[t.ID]; ok {
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
			tags = append(tags, tag)
		}
		mc.Tags = tags
		return cm.UpdateContent(mc)
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
	if screen != nil {
		screen.ID = AssignID(screen.ID)
		cm.ValidScreens[screen.ID] = *screen
		return nil
	}
	return errors.New("ContentManagerMemory no screen instance was passed in to CreateScreen")
}

// TODO: Requires security checks like the DB version.
func (cm ContentManagerMemory) CreateContent(mc *models.Content) error {
	if mc != nil {
		mc.ID = AssignID(mc.ID)
		cm.ValidContent[mc.ID] = *mc
		return nil
	}
	return errors.New("ContentManagerMemory no Instance was passed in to CreateContent")
}

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
		c.ID = AssignID(c.ID)
		cm.ValidContainers[c.ID] = *c
		return nil
	}
	msg := fmt.Sprintf("The directory was not under the config path %s", c.Name)
	return errors.New(msg)
}

func (cm ContentManagerMemory) AddTask(t *models.TaskRequest) error {
	return errors.New("Not implemented")
}

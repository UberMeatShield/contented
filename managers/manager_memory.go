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
    "github.com/gofrs/uuid"
    "log"
    "net/url"
    "regexp"
    "sort"
)

// Provides the support for looking up media by ID while only using memory
type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry

    // Hmmm, this should use the memory manager probably
    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    ValidScreens    models.PreviewScreenMap
    validate        string

    params *url.Values
    Params GetParamsType
}

// We do not allow editing in a memory manager
func (cm ContentManagerMemory) CanEdit() bool {
    return false
}

// Provide the ability to set the configuration for a memory manager.
func (cm *ContentManagerMemory) SetCfg(cfg *utils.DirConfigEntry) {
    cm.cfg = cfg
    log.Printf("It should have a preview %s\n", cm.cfg.Dir)
    log.Printf("Using memory manager %s\n", cm.validate)
}

// Get the currently configuration for this manager.
func (cm ContentManagerMemory) GetCfg() *utils.DirConfigEntry {
    log.Printf("Get the Config validate val %s\n", cm.validate)
    log.Printf("Config is using path %s", cm.cfg.Dir)
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
    cm.ValidMedia = memStorage.ValidMedia
    cm.ValidScreens = memStorage.ValidScreens
    log.Printf("Found %d directories with %d media elements \n", len(cm.ValidContainers), len(cm.ValidMedia))
}

// Kinda strange but it seems hard to assign the type into an interface
// type GetParamsType func() *url.Values
func (cm ContentManagerMemory) GetParams() *url.Values {
    return cm.Params()
}

func (cm ContentManagerMemory) ListMediaContext(cID uuid.UUID) (*models.MediaContainers, error) {
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListMedia(cID, page, limit)
}

func (cm ContentManagerMemory) ListAllMedia(page int, per_page int) (*models.MediaContainers, error) {
    m_arr := models.MediaContainers{}

    // Did I create this just to sort by Idx across all media?  Kinda strange
    for _, m := range cm.ValidMedia {
        m_arr = append(m_arr, m)
    }
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
func (cm ContentManagerMemory) SearchMediaContext() (*models.MediaContainers, int, error) {
    params := cm.Params()
    _, per_page, page := GetPagination(params, cm.cfg.Limit)
    searchStr := StringDefault(params.Get("text"), "")
    cId := StringDefault(params.Get("cID"), "")
    contentType := StringDefault(params.Get("contentType"), "")
    return cm.SearchMedia(searchStr, page, per_page, cId, contentType)
}

func (cm ContentManagerMemory) SearchMedia(search string, page int, per_page int, cID string, contentType string) (*models.MediaContainers, int, error) {
    filteredMedia, cErr := cm.getMediaFiltered(cID, search, contentType)
    if cErr != nil {
        return nil, 0, cErr
    }
    if filteredMedia == nil {
        empty := models.MediaContainers{}
        return &empty, 0, nil
    }

    mc_arr := *filteredMedia
    count := len(mc_arr)
    offset, end := GetOffsetEnd(page, per_page, count)
    if end > 0 { // If it is empty a slice ending in 0 = boom
        mc_arr = mc_arr[offset:end]
        return &mc_arr, count, nil
    }
    return &mc_arr, count, nil
}

func (cm ContentManagerMemory) getMediaFiltered(containerID string, search string, contentType string) (*models.MediaContainers, error) {
    // If a containerID is specified and is totally invalid raise an error, otherwise filter
    var mcArr models.MediaContainers
    cidArr := models.MediaContainers{}
    if containerID != "" {
        cID, cErr := uuid.FromString(containerID)
        if cErr == nil {
            for _, mc := range cm.ValidMedia {
                if mc.ContainerID.Valid && mc.ContainerID.UUID == cID {
                    cidArr = append(cidArr, mc)
                }
            }
            mcArr = cidArr
        } else {
            return nil, cErr
        }
    } else {
        // Empty string for containerID is considered match all media
        for _, mc := range cm.ValidMedia {
            cidArr = append(cidArr, mc)
        }
        mcArr = cidArr
    }

    if search != "" && search != "*" {
        searcher := regexp.MustCompile(search)
        searchArr := models.MediaContainers{}
        for _, mc := range mcArr {
            if searcher.MatchString(mc.Src) {
                searchArr = append(searchArr, mc)
            }
        }
        mcArr = searchArr
    }

    if contentType != "" && contentType != "*" {
        searcher := regexp.MustCompile(contentType)
        contentArr := models.MediaContainers{}
        for _, mc := range mcArr {
            if searcher.MatchString(mc.ContentType) {
                contentArr = append(contentArr, mc)
            }
        }
        mcArr = contentArr
    }

    // Finally sort any content that is matching so that pagination will work
    sort.SliceStable(mcArr, func(i, j int) bool {
        return mcArr[i].Idx < mcArr[j].Idx
    })
    return &mcArr, nil
}

// Awkard GoLang interface support is awkward
func (cm ContentManagerMemory) ListMedia(ContainerID uuid.UUID, page int, per_page int) (*models.MediaContainers, error) {
    m_arr := models.MediaContainers{}
    for _, m := range cm.ValidMedia {
        if m.ContainerID.Valid && m.ContainerID.UUID == ContainerID {
            m_arr = append(m_arr, m)
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
    log.Printf("Get a list of media offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
    return &m_arr, nil
}

// Get a media element by the ID
func (cm ContentManagerMemory) GetMedia(mcID uuid.UUID) (*models.MediaContainer, error) {
    // log.Printf("Memory Get a single media %s", mcID)
    if mc, ok := cm.ValidMedia[mcID]; ok {
        return &mc, nil
    }
    return nil, errors.New("Media was not found in memory")
}

// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateContainer(c *models.Container) error {
    // TODO: Validate that this updates the actual reference in mem storage
    if _, ok := cm.ValidContainers[c.ID]; ok {
        cm.ValidContainers[c.ID] = *c
        return nil
    }
    return errors.New("Container was not found to update")
}

// No updates should be allowed for memory management.
func (cm ContentManagerMemory) UpdateMedia(mc *models.MediaContainer) error {
    if _, ok := cm.ValidMedia[mc.ID]; ok {
        cm.ValidMedia[mc.ID] = *mc
        return nil
    }
    return errors.New("Media was not found to update")
}

func (cm ContentManagerMemory) UpdateScreen(s *models.PreviewScreen) error {
    if _, ok := cm.ValidScreens[s.ID]; ok {
        cm.ValidScreens[s.ID] = *s
        return nil
    }
    return errors.New("Media was not found to update")
}

// Given the current parameters in the buffalo context return a list of matching containers.
func (cm ContentManagerMemory) ListContainersContext() (*models.Containers, error) {
    _, per_page, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListContainers(page, per_page)
}

// Actually list containers using a page and per_page which is consistent with buffalo standard pagination
func (cm ContentManagerMemory) ListContainers(page int, per_page int) (*models.Containers, error) {
    log.Printf("List Containers with page(%d) and per_page(%d)", page, per_page)

    c_arr := models.Containers{}
    for _, c := range cm.ValidContainers {
        c_arr = append(c_arr, c)
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

func (cm ContentManagerMemory) FindFileRef(mcID uuid.UUID) (*models.MediaContainer, error) {
    if mc, ok := cm.ValidMedia[mcID]; ok {
        return &mc, nil
    }
    return nil, errors.New("File was not found in the current list of files")
}

func (cm ContentManagerMemory) GetPreviewForMC(mc *models.MediaContainer) (string, error) {
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

func (cm ContentManagerMemory) FindActualFile(mc *models.MediaContainer) (string, error) {
    cnt, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        return "Memory Manager View no Parent Found", err
    }
    log.Printf("Memory Manager View %s loading up %s\n", mc.ID.String(), mc.Src)
    return utils.GetFilePathInContainer(mc.Src, cnt.GetFqPath())
}

// If you want to do in memory testing and already manually created previews this will
// then try and use the previews for the in memory manager.
func (cm ContentManagerMemory) SetPreviewIfExists(mc *models.MediaContainer) (string, error) {
    c, err := cm.GetContainer(mc.ContainerID.UUID)
    if err != nil {
        log.Fatal(err)
        return "", err
    }
    pFile := utils.AssignPreviewIfExists(c, mc)
    return pFile, nil
}

func (cm ContentManagerMemory) ListScreensContext(mcID uuid.UUID) (*models.PreviewScreens, error) {
    // Could add the context here correctly
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListScreens(mcID, page, limit)
}

// TODO: Get a pattern for each MC, look at a preview Destination, then match against the pattern
// And build out a set of screens.
func (cm ContentManagerMemory) ListScreens(mcID uuid.UUID, page int, per_page int) (*models.PreviewScreens, error) {

    // Did I create this just to sort by Idx across all media?  Kinda strange
    s_arr := models.PreviewScreens{}
    for _, s := range cm.ValidScreens {
        if s.MediaID == mcID {
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

func (cm ContentManagerMemory) ListAllScreensContext() (*models.PreviewScreens, error) {
    _, limit, page := GetPagination(cm.Params(), cm.cfg.Limit)
    return cm.ListAllScreens(page, limit)
}

func (cm ContentManagerMemory) ListAllScreens(page int, per_page int) (*models.PreviewScreens, error) {

    log.Printf("Using memory manager for screen page %d per_page %d \n", page, per_page)
    // Did I create this just to sort by Idx across all media?  Kinda strange
    s_arr := models.PreviewScreens{}
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

func (cm ContentManagerMemory) GetScreen(psID uuid.UUID) (*models.PreviewScreen, error) {
    // Need to build out a memory setup and look the damn thing up :(
    memStorage := utils.GetMemStorage()
    if screen, ok := memStorage.ValidScreens[psID]; ok {
        return &screen, nil
    }
    return nil, errors.New("Screen not found")
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
func (cm ContentManagerMemory) CreateScreen(screen *models.PreviewScreen) error {
    if screen != nil {
        screen.ID = AssignID(screen.ID)
        cm.ValidScreens[screen.ID] = *screen
        return nil
    }
    return errors.New("ContentManagerMemory no screen instance was passed in to CreateScreen")
}

func (cm ContentManagerMemory) CreateMedia(mc *models.MediaContainer) error {
    if mc != nil {
        mc.ID = AssignID(mc.ID)
        cm.ValidMedia[mc.ID] = *mc
        return nil
    }
    return errors.New("ContentManagerMemory no mediainstance was passed in to CreateMedia")
}

// Note that we need to lock this down so that it cannot just access arbitrary files
func (cm ContentManagerMemory) CreateContainer(c *models.Container) error {
    if c != nil {
        c.ID = AssignID(c.ID)
        cm.ValidContainers[c.ID] = *c
        return nil
    }
    return errors.New("ContentManagerMemory no container was passed in to CreateContainer")
}

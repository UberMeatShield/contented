/** 
 * Implements the ContentManager interface but stores all information in memory.  This
 * is using the utility/MemStorage singleton which will load up the disk information only
 * one time.
 */
package managers

import (
    "log"
    "errors"
    "sort"
    "regexp"
    "net/url"
    "contented/models"
    "contented/utils"
    "github.com/gofrs/uuid"
)

// Provides the support for looking up media by ID while only using memory
type ContentManagerMemory struct {
    cfg *utils.DirConfigEntry

    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    validate string 

    params *url.Values
    Params GetParamsType
}

// We do not allow editing in a memory manager
func (cm ContentManagerMemory) CanEdit() bool {
    return false;
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
    if end > 0 {  // If it is empty a slice ending in 0 = boom
        m_arr = m_arr[offset : end]
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
    return cm.SearchMedia(searchStr, page, per_page, cId)
}



func (cm ContentManagerMemory) SearchMedia(search string, page int, per_page int, containerID string) (*models.MediaContainers, int, error) {
    filteredMedia, cErr := cm.getMediaFiltered(containerID, search)
    if cErr != nil {
        return nil, 0, cErr
    }
    if filteredMedia == nil {
        empty := models.MediaContainers{}
        return &empty, 0, nil
    }

    // HMMMM
    mc_arr := *filteredMedia
    count := len(mc_arr)
    offset, end := GetOffsetEnd(page, per_page, count)
    if end > 0 {  // If it is empty a slice ending in 0 = boom
        mc_arr = mc_arr[offset : end]
        return &mc_arr, count, nil
    }
    return &mc_arr, count, nil
}


func (cm ContentManagerMemory) getMediaFiltered(containerID string, search string) (*models.MediaContainers, error){
    // First pass just grab all media in an array OR filter by containerID
    fp_arr := models.MediaContainers{}


    // If a containerID is specified and is totally invalid raise an error, otherwise filter
    if containerID != "" {
        cID, cErr := uuid.FromString(containerID)    
        if cErr == nil {
            for _, mc := range cm.ValidMedia {
                if mc.ContainerID.Valid && mc.ContainerID.UUID == cID {
                    fp_arr = append(fp_arr, mc)
                }
            }
        } else {
            return nil, cErr
        }
    } else {
        // Empty string for containerID is considered match all media
        for _, mc := range cm.ValidMedia {
            fp_arr = append(fp_arr, mc) 
        }
    }

    // Second pass we iterate over the sorted array
    mc_arr := models.MediaContainers{}
    if search != "" && search != "*" {
        // Could optimize by offset end but "eh, good enough for in memory"
        searcher := regexp.MustCompile(search)
        for _, mc := range fp_arr {
            if searcher.MatchString(mc.Src) {
                mc_arr = append(mc_arr, mc)
            }
        }
    } else {
        // If there is no search or the default args just return everything that matched
        mc_arr = fp_arr
    }

    // Finally sort any content that is matching so that pagination will work
    sort.SliceStable(mc_arr, func(i, j int) bool {
        return mc_arr[i].Idx < mc_arr[j].Idx
    })
    return &mc_arr, nil
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
    if end > 0 {  // If it is empty a slice ending in 0 = boom
        m_arr = m_arr[offset : end]
        return &m_arr, nil
    }
    log.Printf("Get a list of media offset(%d), end(%d) we should have some %d", offset, end, len(m_arr))
    return &m_arr, nil
}

// Get a media element by the ID
func (cm ContentManagerMemory) GetMedia(mcID uuid.UUID) (*models.MediaContainer, error) {
    log.Printf("Memory Get a single media %s", mcID)
    if mc, ok := cm.ValidMedia[mcID]; ok {
        return &mc, nil
    }
    return nil, errors.New("Media was not found in memory")
}


// If you already updated the container in memory you are done
func (cm ContentManagerMemory) UpdateContainer(c *models.Container) error {
    // TODO: Validate that this updates the actual reference in mem storage
    return nil
}

// No updates should be allowed for memory management.
func (cm ContentManagerMemory) UpdateMedia(media *models.MediaContainer) error {
    // TODO: Validate that this updates the actual reference in mem storage
    // TODO: Should this just be considered complete if you modified the media?
    return errors.New("Updates are not allowed for in memory management")
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
    c_arr = c_arr[offset : end]
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

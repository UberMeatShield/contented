package utils

/**
* This provides a single instance of the content tree that should be hosted when you
* choose just to use an in memory version of a directory.  It keeps a hash lookup for
* containers and media and is then used by the MemoryManager.
 */
import (
    "contented/models"
    "log"
)

// GoLang is just making this awkward
type MemoryStorage struct {
    Initialized     bool
    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
    ValidScreens    models.ScreenMap
    ValidTags       models.TagsMap
}

var memStorage MemoryStorage = MemoryStorage{Initialized: false}

func GetMemStorage() *MemoryStorage {
    return &memStorage
}

func InitializeMemory(dir_root string) *MemoryStorage {
    log.Printf("Initializing Memory Storage %s\n", dir_root)
    containers, files, screens := PopulateMemoryView(dir_root)

    memStorage.Initialized = true
    memStorage.ValidContainers = containers
    memStorage.ValidMedia = files
    memStorage.ValidScreens = screens
    memStorage.ValidTags = models.TagsMap{}

    return &memStorage
}

/**
 * Populates the memory view (this code is very similar to the DB version in helper.go)
 */
func PopulateMemoryView(dir_root string) (models.ContainerMap, models.MediaMap, models.ScreenMap) {
    containers := models.ContainerMap{}
    files := models.MediaMap{}
    screensMap := models.ScreenMap{}

    cfg := GetCfg()
    log.Printf("PopulateMemoryView searching in %s with depth %d", dir_root, cfg.MaxSearchDepth)
    contentTree, err := CreateStructure(cfg.Dir, cfg, &ContentTree{}, 0)
    if err != nil {
        log.Fatalf("Failed to create the intial in memory structure %s", err)
    }

    tree := *contentTree
    for idx, ct := range tree {
        if cfg.ExcludeEmptyContainers && len(ct.Media) == 0 {
            continue // SKIP empty container directories
        }

        // Careful as sometimes we do want containers even if there is no media
        c := ct.Cnt
        if len(ct.Media) > 0 {
            c.PreviewUrl = "/preview/" + ct.Media[0].ID.String()
            log.Printf("Assigning a preview to %s as %s", c.Name, c.PreviewUrl)

            maybeScreens, screenErr := GetPotentialScreens(&c)
            for _, mc := range ct.Media {
                // Assign anything required to the media before we put it in the lookup hash
                AssignPreviewIfExists(&c, &mc)
                if screenErr == nil {
                    screens := AssignScreensFromSet(&c, &mc, maybeScreens)
                    if screens != nil {
                        for _, screen := range *screens {
                            screensMap[screen.ID] = screen
                        }
                    }
                } else {
                    log.Printf("No potential screens present in container %s", c.Path)
                }
                files[mc.ID] = mc
            }
        }
        // Remember that assigning into a map is also a copy so any changes must be
        // done BEFORE you assign into the map
        c.Idx = idx
        containers[c.ID] = c
    }
    return containers, files, screensMap
}

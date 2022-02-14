package utils

import (
    "log"
    "contented/models"
)
// GoLang is just making this awkward
type MemoryStorage struct {
    Initialized bool
    ValidMedia      models.MediaMap
    ValidContainers models.ContainerMap
}
var memStorage MemoryStorage = MemoryStorage{Initialized: false}

func GetMemStorage() *MemoryStorage {
   return &memStorage 
}

func InitializeMemory(dir_root string) *MemoryStorage {
    log.Printf("Initializing Memory Storage %s\n", dir_root)
    containers, files := PopulateMemoryView(dir_root)

    memStorage.Initialized = true
    memStorage.ValidContainers = containers
    memStorage.ValidMedia = files

    return &memStorage
}

/**
 * Populates the memory view (this code is very similar to the DB version in helper.go)
 */
func PopulateMemoryView(dir_root string) (models.ContainerMap, models.MediaMap) {
    containers := models.ContainerMap{}
    files := models.MediaMap{}

    cfg := GetCfg()
    log.Printf("PopulateMemoryView searching in %s with depth %d", dir_root, cfg.MaxSearchDepth)
    contentTree, err := CreateStructure(cfg.Dir, cfg, &ContentTree{}, 0)
    if err != nil {
        log.Fatalf("Failed to create the intial in memory structure %s", err)
    }

    tree := *contentTree
    for idx, ct := range tree {
        if cfg.ExcludeEmptyContainers && len(ct.Media) == 0 {
            continue  // SKIP empty container directories
        }

        // Ensure that we do grab all the containers
        c := ct.Cnt
        c.Idx = idx
        containers[c.ID] = c

        // Careful as sometimes we do want containers even if there is no media
        if len(ct.Media) > 0 {
            c.PreviewUrl = "/preview/" + ct.Media[0].ID.String()
            for _, mc := range ct.Media {
                AssignPreviewIfExists(&c, &mc)
                files[mc.ID] = mc
            }
        }
    }
    return containers, files
}

func AssignPreviewIfExists(c* models.Container, mc *models.MediaContainer) string {
    // This check is normally to determine if we didn't clear out old previews.
    // For memory only managers it will just consider that a bonus and use the preview.
    previewPath := GetPreviewDst(c.GetFqPath())
    previewFile, exists := ErrorOnPreviewExists(mc.Src, previewPath, mc.ContentType)
    if exists != nil {
        mc.Preview = GetRelativePreviewPath(previewFile, c.GetFqPath())
        log.Printf("Added a preview to media %s", mc.Preview)
    }
    return previewFile
}

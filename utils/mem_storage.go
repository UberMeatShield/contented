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
 * Populates the memory view
 */
func PopulateMemoryView(dir_root string) (models.ContainerMap, models.MediaMap) {
    containers := models.ContainerMap{}
    files := models.MediaMap{}

    log.Printf("Searching in %s", dir_root)
    cfg := GetCfg()

    cnts := FindContainers(dir_root)
    for idx, c := range cnts {
        media := FindMediaMatcher(c, 90001, 0, cfg.IncFiles, cfg.ExcFiles)  // TODO: Config this
        // c.Contents = media
        c.Total = len(media)
        c.Idx = idx
        containers[c.ID] = c

        // This check is normally to determine if we didn't clear out old previews.
        // For memory only managers it will just consider that a bonus and use the preview.
        for _, mc := range media {
            AssignPreviewIfExists(&c, &mc)
            files[mc.ID] = mc
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

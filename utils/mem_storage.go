package utils

/**
* This provides a single instance of the content tree that should be hosted when you
* choose just to use an in memory version of a directory.  It keeps a hash lookup for
* containers and media and is then used by the MemoryManager.
 */
import (
	"contented/models"
    "strings"
	"log"
    "io/ioutil"
    "github.com/gofrs/uuid"
)

// GoLang is just making this awkward
type MemoryStorage struct {
	Initialized     bool
	ValidMedia      models.MediaMap
	ValidContainers models.ContainerMap
	ValidScreens    models.PreviewScreenMap
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
			continue // SKIP empty container directories
		}

		// Careful as sometimes we do want containers even if there is no media
		c := ct.Cnt
		if len(ct.Media) > 0 {
			c.PreviewUrl = "/preview/" + ct.Media[0].ID.String()
			log.Printf("Assigning a preview to %s as %s", c.Name, c.PreviewUrl)
			for _, mc := range ct.Media {
				// I should name this as PreviewUrl
				AssignPreviewIfExists(&c, &mc)
				files[mc.ID] = mc
			}
		}
		// Remember that assigning into a map is also a copy so any changes must be
		// done BEFORE you assign into the map
		c.Idx = idx
		containers[c.ID] = c
	}
	return containers, files
}


func AssignScreensIfExists(c *models.Container, mc *models.MediaContainer) (*models.PreviewScreens) {
    if !strings.Contains(mc.ContentType, "video") {
        log.Printf("Media is not of type video, no screens likely")
        return nil
    }

	previewPath := GetPreviewDst(c.GetFqPath())
    maybeScreens, err := ioutil.ReadDir(previewPath)
    if err != nil {
        return nil
    }
    previewScreens := models.PreviewScreens{}
    screenRe := GetScreensMatcherRE(mc.Src)
    for idx, fRef := range maybeScreens {
        if !fRef.IsDir() {
            name := fRef.Name()
            if screenRe.MatchString(name) {
                log.Printf("Matched file %s idx %d", name, idx) 
                id, _ := uuid.NewV4()
                ps := models.PreviewScreen{
                    ID: id,
                    Path: previewPath,
                    Src: name,
                    MediaID: mc.ID,
                    Idx: idx,
                    SizeBytes: fRef.Size(),
                }
                previewScreens = append(previewScreens, ps)
            } else {
                log.Printf("Did not match %s", name)
            }
        }
    }
    log.Printf("Looking through a container preview directory %s", screenRe)
    return &previewScreens
}


func AssignPreviewIfExists(c *models.Container, mc *models.MediaContainer) string {
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

package utils

/**
* This provides a single instance of the content tree that should be hosted when you
* choose just to use an in memory version of a directory.  It keeps a hash lookup for
* containers and content and is then used by the MemoryManager.
 */
import (
	"contented/models"
	"errors"
	"log"
	"time"
)

// GoLang is just making this awkward
type MemoryStorage struct {
	Initialized     bool
	ValidContent    models.ContentMap
	ValidContainers models.ContainerMap
	ValidScreens    models.ScreenMap
	ValidTags       models.TagsMap
	ValidTasks      models.TaskRequests // Not a Map as we want the order to matter
}

var memStorage MemoryStorage = MemoryStorage{Initialized: false}

func GetMemStorage() *MemoryStorage {
	return &memStorage
}

func InitializeMemory(dir_root string) *MemoryStorage {
	log.Printf("Initializing Memory Storage %s\n", dir_root)
	containers, files, screens, tags := PopulateMemoryView(dir_root)

	memStorage.Initialized = true
	memStorage.ValidContainers = containers
	memStorage.ValidContent = files
	memStorage.ValidScreens = screens
	memStorage.ValidTags = tags
	memStorage.ValidTasks = models.TaskRequests{}
	return &memStorage
}

func InitializeEmptyMemory() *MemoryStorage {
	memStorage.Initialized = true
	memStorage.ValidContainers = models.ContainerMap{}
	memStorage.ValidContent = models.ContentMap{}
	memStorage.ValidScreens = models.ScreenMap{}
	memStorage.ValidTags = models.TagsMap{}
	memStorage.ValidTasks = models.TaskRequests{}
	return &memStorage
}

func (ms MemoryStorage) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequests, error) {
	updated := false
	for idx, task := range ms.ValidTasks {
		// Check to ensure the state is known before the updated which should
		// prevent MOST update errors in the memory view.
		if task.ID == t.ID && currentState == task.Status {
			t.UpdatedAt = time.Now()
			ms.ValidTasks[idx] = *t
			updated = true
			break
		}
	}
	if updated == false {
		return nil, errors.New("Could not find Task to update")
	}
	return &ms.ValidTasks, nil
}

/**
 * Populates the memory view (this code is very similar to the DB version in helper.go)
 */
func PopulateMemoryView(dir_root string) (models.ContainerMap, models.ContentMap, models.ScreenMap, models.TagsMap) {
	containers := models.ContainerMap{}
	files := models.ContentMap{}
	screensMap := models.ScreenMap{}

	cfg := GetCfg()

	log.Printf("PopulateMemoryView searching in %s with depth %d", dir_root, cfg.MaxSearchDepth)
	contentTree, err := CreateStructure(cfg.Dir, cfg, &ContentTree{}, 0)
	if err != nil {
		log.Fatalf("Failed to create the intial in memory structure %s", err)
	}

	tree := *contentTree
	for idx, ct := range tree {
		if cfg.ExcludeEmptyContainers && len(ct.Content) == 0 {
			continue // SKIP empty container directories
		}

		// Careful as sometimes we do want containers even if there is no content
		c := ct.Cnt
		if len(ct.Content) > 0 {
			c.PreviewUrl = "/preview/" + ct.Content[0].ID.String()
			log.Printf("Assigning a preview to %s as %s", c.Name, c.PreviewUrl)

			maybeScreens, screenErr := GetPotentialScreens(&c)
			for _, mc := range ct.Content {
				// Assign anything required to the content before we put it in the lookup hash
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

	// log.Printf("LOADING TAGS %s", cfg.TagFile)
	// Only will work if TAG_FILE is actually set to something
	tags, tErr := ReadTagsFromFile(cfg.TagFile)
	tagsMap := models.TagsMap{}
	if tErr == nil && tags != nil {
		for _, tag := range *tags {
			tagsMap[tag.ID] = tag
		}
	}
	return containers, files, screensMap, tagsMap
}

package utils

/**
* This provides a single instance of the content tree that should be hosted when you
* choose just to use an in memory version of a directory.  It keeps a hash lookup for
* containers and content and is then used by the MemoryManager.
 */
import (
	"contented/models"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

// GoLang is just making this awkward
type MemoryStorage struct {
	Initialized     bool
	Loading         bool
	ValidContent    models.ContentMap
	ValidContainers models.ContainerMap
	ValidScreens    models.ScreenMap
	ValidTags       models.TagsMap
	ValidTasks      models.TaskRequests // Not a Map as we want the order to matter
}

var memStorage MemoryStorage = MemoryStorage{Initialized: false, Loading: false}

func GetMemStorage() *MemoryStorage {
	return &memStorage
}

/**
 * Sketchy initialization based on test or non-test mode.
 */
func InitializeMemory(dirRoot string) *MemoryStorage {
	log.Printf("Initializing Memory Storage %s\n", dirRoot)
	if memStorage.Loading && !testing.Testing() {
		log.Printf("Still loading up memory storage")
		return &memStorage
	}
	memStorage.Loading = true
	containers, files, screens, tags := PopulateMemoryView(dirRoot)

	memStorage.ValidContainers = containers
	memStorage.ValidContent = files
	memStorage.ValidScreens = screens
	memStorage.ValidTags = tags
	memStorage.ValidTasks = models.TaskRequests{}

	memStorage.Initialized = true
	memStorage.Loading = false
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

func AssignID(id uuid.UUID) uuid.UUID {
	emptyID, _ := uuid.FromString("00000000-0000-0000-0000-000000000000")
	if id == emptyID {
		newID, _ := uuid.NewV4()
		return newID
	}
	return id
}

func (ms MemoryStorage) CreateScreen(screen *models.Screen) (*models.Screen, error) {
	screen.ID = AssignID(screen.ID)
	screen.CreatedAt = time.Now()
	screen.UpdatedAt = time.Now()
	memStorage.ValidScreens[screen.ID] = *screen
	return screen, nil
}

func (ms MemoryStorage) UpdateScreen(s *models.Screen) (*models.Screen, error) {
	if _, ok := memStorage.ValidScreens[s.ID]; ok {
		s.UpdatedAt = time.Now()
		memStorage.ValidScreens[s.ID] = *s
		return s, nil
	}
	return nil, errors.New(fmt.Sprintf("Screen not found with %s", s))
}

func (ms MemoryStorage) CreateContent(content *models.Content) (*models.Content, error) {
	content.ID = AssignID(content.ID)
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()
	memStorage.ValidContent[content.ID] = *content
	return content, nil
}

// Reload the contents with the ID?
func (ms MemoryStorage) UpdateContent(content *models.Content) (*models.Content, error) {
	if _, ok := memStorage.ValidContent[content.ID]; ok {
		content.UpdatedAt = time.Now()
		memStorage.ValidContent[content.ID] = *content
		return content, nil
	}
	return nil, errors.New("Content was not found")
}

func (ms MemoryStorage) CreateContainer(c *models.Container) (*models.Container, error) {
	c.ID = AssignID(c.ID)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	memStorage.ValidContainers[c.ID] = *c
	return c, nil
}

func (ms MemoryStorage) UpdateContainer(cnt *models.Container) (*models.Container, error) {
	if _, ok := memStorage.ValidContainers[cnt.ID]; ok {
		cnt.UpdatedAt = time.Now()
		memStorage.ValidContainers[cnt.ID] = *cnt
		return cnt, nil
	}
	return nil, errors.New(fmt.Sprintf("Container was not found with ID %s", cnt))
}

func (ms MemoryStorage) CreateTask(tr *models.TaskRequest) (*models.TaskRequest, error) {
	tr.ID = AssignID(tr.ID)
	tr.CreatedAt = time.Now()
	tr.UpdatedAt = time.Now()
	tr.Status = models.TaskStatus.NEW
	memStorage.ValidTasks = append(memStorage.ValidTasks, *tr)
	return tr, nil
}

func (ms MemoryStorage) CreateTag(tag *models.Tag) (*models.Tag, error) {
	tag.UpdatedAt = time.Now()
	memStorage.ValidTags[tag.ID] = *tag
	return tag, nil
}

func (ms MemoryStorage) UpdateTag(tag *models.Tag) (*models.Tag, error) {
	if _, ok := memStorage.ValidTags[tag.ID]; ok {
		tag.UpdatedAt = time.Now()
		memStorage.ValidTags[tag.ID] = *tag
		return tag, nil
	}
	return nil, errors.New(fmt.Sprintf("Tag was not found %s", tag))
}

func (ms MemoryStorage) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequest, error) {
	updated := false
	for idx, task := range memStorage.ValidTasks {
		// Check to ensure the state is known before the updated which should
		// prevent MOST update errors in the memory view.
		log.Printf("Looking at %s trying to find id(%s) in state %s", task, t.ID.String(), currentState)
		if task.ID == t.ID && (currentState == task.Status || task.Status == t.Status) {
			t.UpdatedAt = time.Now()
			memStorage.ValidTasks[idx] = *t
			updated = true
			break
		}
	}
	if updated == false {
		return nil, errors.New("Could not find Task to update")
	}
	return t, nil
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

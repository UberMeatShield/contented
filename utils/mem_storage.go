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

	"golang.org/x/exp/maps"
)

type SequenceMap map[string]int64

// GoLang is just making this awkward
type MemoryStorage struct {
	Initialized     bool
	Loading         bool
	ValidContent    models.ContentMap
	ValidContainers models.ContainerMap
	ValidScreens    models.ScreenMap
	ValidTags       models.TagsMap
	ValidTasks      models.TaskRequests // Not a Map as we want the order to matter
	Sequences       SequenceMap
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
	memStorage.Sequences = SequenceMap{"screens": 0, "contents": 0, "containers": 0, "taskrequests": 0}
	memStorage.Loading = true
	containers, contents, screens, tags := PopulateMemoryView(dirRoot)

	// Should I tag the container?
	if contents != nil && tags != nil {
		log.Printf("Attempting to assign tags to content")
		taggedContent := AssignTagsToContent(contents, tags)

		// Update the tagged content
		for _, content := range maps.Values(taggedContent) {
			contents[content.ID] = content
		}
	}

	memStorage.ValidContainers = containers
	memStorage.ValidContent = contents
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

// How is it that GoLang doesn't have a more sensible default fallback?
func StringDefault(s1 string, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
}

func AssignNumerical(id int64, tablename string) int64 {
	// TODO: If This is using DB potentially just return 0
	memStorage.Sequences[tablename] += 1
	return memStorage.Sequences[tablename]
}

func (ms MemoryStorage) CreateScreen(screen *models.Screen) (*models.Screen, error) {
	screen.ID = AssignNumerical(screen.ID, "screens")
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
	content.ID = AssignNumerical(content.ID, "contents")
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
	c.ID = AssignNumerical(c.ID, "containers")
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
	return nil, fmt.Errorf("container was not found with ID %s", cnt)
}

func (ms MemoryStorage) CreateTask(tr *models.TaskRequest) (*models.TaskRequest, error) {
	tr.ID = AssignNumerical(tr.ID, "taskrequests")
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
	return nil, fmt.Errorf("tag was not found %s", tag)
}

func (ms MemoryStorage) UpdateTask(t *models.TaskRequest, currentState models.TaskStatusType) (*models.TaskRequest, error) {
	updated := false
	for idx, task := range memStorage.ValidTasks {
		// Check to ensure the state is known before the updated which should
		// prevent MOST update errors in the memory view.
		log.Printf("Looking at %s trying to find id(%d) in state %s", task, t.ID, currentState)
		if task.ID == t.ID && (currentState == task.Status || task.Status == t.Status) {
			t.UpdatedAt = time.Now()
			memStorage.ValidTasks[idx] = *t
			updated = true
			break
		}
	}
	log.Printf("Updated the task %d", t.ID)
	if !updated {
		return nil, errors.New("could not find Task to update")
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
			c.PreviewUrl = fmt.Sprintf("/api/preview/%d", ct.Content[0].ID)
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

package actions

import (
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/stretchr/testify/assert"
)

func CreateVideoContainer(t *testing.T, router *gin.Engine) (*models.Container, *models.Content) {
	cnt, contents := CreateVideoContents("dir2", "donut", t, router)
	assert.NotNil(t, contents, "No donut video found in dir2!")
	ref := *contents
	return cnt, &ref[0]
}

func IsVideoMatch(content models.Content, contentMatch string) bool {
	if !strings.Contains(content.ContentType, "video") {
		return false
	}
	if contentMatch == "" || strings.Contains(content.Src, contentMatch) {
		return true
	}
	return false
}

// Make something that does this in a cleaner fashion
func CreateVideoContents(containerName string, contentMatch string, t *testing.T, router *gin.Engine) (*models.Container, *models.Contents) {
	cntToCreate, contents := test_common.GetContentByDirName(containerName)

	cnt := CreateContainer(cntToCreate, t, router)
	assert.NotZero(t, cnt.ID, "It should create a valid container")

	contentsCreated := models.Contents{}
	for _, contentToCreate := range contents {

		if IsVideoMatch(contentToCreate, contentMatch) {
			contentToCreate.ContainerID = &cnt.ID

			content := CreateContent(&contentToCreate, t, router)
			assert.NotZero(t, content.ID, fmt.Sprintf("It should have created content %s", content))
			contentsCreated = append(contentsCreated, content)
		}
	}
	return &cnt, &contentsCreated
}

func TestTaskRelatedObjects(t *testing.T) {
	assert.Equal(t, models.TaskOperation.SCREENS.String(), "screen_capture")
	assert.Equal(t, models.TaskOperation.ENCODING.String(), "video_encoding")
	assert.Equal(t, models.TaskOperation.WEBP.String(), "webp_from_screens")
	assert.Equal(t, models.TaskOperation.TAGGING.String(), "tag_content")
	assert.Equal(t, models.TaskOperation.DUPES.String(), "detect_duplicates")
}

// Do the screen grab in memory
func TestEditingQueueScreenHandlerMemory(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(false)
	test_common.InitMemoryFakeAppEmpty()

	assert.Equal(t, cfg.ReadOnly, false)
	ValidateEditingQueue(t, router)
}

func TestDbEditingQueueScreenHandlerDB(t *testing.T) {
	_, db, router := InitFakeRouterApp(true)
	assert.NotNil(t, db, "The db should be initialized")
	ValidateEditingQueue(t, router)
}

func ValidateEditingQueue(t *testing.T, router *gin.Engine) {
	_, content := CreateVideoContainer(t, router)
	timeSeconds := 3
	screenCount := 1
	url := fmt.Sprintf("/api/editing_queue/%d/screens/%d/%d", content.ID, screenCount, timeSeconds)
	tr := models.TaskRequest{}
	code, err := PostJson(url, content, &tr, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Should be able to grab a screen %s", err))

	// Huh... ODD
	assert.NotZero(t, tr.ID, fmt.Sprintf("Did not create a Task %s", tr))
	assert.Equal(t, models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	assert.Equal(t, tr.Operation, models.TaskOperation.SCREENS)

	args := worker.Args{"id": strconv.FormatInt(tr.ID, 10)}
	errWrapper := ScreenCaptureWrapper(args)
	assert.NoError(t, errWrapper, fmt.Sprintf("Failed to get screens %s", err))

	screenUrl := fmt.Sprintf("/api/contents/%d/screens", content.ID)
	sRes := ScreensResponse{}
	sCode, sErr := GetJson(screenUrl, "", &sRes, router)
	assert.Equal(t, http.StatusOK, sCode, fmt.Sprintf("Error loading screens %s", sErr))

	assert.Equal(t, screenCount, len(sRes.Results), fmt.Sprintf("We should have a set number of screens %s", sRes.Results))
	assert.Equal(t, int64(screenCount), sRes.Total, "The count should be correct")

	// Validate the task is now done
	checkTask := models.TaskRequest{}
	taskUrl := fmt.Sprintf("/api/task_requests/%d", tr.ID)
	tCode, tErr := GetJson(taskUrl, "", &checkTask, router)
	assert.Equal(t, http.StatusOK, tCode, fmt.Sprintf("Failed with %s", tErr))
	assert.Equal(t, checkTask.Status, models.TaskStatus.DONE, fmt.Sprintf("It should be done %s", checkTask))
}

func TestMemoryEncodingQueueHandlerMemory(t *testing.T) {
	// Should add a config value to completely nuke the encoded video
	_, _, router := InitFakeRouterApp(false)
	test_common.InitMemoryFakeAppEmpty()
	ValidateVideoEncodingQueue(t, router)
}

func TestEncodingQueueHandlerDB(t *testing.T) {
	// Should add a config value to completely nuke the encoded video
	_, _, router := InitFakeRouterApp(true)
	ValidateVideoEncodingQueue(t, router)
}

func ValidateVideoEncodingQueue(t *testing.T, router *gin.Engine) {
	cnt, content := CreateVideoContainer(t, router)
	url := fmt.Sprintf("/api/editing_queue/%d/encoding", content.ID)

	tr := models.TaskRequest{}
	code, err := PostJson(url, content, &tr, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Failed to queue encoding task %s", err))

	assert.NotZero(t, tr.ID, fmt.Sprintf("Did not create a Task %s", tr))
	assert.Equal(t, models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	assert.Equal(t, tr.Operation, models.TaskOperation.ENCODING)

	args := worker.Args{"id": strconv.FormatInt(tr.ID, 10)}
	vErr := VideoEncodingWrapper(args)
	assert.NoError(t, vErr, fmt.Sprintf("Failed to encode video %s", vErr))

	checkTask := models.TaskRequest{}
	rCode, rErr := GetJson(fmt.Sprintf("/api/task_requests/%d", tr.ID), "", &checkTask, router)
	assert.Equal(t, http.StatusOK, rCode, fmt.Sprintf("Failed to get the task request %s", rErr))
	assert.Equal(t, checkTask.Status, models.TaskStatus.DONE, fmt.Sprintf("It should be done %s", checkTask))

	assert.NotNil(t, checkTask.CreatedID, fmt.Sprintf("there sould be an ID created %s", checkTask))
	createdID := *checkTask.CreatedID
	assert.NotZero(t, createdID, "It should create a new piece of content")
	checkContent := models.Content{}
	checkCode, checkErr := GetJson(fmt.Sprintf("/api/contents/%d", createdID), "", &checkContent, router)
	assert.Equal(t, http.StatusOK, checkCode, fmt.Sprintf("Error loading %s", checkErr))

	assert.Equal(t, *checkContent.ContainerID, cnt.ID, "The container ID did not match")
	assert.Contains(t, checkContent.Src, "h265", "The encoded container was not valid")

	// The container path is hidden in the API so load the actual DB element
	// TODO: DO NOT CHECK IN, I really need a faster encoding file.
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	cntActual, pathErr := man.GetContainer(cnt.ID)
	assert.NoError(t, pathErr)

	dstFile := filepath.Join(cntActual.GetFqPath(), checkContent.Src)
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		os.Remove(dstFile)
	} else {
		assert.Fail(t, "It did NOT remove the destination file %s", dstFile)
	}
}

func TestWebpHandlerDB(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(true)
	test_common.Get_VideoAndSetupPaths(cfg) // Resets the screens

	cfg.ScreensOverSize = 1
	cfg.PreviewVideoType = "screens"
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(t, router)
	ValidateWebpCode(t, router, content)
	// Create some screens and then encode it?
}

func TestWebpHandlerMemory(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(false)
	test_common.Get_VideoAndSetupPaths(cfg) // Resets the screens

	cfg.ScreensOverSize = 1
	cfg.PreviewVideoType = "screens"
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(t, router)
	ValidateWebpCode(t, router, content)
}

func ValidateWebpCode(t *testing.T, router *gin.Engine, content *models.Content) {
	assert.Equal(t, content.Preview, "", "It should not have a preview already")
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)
	_, screenErr, _ := managers.CreateScreensForContent(man, content.ID, 10, 1)
	assert.NoError(t, screenErr)

	url := fmt.Sprintf("/api/editing_queue/%d/webp", content.ID)
	tr := models.TaskRequest{}
	code, err := PostJson(url, content, &tr, router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Failed to queue encoding task %s", err))

	assert.NotZero(t, tr.ID, fmt.Sprintf("WebP did not create a Task %s", tr))
	assert.Equal(t, models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	assert.Equal(t, tr.Operation, models.TaskOperation.WEBP)

	args := worker.Args{"id": strconv.FormatInt(tr.ID, 10)}
	wErr := WebpFromScreensWrapper(args)
	assert.NoError(t, wErr, fmt.Sprintf("Failed to create webp for task %s", wErr))

	checkContent := models.Content{}
	cCode, cErr := GetJson(fmt.Sprintf("/api/contents/%d", content.ID), "", &checkContent, router)
	assert.Equal(t, http.StatusOK, cCode, fmt.Sprintf("Error loading %s", cErr))
	// Get the content, check for a preview
	assert.Equal(t, "/container_previews/donut_[special( gunk.mp4.webp", checkContent.Preview)
}

func TestTagHandlerDB(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(true)
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(t, router)
	ValidateTaggingCode(t, router, content)
}

func TestTagHandlerMemory(t *testing.T) {
	cfg, _, router := InitFakeRouterApp(false)
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(t, router)
	ValidateTaggingCode(t, router, content)
}

func ValidateTaggingCode(t *testing.T, router *gin.Engine, content *models.Content) {
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	tag1 := models.Tag{ID: "donut"}
	tag2 := models.Tag{ID: "mp4"}
	badTag := models.Tag{ID: "THIS_WILL_NOT_MATCH"}
	man.CreateTag(&tag1)
	man.CreateTag(&tag2)
	man.CreateTag(&badTag)

	tr := models.TaskRequest{Operation: models.TaskOperation.TAGGING, ContentID: &content.ID}
	task, err := man.CreateTask(&tr)
	assert.NoError(t, err, "Failed to create Task to do tagging")
	assert.NotZero(t, task.ID)

	tagErr := managers.TaggingContentTask(man, task.ID)
	assert.NoError(t, tagErr, "It should not have a problem doing the tagging")

	contentTagged, errLoad := man.GetContent(content.ID)
	assert.NoError(t, errLoad, "It should be able to get the content back")

	assert.Equal(t, 2, len(contentTagged.Tags), fmt.Sprintf("There should be tags now %s", contentTagged))

	taskCheck, taskErr := man.GetTask(task.ID)
	assert.NoError(t, taskErr, "We should still have a task")
	assert.Equal(t, models.TaskStatus.DONE, taskCheck.Status)

	url := fmt.Sprintf("/api/editing_queue/%d/tagging", content.ID)
	checkTask := models.TaskRequest{}
	checkCode, checkErr := PostJson(url, &content, &checkTask, router)
	assert.Equal(t, http.StatusCreated, checkCode, fmt.Sprintf("Failed to queue tagging task %s", checkErr))

}

func TestDuplicateHandlerDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)

	// Create the directory with the duplicate test
	cnt, contents := CreateVideoContents("test_encoding", "", t, router)
	assert.NotNil(t, cnt)
	assert.NotNil(t, contents)

	ValidateDuplicatesTask(t, router, cnt)
	ValidateDuplicateApiCalls(t, router, cnt)
}

func TestDuplicateHandlerMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)

	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	query := managers.ContainerQuery{Name: "test_encoding", PerPage: 1}
	containers, total, err := man.SearchContainers(query)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total, "There should only be one container matching")
	assert.Equal(t, 1, len(*containers), "There should be one container")

	// Create the directory with the duplicate test
	cnt := (*containers)[0]
	assert.NotNil(t, cnt)
	ValidateDuplicatesTask(t, router, &cnt)
	ValidateDuplicateApiCalls(t, router, &cnt)
}

func ValidateDuplicatesTask(t *testing.T, router *gin.Engine, container *models.Container) {
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	tr := models.TaskRequest{Operation: models.TaskOperation.DUPES, ContainerID: &container.ID}
	task, err := man.CreateTask(&tr)
	assert.NoError(t, err, "It should be able to create the duplicates task")

	dupeErr := managers.DetectDuplicatesTask(man, task.ID)
	assert.NoError(t, dupeErr, "It should be able to run the duplicates task")

	taskCheck, errCheck := man.GetTask(task.ID)
	assert.NoError(t, errCheck)
	assert.Equal(t, taskCheck.Status, models.TaskStatus.DONE)
	assert.NotEqual(t, taskCheck.Message, "")

	dupes := managers.DuplicateContents{}
	json.Unmarshal([]byte(taskCheck.Message), &dupes)
	assert.Equal(t, 1, len(dupes), fmt.Sprintf("There should be a duplicate %s", dupes))
	assert.Equal(t, dupes[0].DuplicateSrc, "SampleVideo_1280x720_1mb.mp4")
}

func ValidateDuplicateApiCalls(t *testing.T, router *gin.Engine, container *models.Container) {
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	url := fmt.Sprintf("/api/editing_container_queue/%d/duplicates", container.ID)
	code, err := PostJson(url, container, "", router)
	assert.Equal(t, http.StatusCreated, code, fmt.Sprintf("Failed to queue dupe container task %s", err))

	cq := managers.ContentQuery{Search: "SampleVideo_1280x720_1mb_h265.mp4"}
	contents, total, contentErr := man.SearchContent(cq)
	assert.Equal(t, int64(1), total, fmt.Sprintf("It should have found the content %s", contents))
	assert.NoError(t, contentErr)

	mc := (*contents)[0]
	urlContent := fmt.Sprintf("/api/editing_queue/%d/duplicates", mc.ID)
	checkCode, checkErr := PostJson(urlContent, mc, models.TaskRequest{}, router)
	assert.Equal(t, http.StatusCreated, checkCode, fmt.Sprintf("Failed to queue dupe content task %s", checkErr))
}

/*
func (as *ActionSuite) Test_ContainerEncodingMemory() {
	cfg := test_common.InitFakeApp(false)
	utils.SetCfg(*cfg)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	ValidateContainerEncoding(as, man)
}

func (as *ActionSuite) Test_ContainerEncodingDB() {
	cfg := test_common.InitFakeApp(true)
	utils.SetCfg(*cfg)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	cnt, contents := CreateVideoContents(as, "test_encoding", "")
	assert.NotNil(t, cnt)
	assert.NotNil(t, contents)

	ValidateContainerEncoding(as, man)
}

func ValidateContainerEncoding(as *ActionSuite, man managers.ContentManager) {
	_, total, tErr := man.ListTasks(managers.TaskQuery{})
	assert.NoError(t, tErr)
	assert.Equal(t, 0, total, fmt.Sprintf("There should not be any tasks %d", total))

	query := managers.ContainerQuery{Name: "test_encoding", PerPage: 1}
	containers, total, err := man.SearchContainers(query)
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "There should only be one container matching")
	assert.Equal(t, 1, len(*containers), "There should be one container")

	// Create the directory with the duplicate test
	cnt := (*containers)[0]

	url := fmt.Sprintf("/editing_container_queue/%s/encoding", cnt.ID.String())
	res := as.JSON(url).Post(cnt)

	assert.Equal(t, http.StatusCreated, res.Code)

	// There are two video files so we want to try and test both
	tasks, _, tErr := man.ListTasks(managers.TaskQuery{})
	assert.NoError(t, tErr, "Tasks should exist and not error")
	assert.NotNil(t, tasks, "We should have a task result")
	assert.Equal(t, 2, len(*tasks), "There should be some tasks now defined")

	for _, task := range *tasks {
		assert.Equal(t, task.Operation, models.TaskOperation.ENCODING)
	}
}

func (as *ActionSuite) Test_ContainerScreensMemory() {
	cfg := test_common.InitFakeApp(false)
	utils.SetCfg(*cfg)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	ValidateContainerEncoding(as, man)
}

func (as *ActionSuite) Test_ContainerScreensDB() {
	cfg := test_common.InitFakeApp(true)
	utils.SetCfg(*cfg)

	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	cnt, contents := CreateVideoContents(as, "test_encoding", "")
	assert.NotNil(t, cnt)
	assert.NotNil(t, contents)

	ValidateContainerEncoding(as, man)
}

func ValidateContainerScreens(as *ActionSuite, man managers.ContentManager) {
	_, total, tErr := man.ListTasks(managers.TaskQuery{})
	assert.NoError(t, tErr)
	assert.Equal(t, 0, total, fmt.Sprintf("There should not be any tasks %d", total))

	query := managers.ContainerQuery{Name: "test_encoding", PerPage: 1}
	containers, total, err := man.SearchContainers(query)
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "There should only be one container matching")
	assert.Equal(t, 1, len(*containers), "There should be one container")

	// Create the directory with the duplicate test
	cnt := (*containers)[0]

	url := fmt.Sprintf("/editing_container_queue/%s/screens", cnt.ID.String())
	res := as.JSON(url).Post(cnt)

	assert.Equal(t, http.StatusCreated, res.Code)

	// There are two video files so we want to try and test both
	tasks, _, tErr := man.ListTasks(managers.TaskQuery{})
	assert.NoError(t, tErr, "Tasks should exist and not error")
	assert.NotNil(t, tasks, "We should have a task result")
	assert.Equal(t, 2, len(*tasks), "There should be some tasks now defined")
	for _, task := range *tasks {
		assert.Equal(t, task.Operation, models.TaskOperation.SCREENS)
	}
}

*/

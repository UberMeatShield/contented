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
	"strings"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/nulls"
)

/**
 * Grab the known donut file.
 */
func CreateVideoContainer(as *ActionSuite) (*models.Container, *models.Content) {
	cnt, contents := CreateVideoContents(as, "dir2", "donut")
	as.NotNil(contents, "No donut video found in dir2!")
	ref := *contents
	return cnt, &ref[0]
}

/*
 * TODO: Add to test common and actually do a full video validation.
 */
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
func CreateVideoContents(as *ActionSuite, containerName string, contentMatch string) (*models.Container, *models.Contents) {
	cntToCreate, contents := test_common.GetContentByDirName(containerName)

	cRes := as.JSON("/containers/").Post(&cntToCreate)
	as.Equal(http.StatusCreated, cRes.Code, fmt.Sprintf("It should create the container %s", cRes.Body.String()))

	cnt := models.Container{}
	json.NewDecoder(cRes.Body).Decode(&cnt)
	as.NotZero(cnt.ID, "It should create a valid container")

	contentsCreated := models.Contents{}
	for _, contentToCreate := range contents {

		if IsVideoMatch(contentToCreate, contentMatch) {
			contentToCreate.ContainerID = nulls.NewUUID(cnt.ID)
			contentRes := as.JSON("/contents").Post(&contentToCreate)
			as.Equal(http.StatusCreated, contentRes.Code, fmt.Sprintf("Error %s", contentRes.Body.String()))

			content := models.Content{}
			json.NewDecoder(contentRes.Body).Decode(&content)
			as.NotZero(content.ID, fmt.Sprintf("It should have created content %s", content))
			contentsCreated = append(contentsCreated, content)
		}
	}
	return &cnt, &contentsCreated
}

func (as *ActionSuite) Test_TaskRelatedObjects() {
	as.Equal(models.TaskOperation.SCREENS.String(), "screen_capture")
	as.Equal(models.TaskOperation.ENCODING.String(), "video_encoding")
	as.Equal(models.TaskOperation.WEBP.String(), "webp_from_screens")
	as.Equal(models.TaskOperation.TAGGING.String(), "tag_content")
	as.Equal(models.TaskOperation.DUPES.String(), "detect_duplicates")
}

// Do the screen grab in memory
func (as *ActionSuite) Test_MemoryEditingQueueScreenHandler() {
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = false
	utils.SetCfg(*cfg)
	test_common.InitMemoryFakeAppEmpty()
	as.Equal(cfg.ReadOnly, false)
	ValidateEditingQueue(as)
}

func (as *ActionSuite) Test_DbEditingQueueScreenHandler() {
	test_common.InitFakeApp(true)
	ValidateEditingQueue(as)
}

func ValidateEditingQueue(as *ActionSuite) {
	_, content := CreateVideoContainer(as)
	timeSeconds := 3
	screenCount := 1
	url := fmt.Sprintf("/editing_queue/%s/screens/%d/%d", content.ID.String(), screenCount, timeSeconds)
	res := as.JSON(url).Post(&content)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Should be able to grab a screen %s", res.Body.String()))

	// Huh... ODD
	tr := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&tr)
	as.NotZero(tr.ID, fmt.Sprintf("Did not create a Task %s", res.Body.String()))
	as.Equal(models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	as.Equal(tr.Operation, models.TaskOperation.SCREENS)

	args := worker.Args{"id": tr.ID.String()}
	err := ScreenCaptureWrapper(args)
	as.NoError(err, fmt.Sprintf("Failed to get screens %s", err))

	screenUrl := fmt.Sprintf("/contents/%s/screens", content.ID.String())
	screensRes := as.JSON(screenUrl).Get()
	as.Equal(http.StatusOK, screensRes.Code, fmt.Sprintf("Error loading screens %s", screensRes.Body.String()))

	sres := ScreensResponse{}
	json.NewDecoder(screensRes.Body).Decode(&sres)
	as.Equal(screenCount, len(sres.Results), fmt.Sprintf("We should have a set number of screens %s", sres.Results))
	as.Equal(screenCount, sres.Total, "The count should be correct")

	// Validate the task is now done
	checkR := as.JSON(fmt.Sprintf("/task_requests/%s", tr.ID.String())).Get()
	as.Equal(http.StatusOK, checkR.Code)
	checkTask := models.TaskRequest{}
	json.NewDecoder(checkR.Body).Decode(&checkTask)
	as.Equal(checkTask.Status, models.TaskStatus.DONE, fmt.Sprintf("It should be done %s", checkTask))
}

func (as *ActionSuite) Xest_MemoryEncodingQueueHandler() {
	// Should add a config value to completely nuke the encoded video
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = false
	utils.SetCfg(*cfg)
	test_common.InitMemoryFakeAppEmpty()
	ValidateVideoEncodingQueue(as)
}

func (as *ActionSuite) Xest_DBEncodingQueueHandler() {
	// Should add a config value to completely nuke the encoded video
	models.DB.TruncateAll()
	test_common.InitFakeApp(true)
	ValidateVideoEncodingQueue(as)
}

func ValidateVideoEncodingQueue(as *ActionSuite) {
	cnt, content := CreateVideoContainer(as)
	url := fmt.Sprintf("/editing_queue/%s/encoding", content.ID.String())
	res := as.JSON(url).Post(&content)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Failed to queue encoding task %s", res.Body.String()))

	tr := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&tr)
	as.NotZero(tr.ID, fmt.Sprintf("Did not create a Task %s", res.Body.String()))
	as.Equal(models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	as.Equal(tr.Operation, models.TaskOperation.ENCODING)

	args := worker.Args{"id": tr.ID.String()}
	err := VideoEncodingWrapper(args)
	as.NoError(err, fmt.Sprintf("Failed to encode video %s", err))

	checkR := as.JSON(fmt.Sprintf("/task_requests/%s", tr.ID.String())).Get()
	as.Equal(http.StatusOK, checkR.Code)

	checkTask := models.TaskRequest{}
	json.NewDecoder(checkR.Body).Decode(&checkTask)
	as.Equal(checkTask.Status, models.TaskStatus.DONE, fmt.Sprintf("It should be done %s", checkTask))

	createdID := checkTask.CreatedID.UUID
	as.NotZero(createdID, "It should create a new piece of content")
	check := as.JSON(fmt.Sprintf("/contents/%s", createdID.String())).Get()
	as.Equal(http.StatusOK, check.Code, fmt.Sprintf("Error loading %s", check.Body.String()))
	checkContent := models.Content{}
	json.NewDecoder(check.Body).Decode(&checkContent)

	as.Equal(checkContent.ContainerID.UUID, cnt.ID)
	as.Contains(checkContent.Src, "h256")

	// The container path is hidden in the API so load the actual DB el
	// TODO: DO NOT CHECK IN, I really need a faster encoding file.
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	cntActual, pathErr := man.GetContainer(cnt.ID)
	as.NoError(pathErr)

	dstFile := filepath.Join(cntActual.GetFqPath(), checkContent.Src)
	if _, err := os.Stat(dstFile); !os.IsNotExist(err) {
		os.Remove(dstFile)
	} else {
		as.Fail("It did NOT remove the destination file %s", dstFile)
	}
}

func (as *ActionSuite) Test_DBWebpHandler() {
	models.DB.TruncateAll()
	cfg := test_common.InitFakeApp(true)    // Probably should do the truncate in the InitFakeApp?
	test_common.Get_VideoAndSetupPaths(cfg) // Resets the screens
	cfg.ScreensOverSize = 1
	cfg.PreviewVideoType = "screens"
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(as)
	ValidateWebpCode(as, content)
	// Create some screens and then encode it?
}

func (as *ActionSuite) Test_MemoryWebpHandler() {
	cfg := test_common.InitFakeApp(false)
	test_common.Get_VideoAndSetupPaths(cfg) // Resets the screens

	cfg.ScreensOverSize = 1
	cfg.PreviewVideoType = "screens"
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(as)
	ValidateWebpCode(as, content)
}

func ValidateWebpCode(as *ActionSuite, content *models.Content) {
	as.Equal(content.Preview, "", "It should not have a preview already")
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)
	_, screenErr, _ := managers.CreateScreensForContent(man, content.ID, 10, 1)
	as.NoError(screenErr)

	url := fmt.Sprintf("/editing_queue/%s/webp", content.ID.String())
	res := as.JSON(url).Post(&content)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Failed to queue encoding task %s", res.Body.String()))

	tr := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&tr)
	as.NotZero(tr.ID, fmt.Sprintf("WebP did not create a Task %s", res.Body.String()))
	as.Equal(models.TaskStatus.NEW, tr.Status, fmt.Sprintf("Task invalid %s", tr))
	as.Equal(tr.Operation, models.TaskOperation.WEBP)

	args := worker.Args{"id": tr.ID.String()}
	err := WebpFromScreensWrapper(args)
	as.NoError(err, fmt.Sprintf("Failed to create webp for task %s", err))

	check := as.JSON(fmt.Sprintf("/contents/%s", content.ID.String())).Get()
	as.Equal(http.StatusOK, check.Code, fmt.Sprintf("Error loading %s", check.Body.String()))
	// Get the content, check for a preview
	checkContent := models.Content{}
	json.NewDecoder(check.Body).Decode(&checkContent)
	as.Equal("/container_previews/donut_[special( gunk.mp4.webp", checkContent.Preview)
}

func (as *ActionSuite) Test_DBTagHandler() {
	cfg := test_common.InitFakeApp(true)
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(as)
	ValidateTaggingCode(as, content)
}

func (as *ActionSuite) Test_MemoryTagHandler() {
	cfg := test_common.InitFakeApp(false)
	utils.SetCfg(*cfg)
	_, content := CreateVideoContainer(as)
	ValidateTaggingCode(as, content)
}

func ValidateTaggingCode(as *ActionSuite, content *models.Content) {
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	tag1 := models.Tag{ID: "donut"}
	tag2 := models.Tag{ID: "mp4"}
	badTag := models.Tag{ID: "THIS_WILL_NOT_MATCH"}
	man.CreateTag(&tag1)
	man.CreateTag(&tag2)
	man.CreateTag(&badTag)

	tr := models.TaskRequest{Operation: models.TaskOperation.TAGGING, ContentID: nulls.NewUUID(content.ID)}
	task, err := man.CreateTask(&tr)
	as.NoError(err, "Failed to create Task to do tagging")
	as.NotZero(task.ID)

	tagErr := managers.TaggingContentTask(man, task.ID)
	as.NoError(tagErr, "It should not have a problem doing the tagging")

	contentTagged, errLoad := man.GetContent(content.ID)
	as.NoError(errLoad, "It should be able to get the content back")

	as.Equal(2, len(contentTagged.Tags), fmt.Sprintf("There should be tags now %s", contentTagged))

	taskCheck, taskErr := man.GetTask(task.ID)
	as.NoError(taskErr, "We should still have a task")
	as.Equal(models.TaskStatus.DONE, taskCheck.Status)

	url := fmt.Sprintf("/editing_queue/%s/tagging", content.ID.String())
	res := as.JSON(url).Post(&content)
	as.Equal(http.StatusCreated, res.Code, fmt.Sprintf("Failed to queue tagging task %s", res.Body.String()))

}

func (as *ActionSuite) Test_DuplicateHandlerDB() {
	cfg := test_common.InitFakeApp(true)
	utils.SetCfg(*cfg)

	// Create the directory with the duplicate test
	cnt, contents := CreateVideoContents(as, "test_encoding", "")
	as.NotNil(cnt)
	as.NotNil(contents)

	ValidateDuplicatesTask(as, cnt)
}

func ValidateDuplicatesTask(as *ActionSuite, container *models.Container) {
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	tr := models.TaskRequest{Operation: models.TaskOperation.DUPES, ContainerID: nulls.NewUUID(container.ID)}
	task, err := man.CreateTask(&tr)
	as.NoError(err, "It should be able to create the duplicates task")

	dupeErr := managers.DetectDuplicatesTask(man, task.ID)
	as.NoError(dupeErr, "It should be able to run the duplicates task")

	// TODO: Check the task for our nice pretty format result for the task.
}

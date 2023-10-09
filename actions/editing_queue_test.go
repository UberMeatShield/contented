package actions

import (
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"contented/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/nulls"
)

func CreateVideoContainer(as *ActionSuite) (*models.Container, *models.Content) {
	cntToCreate, contents := test_common.GetContentByDirName("dir2")

	cRes := as.JSON("/containers/").Post(&cntToCreate)
	as.Equal(http.StatusCreated, cRes.Code, fmt.Sprintf("It should create the container %s", cRes.Body.String()))

	cnt := models.Container{}
	json.NewDecoder(cRes.Body).Decode(&cnt)
	as.NotZero(cnt.ID, "It should create a valid container")

	content := models.Content{}
	for _, contentToCreate := range contents {
		if strings.Contains(contentToCreate.Src, "donut") {
			contentToCreate.ContainerID = nulls.NewUUID(cnt.ID)
			contentRes := as.JSON("/content").Post(&contentToCreate)
			as.Equal(http.StatusCreated, contentRes.Code, fmt.Sprintf("Error %s", contentRes.Body.String()))
			json.NewDecoder(contentRes.Body).Decode(&content)
			break
		}
	}
	as.NotZero(content.ID, fmt.Sprintf("It should have created a donut content %s", content))
	return &cnt, &content
}

func (as *ActionSuite) Test_TaskRelatedObjects() {
	as.Equal(models.TaskOperation.SCREENS.String(), "screen_capture")
	as.Equal(models.TaskOperation.ENCODING.String(), "video_encoding")
	as.Equal(models.TaskOperation.WEBP.String(), "web_from_screens")
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

	screenUrl := fmt.Sprintf("/content/%s/screens", content.ID.String())
	screensRes := as.JSON(screenUrl).Get()
	as.Equal(http.StatusOK, screensRes.Code, fmt.Sprintf("Error loading screens %s", screensRes.Body.String()))

	screens := models.Screens{}
	json.NewDecoder(screensRes.Body).Decode(&screens)
	as.Equal(screenCount, len(screens), fmt.Sprintf("We should have a set number of screens %s", screens))

	// Validate the task is now done
	checkR := as.JSON(fmt.Sprintf("/task_requests/%s", tr.ID.String())).Get()
	as.Equal(http.StatusOK, checkR.Code)
	checkTask := models.TaskRequest{}
	json.NewDecoder(checkR.Body).Decode(&checkTask)
	as.Equal(checkTask.Status, models.TaskStatus.DONE, fmt.Sprintf("It should be done %s", checkTask))
}

func (as *ActionSuite) Test_MemoryEncodingQueueHandler() {
	// Should add a config value to completely nuke the encoded video
	cfg := test_common.ResetConfig()
	cfg.UseDatabase = false
	utils.SetCfg(*cfg)
	test_common.InitMemoryFakeAppEmpty()
	ValidateVideoEncodingQueue(as)
}

func (as *ActionSuite) Test_DBEncodingQueueHandler() {
	// Should add a config value to completely nuke the encoded video
	models.DB.TruncateAll()
	test_common.InitFakeApp(true)
	ValidateVideoEncodingQueue(as)
}

func ValidateVideoEncodingQueue(as *ActionSuite) {
	_, content := CreateVideoContainer(as)
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

	check := as.JSON(fmt.Sprintf("/content/%s", content.ID.String())).Get()
	as.Equal(http.StatusOK, check.Code, fmt.Sprintf("Error loading %s", check.Body.String()))
	// Get the content, check for a preview
	checkContent := models.Content{}
	json.NewDecoder(check.Body).Decode(&checkContent)
	as.Equal("/container_previews/donut_[special( gunk.mp4.webp", checkContent.Preview)
}

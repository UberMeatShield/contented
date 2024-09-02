package actions

import (
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateTask(contentID int64, t *testing.T, man managers.ContentManager) *models.TaskRequest {
	tr := &models.TaskRequest{
		ContentID: &contentID,
		Status:    models.TaskStatus.NEW,
	}
	task, err := man.CreateTask(tr)
	assert.NoError(t, err)
	return task
}

func TestTaskRequestApiMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTaskRequestListApi(t, router)
}

func TestTaskRequestApiDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	test_common.CreateContentByDirName("dir1")
	ValidateTaskRequestListApi(t, router)
}

func ValidateTaskRequestListApi(t *testing.T, router *gin.Engine) {
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	contents, count, err := man.ListContent(managers.ContentQuery{PerPage: 5})
	assert.NoError(t, err)
	assert.Greater(t, len(*contents), 0)
	assert.Greater(t, count, int64(0))
	for _, content := range *contents {
		CreateTask(content.ID, t, man)
	}
	taskRequests := TaskRequestResponse{}
	code, tErr := GetJson("/api/task_requests", "", &taskRequests, router)
	assert.Equal(t, http.StatusOK, code, fmt.Sprintf("Failed /api/task_requests call %s", tErr))
	assert.Equal(t, len(*contents), len(taskRequests.Results), "We should have a task for each content")

	// Update a status and query by status
	tUp := taskRequests.Results[0]
	assert.Equal(t, models.TaskStatus.NEW, tUp.Status)

	tUp.Status = models.TaskStatus.DONE
	tUp.Message = "Yay"
	_, upErr := man.UpdateTask(&tUp, models.TaskStatus.NEW)
	assert.NoError(t, upErr)

	// Do a query by container ID
	upFilter := TaskRequestResponse{}
	url := fmt.Sprintf("/api/task_requests?status=%s", tUp.Status.String())
	upCode, upErr := GetJson(url, "", &upFilter, router)
	assert.Equal(t, http.StatusOK, upCode, fmt.Sprintf("Updated query failed %s", upErr))
	assert.Equal(t, 1, len(upFilter.Results), fmt.Sprintf("There should be only one task %s", upFilter.Results))

	content := (*contents)[2]

	cUrl := fmt.Sprintf("/api/task_requests?content_id=%d", content.ID)
	tIdFilter := TaskRequestResponse{}
	contentCode, contentErr := GetJson(cUrl, "", &tIdFilter, router)
	assert.Equal(t, http.StatusOK, contentCode, fmt.Sprintf("Failed to filter on content ID %s", contentErr))
	assert.Equal(t, 1, len(tIdFilter.Results), fmt.Sprintf("It should search by contentID %s", tIdFilter.Results))
}

func TestCancelTaskMemory(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTaskRequestUpdate(t, router)
}

func TestCancelTaskDB(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	test_common.CreateContentByDirName("dir1")
	ValidateTaskRequestUpdate(t, router)
}

// Should do this for DB and memory
func ValidateTaskRequestUpdate(t *testing.T, router *gin.Engine) {
	ctx := test_common.GetContext()
	man := managers.GetManager(ctx)

	contents, _, err := man.ListContent(managers.ContentQuery{PerPage: 1})
	assert.NoError(t, err)
	assert.Equal(t, len(*contents), 1)

	content := (*contents)[0]
	task := CreateTask(content.ID, t, man)
	assert.NotEmpty(t, task.ID)
	assert.NotEqual(t, task.ID, 0)

	url := fmt.Sprintf("/api/task_requests/%d", task.ID)
	taskCheck := models.TaskRequest{}
	taskCode, taskErr := GetJson(url, "", &taskCheck, router)
	assert.Equal(t, http.StatusOK, taskCode, fmt.Sprintf("Failed to get task %s", taskErr))
	assert.Equal(t, taskCheck.ID, task.ID)

	taskCheck.Status = models.TaskStatus.ERROR
	upUrl := fmt.Sprintf("/api/task_requests/%d", taskCheck.ID)
	upCode, upErr := PutJson(upUrl, taskCheck, &models.TaskRequest{}, router)
	assert.Equal(t, http.StatusBadRequest, upCode, fmt.Sprintf("It should fail cancel only %s", upErr))

	taskCheck.Status = models.TaskStatus.CANCELED
	validateTask := models.TaskRequest{}
	upOk, unexpectedErr := PutJson(upUrl, taskCheck, &validateTask, router)
	assert.Equal(t, upOk, http.StatusOK, fmt.Sprintf("This should work as this is a cancel %s", unexpectedErr))
	assert.Equal(t, models.TaskStatus.CANCELED, validateTask.Status, "Its should have updated")
}

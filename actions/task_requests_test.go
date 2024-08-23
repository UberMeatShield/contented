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

//    "net/url"

func CreateTask(contentID int64, t *testing.T, man managers.ContentManager) *models.TaskRequest {
	tr := &models.TaskRequest{
		ContentID: &contentID,
		Status:    models.TaskStatus.NEW,
	}
	task, err := man.CreateTask(tr)
	assert.NoError(t, err)
	return task
}

func TestMemoryTaskRequestApi(t *testing.T) {
	_, _, router := InitFakeRouterApp(false)
	ValidateTaskRequestListApi(t, router)
}

func Test_DatabaseTaskRequestApi(t *testing.T) {
	_, _, router := InitFakeRouterApp(true)
	test_common.CreateContentByDirName("dir1")
	ValidateTaskRequestListApi(t, router)
}

/*
func (as *ActionSuite) Test_MemoryCancelTask() {
	useDB := false
	test_common.InitFakeApp(useDB)
	ValidateTaskRequestUpdate(as)
}

func (as *ActionSuite) Test_DatabaseCancelTask() {
	useDB := true
	test_common.InitFakeApp(useDB)
	test_common.CreateContentByDirName("dir1")
	ValidateTaskRequestUpdate(as)
}
*/

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

	/*
		// Update a status and query by status
		tUp := taskRequests.Results[0]
		assert.Equal(t, models.TaskStatus.NEW, tUp.Status)

		tUp.Status = models.TaskStatus.DONE
		tUp.Message = "Yay"
		_, upErr := man.UpdateTask(&tUp, models.TaskStatus.NEW)
		assert.NoError(t, upErr)

		// Do a query by container ID
		upFilter := TaskRequestResponse{}
		url := fmt.Sprintf("/task_requests?status=%s", tUp.Status.String())
		resUp := as.JSON(url).Get()
		assert.Equal(t, http.StatusOK, resUp.Code, fmt.Sprintf("Updated query failed %s", resUp.Body.String()))
		json.NewDecoder(resUp.Body).Decode(&upFilter)
		assert.Equal(t, 1, len(upFilter.Results), fmt.Sprintf("There should be only one task %s", upFilter.Results))

		content := (*contents)[2]
		cUrl := fmt.Sprintf("/task_requests?content_id=%s", content.ID.String())
		resContentFilter := as.JSON(cUrl).Get()
		assert.Equal(t, http.StatusOK, resContentFilter.Code, fmt.Sprintf("Failed to filter on content ID %s", resContentFilter.Body.String()))

		tIdFilter := TaskRequestResponse{}
		json.NewDecoder(resContentFilter.Body).Decode(&tIdFilter)
		assert.Equal(t, 1, len(tIdFilter.Results), fmt.Sprintf("It should search by contentID %s", tIdFilter.Results))
	*/
}

/*
// Should do this for DB and memory
func ValidateTaskRequestUpdate(as *ActionSuite) {
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	contents, _, err := man.ListContent(managers.ContentQuery{PerPage: 1})
	assert.NoError(t, err)
	assert.Equal(t, len(*contents), 1)

	content := (*contents)[0]
	task := CreateTask(content.ID, as, man)
	as.NotEmpty(task.ID)
	as.NotEqual(task.ID.String(), "00000000-0000-0000-0000-000000000000")

	url := fmt.Sprintf("/task_requests/%s", task.ID.String())
	res := as.JSON(url).Get()
	assert.Equal(t, http.StatusOK, res.Code)

	taskCheck := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&taskCheck)
	assert.Equal(t, taskCheck.ID, task.ID)

	taskCheck.Status = models.TaskStatus.ERROR
	upUrl := fmt.Sprintf("/task_requests/%s", taskCheck.ID.String())
	upErr := as.JSON(upUrl).Put(&taskCheck)
	assert.Equal(t, upErr.Code, http.StatusBadRequest, "It should fail cancel only")

	taskCheck.Status = models.TaskStatus.CANCELED
	upOk := as.JSON(upUrl).Put(&taskCheck)
	assert.Equal(t, upOk.Code, http.StatusOK, "This should work as this is a cancel")
}
*/

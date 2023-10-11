package actions

import (

	//    "net/url"
	"contented/managers"
	"contented/models"
	"contented/test_common"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
)

func CreateTask(contentID uuid.UUID, as *ActionSuite, man managers.ContentManager) *models.TaskRequest {
	tr := &models.TaskRequest{
		ContentID: contentID,
		Status:    models.TaskStatus.NEW,
	}
	t, err := man.CreateTask(tr)
	as.NoError(err)
	return t
}

func (as *ActionSuite) Test_MemoryTaskRequestApi() {
	useDB := false
	test_common.InitFakeApp(useDB)
	ValidateTaskRequestListApi(as)
}

func (as *ActionSuite) Test_DatabaseTaskRequestApi() {
	useDB := true
	test_common.InitFakeApp(useDB)
	test_common.CreateContentByDirName("dir1")
	ValidateTaskRequestListApi(as)
}

func ValidateTaskRequestListApi(as *ActionSuite) {
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	contents, count, err := man.ListContent(managers.ContentQuery{PerPage: 5})
	as.NoError(err)
	as.Greater(len(*contents), 0)
	as.Greater(count, 0)
	for _, content := range *contents {
		CreateTask(content.ID, as, man)
	}
	taskRequests := models.TaskRequests{}
	res := as.JSON("/task_requests/").Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Failed /task_requests call %s", res.Body))
	json.NewDecoder(res.Body).Decode(&taskRequests)
	as.Equal(len(*contents), len(taskRequests), "We should have a task for each content")

	// Update a status and query by status
	tUp := taskRequests[0]
	as.Equal(models.TaskStatus.NEW, tUp.Status)

	tUp.Status = models.TaskStatus.DONE
	tUp.Message = "Yay"
	_, upErr := man.UpdateTask(&tUp, models.TaskStatus.NEW)
	as.NoError(upErr)

	// Do a query by container ID
	upFilter := models.TaskRequests{}
	url := fmt.Sprintf("/task_requests?status=%s", tUp.Status.String())
	resUp := as.JSON(url).Get()
	as.Equal(http.StatusOK, resUp.Code, fmt.Sprintf("Updated query failed %s", resUp.Body.String()))
	json.NewDecoder(resUp.Body).Decode(&upFilter)
	as.Equal(1, len(upFilter), fmt.Sprintf("There should be only one task %s", upFilter))

	content := (*contents)[2]
	cUrl := fmt.Sprintf("/task_requests?content_id=%s", content.ID.String())
	resContentFilter := as.JSON(cUrl).Get()
	as.Equal(http.StatusOK, resContentFilter.Code, fmt.Sprintf("Failed to filter on content ID %s", resContentFilter.Body.String()))

	tIdFilter := models.TaskRequests{}
	json.NewDecoder(resContentFilter.Body).Decode(&tIdFilter)
	as.Equal(1, len(tIdFilter), fmt.Sprintf("It should search by contentID %s", tIdFilter))
}

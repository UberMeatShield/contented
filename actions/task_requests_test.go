package actions

//    "net/url"

/*
func CreateTask(contentID uuid.UUID, as *ActionSuite, man managers.ContentManager) *models.TaskRequest {
	tr := &models.TaskRequest{
		ContentID: nulls.NewUUID(contentID),
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
	taskRequests := TaskRequestResponse{}
	res := as.JSON("/task_requests/").Get()
	as.Equal(http.StatusOK, res.Code, fmt.Sprintf("Failed /task_requests call %s", res.Body))
	json.NewDecoder(res.Body).Decode(&taskRequests)
	as.Equal(len(*contents), len(taskRequests.Results), "We should have a task for each content")

	// Update a status and query by status
	tUp := taskRequests.Results[0]
	as.Equal(models.TaskStatus.NEW, tUp.Status)

	tUp.Status = models.TaskStatus.DONE
	tUp.Message = "Yay"
	_, upErr := man.UpdateTask(&tUp, models.TaskStatus.NEW)
	as.NoError(upErr)

	// Do a query by container ID
	upFilter := TaskRequestResponse{}
	url := fmt.Sprintf("/task_requests?status=%s", tUp.Status.String())
	resUp := as.JSON(url).Get()
	as.Equal(http.StatusOK, resUp.Code, fmt.Sprintf("Updated query failed %s", resUp.Body.String()))
	json.NewDecoder(resUp.Body).Decode(&upFilter)
	as.Equal(1, len(upFilter.Results), fmt.Sprintf("There should be only one task %s", upFilter.Results))

	content := (*contents)[2]
	cUrl := fmt.Sprintf("/task_requests?content_id=%s", content.ID.String())
	resContentFilter := as.JSON(cUrl).Get()
	as.Equal(http.StatusOK, resContentFilter.Code, fmt.Sprintf("Failed to filter on content ID %s", resContentFilter.Body.String()))

	tIdFilter := TaskRequestResponse{}
	json.NewDecoder(resContentFilter.Body).Decode(&tIdFilter)
	as.Equal(1, len(tIdFilter.Results), fmt.Sprintf("It should search by contentID %s", tIdFilter.Results))
}

// Should do this for DB and memory
func ValidateTaskRequestUpdate(as *ActionSuite) {
	ctx := test_common.GetContext(as.App)
	man := managers.GetManager(&ctx)

	contents, _, err := man.ListContent(managers.ContentQuery{PerPage: 1})
	as.NoError(err)
	as.Equal(len(*contents), 1)

	content := (*contents)[0]
	task := CreateTask(content.ID, as, man)
	as.NotEmpty(task.ID)
	as.NotEqual(task.ID.String(), "00000000-0000-0000-0000-000000000000")

	url := fmt.Sprintf("/task_requests/%s", task.ID.String())
	res := as.JSON(url).Get()
	as.Equal(http.StatusOK, res.Code)

	taskCheck := models.TaskRequest{}
	json.NewDecoder(res.Body).Decode(&taskCheck)
	as.Equal(taskCheck.ID, task.ID)

	taskCheck.Status = models.TaskStatus.ERROR
	upUrl := fmt.Sprintf("/task_requests/%s", taskCheck.ID.String())
	upErr := as.JSON(upUrl).Put(&taskCheck)
	as.Equal(upErr.Code, http.StatusBadRequest, "It should fail cancel only")

	taskCheck.Status = models.TaskStatus.CANCELED
	upOk := as.JSON(upUrl).Put(&taskCheck)
	as.Equal(upOk.Code, http.StatusOK, "This should work as this is a cancel")
}
*/

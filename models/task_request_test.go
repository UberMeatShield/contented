package models

func (ms *ModelSuite) Test_TaskRequests() {
	content := Content{Src: "A"}
	err := ms.DB.Save(&content)
	ms.NoError(err, "It should create content")

	tr := TaskRequest{
		Status:    TaskStatus.PENDING,
		Operation: TaskOperation.ENCODING,
		ContentID: content.ID,
	}
	ms.NoError(ms.DB.Create(&tr), "It should be able to create a Task Request")
	ms.NotZero(tr.ID, "It should have inserted a uuid")

	// Validate it did the right thing with the 'enum'
	check := TaskRequest{}
	ms.NoError(ms.DB.Find(&check, tr.ID), "It should have the task request")
	ms.Equal(check.Status, TaskStatus.PENDING)
	ms.Equal(check.Operation, TaskOperation.ENCODING)

	check.ErrMsg = "Something wicked this way comes"
	check.Status = TaskStatus.ERROR
	ms.NoError(ms.DB.Update(&check))
}

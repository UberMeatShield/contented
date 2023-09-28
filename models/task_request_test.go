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

	check := TaskRequest{}
	ms.NoError(ms.DB.Find(&check, tr.ID), "It should have the task request")
	ms.Equal(tr.Status, "Yar")
	ms.Equal(tr.Operation, "Zug")
}

package models

import (
	"testing"
)

func TestTaskRequestContent(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	content := Content{Src: "A"}
	cTx := db.Save(&content)
	NoError(cTx, "It should create content to make a task", t)

	tr := TaskRequest{
		Status:    TaskStatus.PENDING,
		Operation: TaskOperation.ENCODING,
		ContentID: &content.ID,
	}

	NoError(db.Create(&tr), "it should be able to create a Task Request", t)
	if tr.ID == 0 {
		t.Errorf("Did not create a task request %d", tr.ID)
	}

	// Validate it did the right thing with the 'enum'
	check := TaskRequest{}
	NoError(db.Find(&check, tr.ID), "It should have the task request", t)
	if check.Status != TaskStatus.PENDING {
		t.Errorf("Task had the wrong status %s", check.Status)
	}
	if check.Operation != TaskOperation.ENCODING {
		t.Errorf("Check had the wrong operation %s", check.Operation)
	}
	if check.ID == 0 {
		t.Errorf("Task request was not created")
	}

	check.ErrMsg = "Something wicked this way comes"
	check.Status = TaskStatus.ERROR
	NoError(db.Save(&check), "It should update the task request", t)
}

func TestTaskRequestsContainer(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	container := Container{Name: "ContainerPath"}
	cTx := db.Save(&container)
	NoError(cTx, "It should create a container", t)

	tr := TaskRequest{
		Status:      TaskStatus.PENDING,
		Operation:   TaskOperation.DUPES,
		ContainerID: &container.ID,
	}
	NoError(db.Create(&tr), "it should be able to create a Task Request", t)
	if tr.ID == 0 {
		t.Errorf("It should have inserted an id %d", tr.ID)
	}

	// Validate it did the right thing with the 'enum'
	check := TaskRequest{}
	NoError(db.Find(&check, tr.ID), "It should have the task request", t)
	if check.Status != TaskStatus.PENDING {
		t.Errorf("Status was not pending %s", check.Status)
	}
	if check.Operation != TaskOperation.DUPES {
		t.Errorf("Operation was not set to dupes %s", check.Operation)
	}

	check.ErrMsg = "Something wicked this way comes"
	check.Status = TaskStatus.ERROR
	NoError(db.Save(&check), "Could not update task status", t)
}

package actions

import (
	"errors"

	//"fmt"
	"net/http"
	// "errors"
	"contented/managers"

	"github.com/gobuffalo/buffalo"

	//"github.com/gobuffalo/pop/v5"

	"github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (TaskRequest)
// DB Table: Plural (task_request)
// Resource: Plural (TaskRequest)
// Path: Plural (/task_request)
// View Template Folder: Plural (/templates/task_request/)

// TaskRequestResource is the resource for the TaskRequest model
type TaskRequestResource struct {
	buffalo.Resource
}

// List gets all TaskRequest. This function is mapped to the path
// GET /task_request
func (v TaskRequestResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	man := managers.GetManager(&c)
	tasks, err := man.ListTasksContext()
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	return c.Render(200, r.JSON(tasks))
}

// Show gets the data for one TaskRequest. This function is mapped to
// the path GET /task_request/{task_request_id} ?
func (v TaskRequestResource) Show(c buffalo.Context) error {
	// TODO: Test this (not sure what the ID comes in as)
	tStrID := c.Param("task_request_id")
	tID, badUUID := uuid.FromString(tStrID)
	if badUUID != nil {
		return c.Error(400, badUUID)
	}
	man := managers.GetManager(&c)
	task, err := man.GetTask(tID)
	if err != nil {
		return c.Error(404, err)
	}
	return c.Render(http.StatusOK, r.JSON(task))
}

// Currently this is a private setup not accessible from the UI
func (v TaskRequestResource) Create(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
}

// Another private method, might be opened to just canceling a task (if possible)
func (v TaskRequestResource) Update(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
}

// Also a private setup, it is saner to not have somebody messing with the task queue.
func (v TaskRequestResource) Destroy(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
}

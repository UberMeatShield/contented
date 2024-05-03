package actions

import (
	"errors"
	"fmt"
	"log"

	//"fmt"
	"net/http"
	// "errors"
	"contented/managers"
	"contented/models"

	"github.com/gobuffalo/buffalo"
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

type TaskRequestResponse struct {
	Total   int                 `json:"total"`
	Results models.TaskRequests `json:"results"`
}

// List gets all TaskRequest. This function is mapped to the path
// GET /task_request
func (v TaskRequestResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	man := managers.GetManager(&c)
	tasks, total, err := man.ListTasksContext()
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	tres := TaskRequestResponse{
		Total:   total,
		Results: *tasks,
	}
	return c.Render(200, r.JSON(tres))
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
	_, _, err := managers.ManagerCanCUD(&c)
	if err != nil {
		return err
	}

	man := managers.GetManager(&c)
	id, _ := uuid.FromString(c.Param("task_request_id"))
	exists, err := man.GetTask(id)
	if err != nil || exists == nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Maybe this would be fine with a custom route an /ID/state on a put
	taskUp := models.TaskRequest{}
	if err := c.Bind(&taskUp); err != nil {
		msg := fmt.Sprintf("Bad TaskRequest passed %s", taskUp)
		log.Printf(msg)
		return c.Error(http.StatusBadRequest, errors.New(msg))
	}

	state := taskUp.Status
	if state == models.TaskStatus.INVALID {
		msg := fmt.Sprintf("Invalid state passed %s to task update", state)
		log.Printf(msg)
		return c.Error(http.StatusBadRequest, errors.New(msg))
	}

	if state != models.TaskStatus.CANCELED {
		msg := fmt.Sprintf("Currently only supports canceled. %s", state)
		log.Printf(msg)
		return c.Error(http.StatusBadRequest, errors.New(msg))
	}

	// Awkward states to handle, but the basic one is just going to be can we cancel
	task := *exists
	if !(task.Status == models.TaskStatus.NEW || task.Status == models.TaskStatus.PENDING) {
		msg := fmt.Sprintf("Cannot change state from current (%s) to %s", task.Status, state)
		log.Printf(msg)
		return c.Error(http.StatusBadRequest, errors.New(msg))
	}

	currentState := task.Status
	task.Status = state
	taskUpdated, upErr := man.UpdateTask(&task, currentState)
	if upErr != nil || taskUpdated == nil {
		log.Printf("Failed to update resource %s", upErr)
		return upErr
	}
	return c.Render(http.StatusOK, r.JSON(taskUpdated))
}

// Also a private setup, it is saner to not have somebody messing with the task queue.
func (v TaskRequestResource) Destroy(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
}

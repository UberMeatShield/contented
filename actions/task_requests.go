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

// Create adds a TaskRequest to the DB. This function is mapped to the
// path POST /task_request
func (v TaskRequestResource) Create(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
	/*
		_, _, err := managers.ManagerCanCUD(&c)
		if err != nil {
			return err
		}
		// Bind previewTaskRequest to the html form/JSON elements
		screen := &models.TaskRequest{}
		if err := c.Bind(screen); err != nil {
			return err
		}
		man := managers.GetManager(&c)
		cErr := man.CreateTaskRequest(screen)
		if cErr != nil {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(cErr))
		}
		return c.Render(http.StatusCreated, r.JSON(screen))
	*/
}

// Update changes a TaskRequest in the DB. This function is mapped to
// the path PUT /task_request/{screen_id}
func (v TaskRequestResource) Update(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
	// Get the DB connection from the context
	/*
		_, _, err := managers.ManagerCanCUD(&c)
		if err != nil {
			return err
		}

		man := managers.GetManager(&c)
		id, idErr := uuid.FromString(c.Param("screen_id"))
		if idErr != nil {
			return c.Error(http.StatusBadRequest, idErr)
		}
		screen, notFoundErr := man.GetTaskRequest(id)
		if notFoundErr != nil {
			return c.Error(http.StatusNotFound, err)
		}
		if err := c.Bind(screen); err != nil {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(err))
		}
		checkTaskRequest, _ := man.GetTaskRequest(id)
		return c.Render(http.StatusOK, r.JSON(checkTaskRequest))
	*/
}

// Destroy deletes a TaskRequest from the DB. This function is mapped
// to the path DELETE /task_request/{screen_id}
func (v TaskRequestResource) Destroy(c buffalo.Context) error {
	return c.Error(http.StatusNotImplemented, errors.New("Not implemented"))
	/*
		// Get the DB connection from the context
		_, _, err := managers.ManagerCanCUD(&c)
		if err != nil {
			return err
		}

		man := managers.GetManager(&c)
		screen, dErr := man.DestroyTaskRequest(c.Param("screen_id"))
		if dErr != nil {
			return c.Render(http.StatusBadRequest, r.JSON(dErr))
		}
		return c.Render(http.StatusOK, r.JSON(screen))
	*/
}

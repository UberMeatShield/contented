package actions

import (
	"contented/managers"
	//    "contented/models"
	"errors"
	"log"
	"time"

	//"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

// https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
type TaskRequestStatus struct {
	State string
}

// What are we doing here
type TaskRequest struct {
	contentID uuid.UUID
	status    TaskRequestStatus

	CreatedAt time.Time
	UpdatedAt time.Time

	Message string
	ErrMsg  string
}

// Should deny quickly if the media content type is incorrect for the action
func TaskScreenHandler(c buffalo.Context) error {
	mcID, bad_uuid := uuid.FromString(c.Param("mcID"))
	if bad_uuid != nil {
		return c.Error(400, bad_uuid)
	}
	man := managers.GetManager(&c)
	content, err := man.GetContent(mcID)
	if err != nil {
		return c.Error(404, err)
	}
	log.Printf("Not implemented")
	return c.Render(400, r.JSON(content))
}

func TaskScreensHandler(c buffalo.Context) error {

	return c.Error(400, errors.New("Not implemented"))
}

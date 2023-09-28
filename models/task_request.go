package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// This might move into a model
type TaskRequestStatus struct {
	State string
}
type TaskStatusType string

var TaskStatus = struct {
	NEW         TaskStatusType
	PENDING     TaskStatusType
	IN_PROGRESS TaskStatusType
	CANCELED    TaskStatusType
	ERROR       TaskStatusType
	DONE        TaskStatusType
}{
	NEW:         "new",
	PENDING:     "pending",
	IN_PROGRESS: "in_progress",
	CANCELED:    "canceled",
	ERROR:       "error",
	DONE:        "done",
}

type TaskOperationType string

var TaskOperation = struct {
	ENCODING TaskOperationType
	SCREENS  TaskOperationType
}{
	ENCODING: "video_encoding",
	SCREENS:  "screen_capture",
}

// GoLang makes this a little annoying to have random metadata
type TaskRequest struct {
	ID        int               `json:"ID"`
	ContentID uuid.UUID         `json:"content_id"`
	Status    TaskStatusType    `json:"status" default:"new"`
	Operation TaskOperationType `json:"operation"`

	// Initial default time would be nice
	CreatedAt time.Time `json:"created_at" default:"time.Now()"`
	UpdatedAt time.Time `json:"updated_at" default:"time.Now()"`
	Message   string    `json:"message" default:""`
	ErrMsg    string    `json:"err_msg" default:""`

	// Is it worth having two different queues for this?  Probably not, both use ffmpeg resource
	NumberOfScreens int    `json:"number_of_screens" default:"12"`
	StartTime       int    `json:"start_time" default:"0"`
	Codec           string `json:"codec" default:"libx265"`
	Width           int    `json:"width" default:"-1"`
	Height          int    `json:"height" default:"-1"`
}

// String is not required by pop and may be deleted
func (m TaskRequest) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Probably this interface will work and be less complex / anoying than having two types in the DB
type TaskRequests []TaskRequest

// String is not required by pop and may be deleted
func (m TaskRequests) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *TaskRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *TaskRequest) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *TaskRequest) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

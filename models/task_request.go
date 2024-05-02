package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/nulls"
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
	INVALID     TaskStatusType
}{
	NEW:         "new",
	PENDING:     "pending",
	IN_PROGRESS: "in_progress",
	CANCELED:    "canceled",
	ERROR:       "error",
	DONE:        "done",
	INVALID:     "invalid",
}

func (ts TaskStatusType) String() string {
	switch ts {
	case TaskStatus.NEW:
		return "new"
	case TaskStatus.PENDING:
		return "pending"
	case TaskStatus.IN_PROGRESS:
		return "in_progress"
	case TaskStatus.CANCELED:
		return "canceled"
	case TaskStatus.ERROR:
		return "error"
	case TaskStatus.DONE:
		return "done"
	case TaskStatus.INVALID:
		return "invalid"
	}
	return "unknown"
}

func GetTaskStatus(name string) TaskStatusType {
	switch name {
	case "new":
		return TaskStatus.NEW
	case "pending":
		return TaskStatus.PENDING
	case "in_progress":
		return TaskStatus.IN_PROGRESS
	case "canceled":
		return TaskStatus.CANCELED
	case "error":
		return TaskStatus.ERROR
	case "done":
		return TaskStatus.DONE
	}
	return TaskStatus.INVALID
}

func (ts TaskStatusType) Copy() TaskStatusType {
	return GetTaskStatus(ts.String())
}

type TaskOperationType string

var TaskOperation = struct {
	ENCODING TaskOperationType
	SCREENS  TaskOperationType
	WEBP     TaskOperationType
	TAGGING  TaskOperationType
}{
	ENCODING: "video_encoding",
	SCREENS:  "screen_capture",
	WEBP:     "webp_from_screens",
	TAGGING:  "tag_content",
}

func (to TaskOperationType) String() string {
	switch to {
	case TaskOperation.ENCODING:
		return "video_encoding"
	case TaskOperation.SCREENS:
		return "screen_capture"
	case TaskOperation.WEBP:
		return "webp_from_screens"
	case TaskOperation.TAGGING:
		return "tag_content"
	}
	return "unknown"
}

// TaskRequest is used by pop to map your task_requests database table to your go code.
type TaskRequest struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ContentID uuid.UUID `json:"content_id" db:"content_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	StartedAt time.Time `json:"started_at" db:"started_at"`

	Status    TaskStatusType    `json:"status" db:"status" default:"new" `
	Operation TaskOperationType `json:"operation" db:"operation"`
	CreatedID nulls.UUID        `json:"created_id" db:"created_id"`

	// Initial default time would be nice
	Message string `json:"message" default:"" db:"message"`
	ErrMsg  string `json:"err_msg" default:"" db:"err_message"`

	// Is it worth having two different queues for this?  Probably not, both use ffmpeg resource
	// Add once I have the basic processor in place
	NumberOfScreens  int    `json:"number_of_screens" default:"12" db:"number_of_screens"`
	StartTimeSeconds int    `json:"start_time_seconds" default:"0" db:"start_time_seconds"`
	Codec            string `json:"codec" default:"libx265" db:"codec"`
	Width            int    `json:"width" default:"-1" db:"width"`
	Height           int    `json:"height" default:"-1" db:"height"`
}

// String is not required by pop and may be deleted
func (t TaskRequest) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// TaskRequests is not required by pop and may be deleted
type TaskRequests []TaskRequest
type TaskRequestMap map[uuid.UUID]TaskRequest

// String is not required by pop and may be deleted
func (t TaskRequests) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *TaskRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *TaskRequest) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *TaskRequest) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// TODO: Potentially need to add in Retry helper logic.

package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
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
	ENCODING               TaskOperationType
	SCREENS                TaskOperationType
	WEBP                   TaskOperationType
	TAGGING                TaskOperationType
	DUPES                  TaskOperationType
	REMOVE_DUPLICATE_FILES TaskOperationType
}{
	ENCODING:               "video_encoding",
	SCREENS:                "screen_capture",
	WEBP:                   "webp_from_screens",
	TAGGING:                "tag_content",
	DUPES:                  "detect_duplicates",
	REMOVE_DUPLICATE_FILES: "remove_duplicate_files",
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
	case TaskOperation.DUPES:
		return "detect_duplicates"
	case TaskOperation.REMOVE_DUPLICATE_FILES:
		return "remove_duplicate_files"
	}
	return "unknown"
}

// TaskRequest is used by pop to map your task_requests database table to your go code.
type TaskRequest struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" db:"deleted_at"`

	// Need to get these all properly fk constrained
	ContentID   *int64 `json:"content_id" db:"content_id" gorm:"default:null"`
	ContainerID *int64 `json:"container_id" db:"container_id" gorm:"default:null"`
	CreatedID   *int64 `json:"created_id" db:"created_id" gorm:"default:null"`

	// TODO: Make it optional on ContentId so things cna work on a container?
	StartedAt time.Time `json:"started_at" db:"started_at"`

	Status    TaskStatusType    `json:"status" db:"status" default:"new" `
	Operation TaskOperationType `json:"operation" db:"operation"`

	// Initial default time would be nice
	Message string `json:"message" default:"" db:"message"`
	ErrMsg  string `json:"err_msg" default:"" db:"err_msg"`

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
type TaskRequestMap map[int]TaskRequest

// String is not required by pop and may be deleted
func (t TaskRequests) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

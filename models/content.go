package models

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	//"contented/actions"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// Content is used by pop to map your contents database table to your go code.
type Content struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated" db:"updated_at"`
	//DeletedAt   gorm.DeletedAt `gorm:"index"`
	Src         string `json:"src" db:"src"`
	ContentType string `json:"content_type" db:"content_type"`
	Preview     string `json:"preview" db:"preview"`
	ContainerID int    `json:"container_id" db:"container_id" default:"nil"`
	Idx         int    `json:"idx" db:"idx" default:"0"`
	Active      bool   `json:"active" db:"active" default:"true"`
	Corrupt     bool   `json:"corrupt" db:"corrupt" default:"false"`
	SizeBytes   int64  `json:"size" db:"size_bytes" default:"0"`
	Description string `json:"description" db:"description" default:""`
	NoFile      bool   `json:"no_file" db:"no_file" default:"false"` // Actual file or just description etc
	Hidden      bool   `json:"-" db:"hidden" default:"false"`        // Should it be visible in basic list queries

	// This is for information about the file content (video / image mostly stats, rez etc)
	Meta string `json:"meta" db:"meta" default:""`

	// Joins (Eager loading is not working?)
	Screens Screens `json:"screens" has_many:"preview_screens"`
	Tags    Tags    `json:"tags" many_to_many:"contents_tags"`

	// TODO: Maybe, MAYBE drop this?  None of the code currently really looks at the encoding
	// till actually creating a preview.
	Encoding string `json:"encoding" db:"encoding"`

	// Useful for when we built out media in a container and want to associate it.
	FqPath string `json:"-" db:"-" default:""` // NOT SET BY DEFAULT
}

// It seems odd there is no arbitrary json field => proper sort on the struct but then many of
// these struct elements do not have a default sort implemented soooo I guess this makes sense.
type ContentJsonSort func(i, j int) bool

var VALID_CONTENT_ORDERS = []string{
	"created_at",
	"updated_at",
	"content_type",
	"container_id",
	"idx",
	"size",
	"description",
}

// Contents is not required by pop and may be deleted
type Contents []Content
type ContentMap map[int]Content
type ContentMapBySrc map[string]Content

func GetContentSort(arr Contents, jsonFieldName string) ContentJsonSort {
	var theSort ContentJsonSort
	switch jsonFieldName {
	case "updated":
		theSort = func(i, j int) bool {
			return arr[i].UpdatedAt.Unix() < arr[j].UpdatedAt.Unix()
		}
	case "src":
		theSort = func(i, j int) bool {
			return strings.ToLower(arr[i].Src) < strings.ToLower(arr[j].Src)
		}
	case "content_type":
		theSort = func(i, j int) bool {
			return arr[i].ContentType < arr[j].ContentType
		}
	case "container_id":
		theSort = func(i, j int) bool {
			return arr[i].ContainerID < arr[j].ContainerID
		}
	case "size":
		theSort = func(i, j int) bool {
			return arr[i].SizeBytes < arr[j].SizeBytes
		}
	case "description":
		theSort = func(i, j int) bool {
			return arr[i].Description < arr[j].Description
		}
	case "created_at":
		theSort = func(i, j int) bool {
			return arr[i].CreatedAt.Unix() < arr[j].CreatedAt.Unix()
		}
	case "idx":
		theSort = func(i, j int) bool {
			return arr[i].Idx < arr[j].Idx
		}
	default:
		theSort = func(i, j int) bool {
			return arr[i].Idx < arr[j].Idx
		}
	}
	return theSort
}

func GetValidOrder(validOrders []string, order string, direction string, defaultOrder string) string {
	valid_order := defaultOrder
	if slices.Contains(VALID_CONTENT_ORDERS, order) {
		valid_order = order
	}
	valid_direction := "desc"
	if direction == "asc" || direction == "desc" {
		valid_direction = direction
	}
	return fmt.Sprintf("%s %s", valid_order, valid_direction)
}

func GetContentOrder(order string, direction string) string {
	return GetValidOrder(VALID_CONTENT_ORDERS, order, direction, "idx")
}

// String is not required by pop and may be deleted
func (m Content) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

func (arr Contents) Reverse() Contents {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// String is not required by pop and may be deleted
func (m Contents) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// String is not required by pop and may be deleted
func (content Content) IsVideo() bool {
	return strings.Contains(content.ContentType, "video")
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Content) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *Content) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *Content) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// This is a little risky as the tags might not be loaded on the object and there isn't
// a great way to tell 'loaded' vs just doesn't have tags
func (m *Content) HasTag(tag string) bool {
	tags := m.Tags
	if len(tags) == 0 {
		return false
	}
	for _, t := range tags {
		if t.ID == tag {
			return true
		}
	}
	return false
}

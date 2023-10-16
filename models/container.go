package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Container is used by pop to map your containers database table to your go code.
type Container struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Total       int       `json:"total" db:"total" default:"0"`
	Path        string    `json:"-" db:"path"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description" default:""`
	CreatedAt   time.Time `json:"created" db:"created_at"`
	UpdatedAt   time.Time `json:"updated" db:"updated_at"`
	Active      bool      `json:"active" db:"active" default:"true"`
	Idx         int       `json:"idx" db:"idx" default:"0"`
	Contents    Contents  `json:"contents" has_many:"contents" db:"-"`
	Hidden      bool      `json:"-" db:"hidden" default:"false"`

	// This is expected to be a URL where often a configured /preview/{mcID} is going
	// to be assigned by default.  However you should be able to use any link but it is
	// going to assume it is an image and won't do anything smart with it.
	PreviewUrl string `json:"previewUrl" db:"preview_url"`
	// TODO:  Should I add a preview type in the future?

}
type ContainerJsonSort func(i, j int) bool

func GetContainerSort(arr Containers, jsonFieldName string) ContentJsonSort {
	var theSort ContentJsonSort
	switch jsonFieldName {
	case "updated":
		theSort = func(i, j int) bool {
			return arr[i].UpdatedAt.Unix() < arr[j].UpdatedAt.Unix()
		}
	case "name":
		theSort = func(i, j int) bool {
			return arr[i].Name < arr[j].Name
		}
	case "total":
		theSort = func(i, j int) bool {
			return arr[i].Total < arr[j].Total
		}
	case "preview_url":
		theSort = func(i, j int) bool {
			return arr[i].PreviewUrl < arr[j].PreviewUrl
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
	default:
		theSort = func(i, j int) bool {
			return arr[i].Idx < arr[j].Idx
		}
	}
	return theSort
}

var VALID_CONTAINER_ORDERS = []string{"created_at", "updated_at", "total", "name", "preview_url", "description"}

func GetContainerOrder(order string, direction string) string {
	valid_order := "created_at"
	if slices.Contains(VALID_CONTENT_ORDERS, order) {
		valid_order = order
	}
	valid_direction := "desc"
	if direction == "asc" || direction == "desc" {
		valid_direction = direction
	}
	return fmt.Sprintf("%s %s", valid_order, valid_direction)
}

// String is not required by pop and may be deleted
func (c Container) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Containers is not required by pop and may be deleted
type Containers []Container
type ContainerMap map[uuid.UUID]Container

func (arr Containers) Reverse() Containers {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// String is not required by pop and may be deleted
func (c Containers) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Hmmmm (Unit tests were creating bad files in the mock dir)
func (c Container) GetFqPath() string {
	return filepath.Join(c.Path, c.Name)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Container) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *Container) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *Container) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

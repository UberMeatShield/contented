package models

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Container is used by pop to map your containers database table to your go code.
type Container struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Total     int       `json:"total" db:"total"`
	Path      string    `json:"-" db:"path"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created" db:"created_at"`
	UpdatedAt time.Time `json:"updated" db:"updated_at"`
	Active    bool      `json:"active" db:"active"`
	Idx       int       `json:"idx" db:"idx"`
	Contents  Contents  `json:"contents" has_many:"contents" db:"-"`
	Hidden    bool      `json:"-" db:"hidden"`

	// This is expected to be a URL where often a configured /preview/{mcID} is going
	// to be assigned by default.  However you should be able to use any link but it is
	// going to assume it is an image and won't do anything smart with it.
	PreviewUrl string `json:"previewUrl" db:"preview_url"`
	// TODO:  Should I add a preview type in the future?

}

// String is not required by pop and may be deleted
func (c Container) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Containers is not required by pop and may be deleted
type Containers []Container
type ContainerMap map[uuid.UUID]Container

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

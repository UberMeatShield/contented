package models

import (
	"time"
	"encoding/json"
    "path/filepath"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Container is used by pop to map your containers database table to your go code.
type Container struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Total     int             `json:"total" db:"total"`
	Path      string          `json:"path" db:"path"`
	Name      string          `json:"name" db:"name"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
    Active    bool            `json:"active" db:"active"`
    Idx       int             `json:"idx" db:"idx"`
	Contents  MediaContainers `json:"contents" has_many:"media_containers" db:"-"`

    // This could be made to be a media container reference but currently I am not sure
    // if that is better vs storing something that is a valid string and could be used
    // with more data sources?   As I type this I am thinking string....
    PreviewSrc string  `json:"preview_src" db:"preview_src"`
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

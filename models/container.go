package models

import (
	"encoding/json"
	"time"

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
	Contents  MediaContainers `json:"contents" has_many:"media_containers" db:"-"`
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

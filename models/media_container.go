package models

import (
    "encoding/json"
    "time"

    //"contented/actions"
    "github.com/gobuffalo/nulls"
    "github.com/gobuffalo/pop/v6"
    "github.com/gobuffalo/validate/v3"
    "github.com/gofrs/uuid"
)

// MediaContainer is used by pop to map your media_containers database table to your go code.
type MediaContainer struct {
    ID          uuid.UUID      `json:"id" db:"id"`
    CreatedAt   time.Time      `json:"created" db:"created_at"`
    UpdatedAt   time.Time      `json:"updated" db:"updated_at"`
    Src         string         `json:"src" db:"src"`
    ContentType string         `json:"content_type" db:"content_type"`
    Preview     string         `json:"preview" db:"preview"`
    ContainerID nulls.UUID     `json:"container_id" db:"container_id"`
    Idx         int            `json:"idx" db:"idx"`
    Active      bool           `json:"active" db:"active"`
    Corrupt     bool           `json:"corrupt" db:"corrupt"`
    SizeBytes   int64          `json:"size" db:"size_bytes"`
    Description string         `json:"description" db:"description"`

    // Joins (Eager loading is not working?)
    Screens Screens `json:"screens" has_many:"preview_screens"`
    Tags Tags `json:"tags" many_to_many:"media_containers_tags"`

    // TODO: Maybe, MAYBE drop this?  None of the code currently really looks at the encoding
    // till actually creating a preview.
    Encoding string `json:"encoding" db:"encoding"`
}

// String is not required by pop and may be deleted
func (m MediaContainer) String() string {
    jm, _ := json.Marshal(m)
    return string(jm)
}

// MediaContainers is not required by pop and may be deleted
type MediaContainers []MediaContainer
type MediaMap map[uuid.UUID]MediaContainer

// String is not required by pop and may be deleted
func (m MediaContainers) String() string {
    jm, _ := json.Marshal(m)
    return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *MediaContainer) Validate(tx *pop.Connection) (*validate.Errors, error) {
    return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *MediaContainer) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
    return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *MediaContainer) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
    return validate.NewErrors(), nil
}

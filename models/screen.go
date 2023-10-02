package models

// Damn it, this should have just been named screen
import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// A set of previews for a particular content element.
type Screen struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ContentID uuid.UUID `json:"content_id" db:"content_id"`
	CreatedAt time.Time `json:"created" db:"created_at"`
	UpdatedAt time.Time `json:"updated" db:"updated_at"`
	Path      string    `json:"-" db:"path"`
	Src       string    `json:"src" db:"src"`
	Idx       int       `json:"idx" db:"idx"`
	SizeBytes int64     `json:"size" db:"size_bytes"`
}

// String is not required by pop and may be deleted
func (m Screen) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Screens is not required by pop and may be deleted
type Screens []Screen
type ScreenMap map[uuid.UUID]Screen
type ScreenCollection map[uuid.UUID]Screens

// String is not required by pop and may be deleted
func (m Screens) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

func (m Screen) GetFqPath() string {
	return filepath.Join(m.Path, m.Src)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.                 ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Screen) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *Screen) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *Screen) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

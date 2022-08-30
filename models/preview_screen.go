package models

// Damn it, this should have just been named screen
import (
	"encoding/json"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"path/filepath"
	"time"
)

// A set of previews for a particular media element.
type PreviewScreen struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MediaID   uuid.UUID `json:"media_id" db:"media_container_id"`
	CreatedAt time.Time `json:"created" db:"created_at"`
	UpdatedAt time.Time `json:"updated" db:"updated_at"`
	Path      string    `json:"-" db:"path"`
	Src       string    `json:"src" db:"src"`
	Idx       int       `json:"idx" db:"idx"`
	SizeBytes int64     `json:"size" db:"size_bytes"`
}

// String is not required by pop and may be deleted
func (m PreviewScreen) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// PreviewScreens is not required by pop and may be deleted
type PreviewScreens []PreviewScreen
type PreviewScreenMap map[uuid.UUID]PreviewScreen
type PreviewScreenCollection map[uuid.UUID]PreviewScreens

// String is not required by pop and may be deleted
func (m PreviewScreens) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

func (m PreviewScreen) GetFqPath() string {
	return filepath.Join(m.Path, m.Src)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.                 ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *PreviewScreen) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *PreviewScreen) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *PreviewScreen) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

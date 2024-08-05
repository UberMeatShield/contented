package models

// Damn it, this should have just been named screen
import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"gorm.io/gorm"
)

// A set of previews for a particular content element.
type Screen struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	ContentID uint           `json:"content_id" db:"content_id"`
	Path      string         `json:"-" db:"path"`
	Src       string         `json:"src" db:"src"`
	Idx       int            `json:"idx" db:"idx"`
	SizeBytes int64          `json:"size_bytes" db:"size_bytes"`
}

type ScreensJsonSort func(i, j int) bool

var VALID_SCREENS_ORDERS = []string{"created_at", "updated_at", "idx", "size", "src", "content_id", "src"}

func GetScreensSort(arr Screens, jsonFieldName string) ContentJsonSort {
	var theSort ContentJsonSort
	switch jsonFieldName {
	case "updated_at":
		theSort = func(i, j int) bool {
			return arr[i].UpdatedAt.Unix() < arr[j].UpdatedAt.Unix()
		}
	case "created_at":
		theSort = func(i, j int) bool {
			return arr[i].CreatedAt.Unix() < arr[j].CreatedAt.Unix()
		}
	case "content_id":
		theSort = func(i, j int) bool {
			return arr[i].ContentID < arr[j].ContentID
		}
	case "size":
		theSort = func(i, j int) bool {
			return arr[i].SizeBytes < arr[j].SizeBytes
		}
	case "src":
		theSort = func(i, j int) bool {
			return arr[i].Src < arr[j].Src
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

// String is not required by pop and may be deleted
func (m Screen) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

func GetScreensOrder(order string, direction string) string {
	return GetValidOrder(VALID_SCREENS_ORDERS, order, direction, "idx")
}

// Screens is not required by pop and may be deleted
type Screens []Screen
type ScreenMap map[uint]Screen
type ScreenCollection map[uint]Screens

func (arr Screens) Reverse() Screens {
	if len(arr) > 1 {
		for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	return arr
}

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

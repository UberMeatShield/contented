package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Tag is used by pop to map your taggings database table to your go code.
type Tag struct {
	ID        string         `json:"id" db:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" db:"deleted_at"`

	Description string   `json:"description" db:"description" default:""`
	TagType     string   `json:"tag_type" db:"tag_type" default:"keyword"`
	Contents    Contents `json:"contents,omitempty" gorm:"many2many:contents_tags;"`
}

// String is not required by pop and may be deleted
func (t Tag) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

type TagsJsonSort func(i, j int) bool

var VALID_TAG_ORDERS = []string{"created_at", "updated_at", "id", "description", "tag_type"}

func GetTagSort(arr Tags, jsonFieldName string) TagsJsonSort {
	var theSort TagsJsonSort
	switch jsonFieldName {
	case "updated":
		theSort = func(i, j int) bool {
			return arr[i].UpdatedAt.Unix() < arr[j].UpdatedAt.Unix()
		}
	case "id":
		theSort = func(i, j int) bool {
			return arr[i].ID < arr[j].ID
		}
	case "tag_type":
		theSort = func(i, j int) bool {
			return arr[i].TagType < arr[j].TagType
		}
	case "description":
		theSort = func(i, j int) bool {
			return arr[i].Description < arr[j].Description
		}
	case "created_at":
		theSort = func(i, j int) bool {
			return arr[i].CreatedAt.Unix() < arr[j].CreatedAt.Unix()
		}
	default:
		theSort = func(i, j int) bool {
			return arr[i].ID < arr[j].ID
		}
	}
	return theSort
}

func GetTagsOrder(order string, direction string) string {
	return GetValidOrder(VALID_TAG_ORDERS, order, direction, "id")
}

// Tags is not required by pop and may be deleted
type Tags []Tag
type TagsMap map[string]Tag
type TagsCollection map[string]Tags

// String is not required by pop and may be deleted
func (t Tags) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

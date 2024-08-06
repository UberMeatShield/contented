package models

import (
	"testing"
)

func Test_Tag(t *testing.T) {
	tag := Tag{
		ID: "Test",
	}
	db := InitGorm(false)
	if db.Error != nil {
		t.Errorf("Could not connect to db %s", db.Error)
	}
	res := db.Create(&tag)
	if res.Error != nil {
		t.Errorf("Failed to create tag %s", res.Error)
	}
	tCheck := Tag{}
	db.First(&tCheck, "id = ?", tag.ID)
	if tCheck.ID == "" {
		t.Errorf("Failed to load tag %s", tCheck)
	}
}

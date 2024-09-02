package models

import (
	"testing"
)

func TestTag(t *testing.T) {
	db := InitGorm(false)
	NoError(db, "Could not connect to the db", t)

	SetupTests(db, t)
	tag := Tag{
		ID: "Test",
	}
	res := db.Create(&tag)
	NoError(res, "Failed to create tag", t)

	tCheck := Tag{}
	db.First(&tCheck, "id = ?", tag.ID)
	if tCheck.ID == "" {
		t.Errorf("Failed to load tag %s", tCheck)
	}
}

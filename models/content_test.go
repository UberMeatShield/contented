package models

import (
	"testing"
)

func TestContent(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	var count int64
	tx := db.Model(&Content{}).Count(&count)
	if tx.Error != nil {
		t.Errorf("Failed to count %s", tx.Error)
	}
	if count > 0 {
		t.Errorf("Test_Content There are still contents in the DB %d", count)
	}

	tags := Tags{
		Tag{ID: "A"},
		Tag{ID: "B"},
	}
	mc := Content{
		Src:         "This is the name of the item",
		Preview:     "preview_location",
		ContentType: "image/png",
		Tags:        tags,
	}
	// Tags MUST be created or the association will not be made
	for _, tag := range tags {
		tagTx := db.Create(&tag)
		if tagTx.Error != nil {
			t.Errorf("Not creating tag %s\n", tagTx.Error)
		}
	}
	mc.Tags = tags

	mcTx := db.Save(&mc)
	if mcTx.Error != nil {
		t.Errorf("Failed to save content %s", mcTx.Error)
	}
	if mc.ID <= 0 {
		t.Errorf("Did not actually save content")
	}

	// TODO: At least query the join table and see what comes back
	check := Content{}
	checkRes := db.Preload("Tags").Find(&check, mc.ID)
	if checkRes.Error != nil {
		t.Errorf("Error loading the content %s", checkRes.Error)
	}
	if len(check.Tags) != 2 {
		t.Errorf("Failed to get the correct tags back %d", len(check.Tags))
	}
}

func TestContentPagination(t *testing.T) {
	t.Errorf("IMPLEMENT AND THEN FIX DB Manager")
}

package models

import (
	"path/filepath"
	"testing"
)

func TestScreensContent(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	var count int64
	tx := db.Model(&Screen{}).Count(&count)
	NoError(tx, "Failed to count screens", t)

	if count > 0 {
		t.Errorf("There are still screens in the DB %d", count)
	}

	mc := Content{
		Src:         "We should be able to create a set of screens",
		Preview:     "preview_location",
		ContentType: "image/png",
	}
	mcTx := db.Create(&mc)
	NoError(mcTx, "Could not create content in screens test", t)
	if mc.ID == 0 {
		t.Errorf("Failed to save %d", mc.ID)
	}

	p1 := Screen{
		Src:       "fake1",
		Idx:       0,
		ContentID: mc.ID,
	}
	p2 := Screen{
		Src:       "fake2.png",
		Idx:       1,
		Path:      "Derp/Monkey",
		ContentID: mc.ID,
	}

	p2Loc := p2.GetFqPath()
	if p2Loc != filepath.Join(p2.Path, p2.Src) {
		t.Errorf("Didn't create the right fq path %s", p2Loc)
	}

	s1Tx := db.Create(&p1)
	s2Tx := db.Create(&p2)
	NoError(s1Tx, "Couldn't create preview screen 1", t)
	NoError(s2Tx, "Couldn't create preview screen 2", t)

	check := Content{}
	checkTx := db.Preload("Screens").Find(&check, mc.ID)
	NoError(checkTx, "Could not load back content", t)
	if check.Screens == nil {
		t.Errorf("Failed to load screens %d", mc.ID)
	}
	if len(check.Screens) != 2 {
		t.Errorf("The screens did not load back: %d", mc.ID)
	}
}

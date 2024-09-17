package models

import (
	"testing"
)

func TestContainerCreate(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	var count int64
	tx := db.Model(&Container{}).Count(&count)
	NoError(tx, "There are still containers in the DB", t)

	if count > 0 {
		t.Errorf("The DB was not reset %d", count)
	}

	c := Container{
		Total: 2,
		Path:  "root_path",
		Name:  "test_dir",
	}
	cTx := db.Create(&c)
	NoError(cTx, "Could not create the container", t)
	if c.ID == 0 {
		t.Errorf("Failed to create a container %d", c.ID)
	}
}

func TestContainerQuery(t *testing.T) {
	db := InitGorm(false)
	SetupTests(db, t)

	c := Container{
		Total: 2,
		Path:  "query_test",
		Name:  "KindaCrap",
	}
	cTx := db.Create(&c)
	NoError(cTx, "Could not create container", t)

	mc1 := Content{
		Src:         "first",
		ContainerID: &c.ID,
	}
	mc2 := Content{
		Src:         "second",
		ContainerID: &c.ID,
	}

	mc1Tx := db.Create(&mc1)
	NoError(mc1Tx, "Failed to create a content element", t)
	mc2Tx := db.Create(&mc2)
	NoError(mc2Tx, "Failed to create second content element", t)

	load_back := Container{}
	cLoadTx := db.Preload("Contents").Find(&load_back, c.ID)
	NoError(cLoadTx, "Could not load container", t)

	if len(load_back.Contents) != 2 {
		t.Errorf("Could not load up the contents content containers %s", load_back)
	}
}

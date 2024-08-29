package models

import (
	"log"
	"os"
	"testing"

	"gorm.io/gorm"
)

func NoError(tx *gorm.DB, msg string, t *testing.T) {
	if tx.Error != nil {
		t.Errorf("%s error: %s", msg, tx.Error)
	}
}

func SetupTests(db *gorm.DB, t *testing.T) {
	NoError(db.Exec("DELETE FROM contents_tags"), "Failed contents tags delete", t)
	NoError(db.Exec("DELETE FROM tags"), "Failed tags delete", t)
	NoError(db.Exec("DELETE FROM screens"), "Failed screens delete", t)
	NoError(db.Exec("DELETE FROM task_requests"), "Failed task request delete", t)
	NoError(db.Exec("DELETE FROM contents"), "Failed contents delete", t)
	NoError(db.Exec("DELETE FROM containers"), "Failed containers delete", t)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	if db := MigrateDb(InitGorm(true)); db.Error != nil {
		log.Fatalf("Could not init and migrate the db %s", db.Error)
	}
}

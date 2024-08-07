package models

import (
	"testing"

	"github.com/gobuffalo/suite/v4"
	"gorm.io/gorm"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	model := suite.NewModel()
	as := &ModelSuite{
		Model: model,
	}
	suite.Run(t, as)
}

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

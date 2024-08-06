package models

import (
	"log"

	"github.com/gobuffalo/pop/v6"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

/*
func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"
}
*/

/**
 * Build out a set of gorm related connections
 */
var GormDB *gorm.DB = nil

func InitGorm(reset bool) *gorm.DB {
	if GormDB == nil || reset {
		dsn := "host=localhost user=postgres dbname=content_test port=5432 sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err != nil {
			panic("Failed to connect database")
		}
		if db.Error == nil {
			log.Printf("Conected to the db")
			GormDB = db
		}
	}
	return GormDB
}

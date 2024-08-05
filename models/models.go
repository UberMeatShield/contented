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
var GormDB *gorm.DB

func InitGorm() {
	GormDB, err := gorm.Open(postgres.Open("db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	if GormDB != nil {
		log.Printf("Conected to the db")
	}
}

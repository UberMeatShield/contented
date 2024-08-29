package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// These need to be valid or we will bail the app
func GetEnvString(key string, defaultVal string) string {
	valStr := os.Getenv(key)
	if valStr != "" {
		return valStr
	}
	return defaultVal
}

/**
 * Build out a set of gorm related connections and ensure this is smarter about a close?
 */
var GormDB *gorm.DB = nil

// Need to get the envy version of this working properly
func InitGorm(reset bool) *gorm.DB {
	if GormDB == nil || reset {
		env := GetEnvString("GO_ENV", "development")
		host := GetEnvString("DB_HOST", "localhost")
		user := GetEnvString("DB_USER", "postgres")
		dbName := GetEnvString("DB_NAME", fmt.Sprintf("content_%s", env))
		password := GetEnvString("DB_PASSWORD", "")

		dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=5432 sslmode=disable", host, user, dbName, password)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err != nil {
			panic("Failed to connect database")
		}
		if db.Error == nil {
			log.Printf("Conected to the db")
			GormDB = db
		} else {
			log.Printf("Error in the DB Connection %s", db.Error)
		}
	}
	return GormDB
}

// Not sure if I can use this in a saner way wrapped by Gorm
func InitSqlPool(db *gorm.DB) *sql.DB {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to create a pool %s", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
	return sqlDB
}

func MigrateDb(db *gorm.DB) *gorm.DB {
	db.AutoMigrate(&Container{}, &Content{}, &Screen{}, &Tag{}, &TaskRequest{})
	return db
}

func ResetDB(db *gorm.DB) *gorm.DB {
	db.Exec("DELETE FROM contents_tags")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM screens")
	db.Exec("DELETE FROM task_requests")
	db.Exec("DELETE FROM contents")
	db.Exec("DELETE FROM containers")
	return db
}

func CheckReset(tx *gorm.DB) {
	if tx.Error != nil {
		log.Fatalf("Failed to execute request %s", tx.Error)
	}
}

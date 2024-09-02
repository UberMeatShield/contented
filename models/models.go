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
		host, user, dbName, password, port := GetDbConfig(env)
		dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s sslmode=disable", host, user, dbName, password, port)
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

func connectToDatabase(environment string) (*gorm.DB, string) {
	host, user, dbName, password, port := GetDbConfig(environment)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, user, password)
	DB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return DB, dbName
}
func createDatabase(db *gorm.DB, dbName string) error {
	// For some reason it doesn't like a bind in a create
	return db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
}

func dropDatabase(db *gorm.DB, dbName string) error {
	return db.Exec(fmt.Sprintf("DROP DATABASE %s", dbName)).Error
}

func databaseExists(db *gorm.DB, dbName string) (bool, error) {
	query := "SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE lower(datname) = lower(?)"
	var count int64
	if tx := db.Raw(query, dbName).Scan(&count); tx.Error != nil {
		log.Fatalf("failed to even determine the database exists eventually this should not be fatal %s", tx.Error)
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func GetDbConfig(environmentDefault string) (string, string, string, string, string) {
	dbName := GetEnvString("DB_NAME", fmt.Sprintf("content_%s", environmentDefault))
	host := GetEnvString("DB_HOST", "localhost")
	user := GetEnvString("DB_USER", "postgres")
	password := GetEnvString("DB_PASSWORD", "")
	port := GetEnvString("DB_PORT", "5432")

	return host, user, dbName, password, port
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

func InitializeAppDatabase(env string) {
	db, dbName := connectToDatabase(env)
	exists, _ := databaseExists(db, dbName)

	if !exists {
		if errCreate := createDatabase(db, dbName); errCreate != nil {
			log.Fatalf("Could not create the DB %s", errCreate)
		}
	}
	tx := InitGorm(true)
	if migrateTx := MigrateDb(tx); migrateTx.Error != nil {
		log.Fatalf("Could not init and migrate the db %s", migrateTx.Error)
	}
}

// Use for a clean reset but this is mostly for tests.
func RebuildDatabase(env string) {
	db, dbName := connectToDatabase(env)
	if db.Error != nil {
		log.Fatalf("Database connection failed, probably fatal. %s", db.Error)
	}
	exists, _ := databaseExists(db, dbName) // Currently fatal if
	if exists {
		if errDrop := dropDatabase(db, dbName); errDrop != nil {
			log.Printf("Did not drop database %s", errDrop)
		}
	}
	if errCreate := createDatabase(db, dbName); errCreate != nil {
		log.Fatalf("Could not create the DB %s", errCreate)
	}

	tx := InitGorm(true)
	if migrateTx := MigrateDb(tx); migrateTx.Error != nil {
		log.Fatalf("Could not init and migrate the db %s", migrateTx.Error)
	}
}

func ResetDB(db *gorm.DB) *gorm.DB {
	CheckReset(db.Exec("DELETE FROM contents_tags"))
	CheckReset(db.Exec("DELETE FROM tags"))
	CheckReset(db.Exec("DELETE FROM screens"))
	CheckReset(db.Exec("DELETE FROM task_requests"))
	CheckReset(db.Exec("DELETE FROM contents"))
	CheckReset(db.Exec("DELETE FROM containers"))
	return db
}

func CheckReset(tx *gorm.DB) {
	if tx.Error != nil {
		log.Fatalf("Failed to execute request %s", tx.Error)
	}
}

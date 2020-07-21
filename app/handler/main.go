package handler

import (
	"container/list"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"gitlab.com/ocvt/dolabra/utils"
)

var db *sql.DB
var emailQueue *list.List

// Initialize global variables
func Initialize() {
	log.SetFlags(log.Lshortfile)

	// Setup db
	config := utils.GetConfig()
	dbURI := fmt.Sprintf("./data/%s.sqlite3?_foreign_keys=1", config.DBName)
	var err error
	db, err = sql.Open("sqlite3", dbURI)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	utils.DBMigrate(db)

	// Generate cookie encryption key
	key = make([]byte, 512)
	_, err = rand.Read(key)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize email queue
	emailQueue = list.New()

	err = dbCreateSystemMember()
	if err != nil {
		log.Fatal("Error creating System Member: ", err)
	}

	err = dbCreateNullTrip()
	if err != nil {
		log.Fatal("Error creating null trip (for announcements): ", err)
	}
}

// Allow db to be closed from app package
func DBClose() error {
	return db.Close()
}

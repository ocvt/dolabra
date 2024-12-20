package handler

import (
	"container/list"
	crypto_rand "crypto/rand"
	"database/sql"
	"fmt"
	"log"
	math_rand "math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	//	"github.com/microcosm-cc/bluemonday"
	"github.com/ocvt/dolabra/utils"
)

var db *sql.DB
var emailQueue *list.List

//var strictHTML *bluemonday.Policy

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

	// Initialize HTML sanitizer
	//	strictHTML = bluemonday.UGCPolicy()

	// Seed math/rand, used for trip approval GUIDs
	math_rand.Seed(time.Now().UnixNano())

	// Generate cookie encryption key
	key = make([]byte, 512)
	_, err = crypto_rand.Read(key)
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

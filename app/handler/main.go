package handler

import (
	"container/list"
	crypto_rand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Setup db. WAL keeps readers from blocking the writer, the busy timeout
	// retries instead of surfacing SQLITE_BUSY (which we log.Fatal on), and a
	// single connection bounds memory and serializes writes.
	config := utils.GetConfig()
	dbURI := fmt.Sprintf("./data/%s.sqlite3?_foreign_keys=1&_journal_mode=WAL&_busy_timeout=5000", config.DBName)
	var err error
	db, err = sql.Open("sqlite3", dbURI)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	utils.DBMigrate(db)

	// Photo disk cache
	err = os.MkdirAll(PHOTO_CACHE_DIR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize HTML sanitizer
	//	strictHTML = bluemonday.UGCPolicy()

	// Load persistent signing key (JWTs + unsubscribe link HMACs), creating
	// it on first startup. Lives on the data volume so sessions and
	// unsubscribe links survive restarts.
	keyPath := "./data/secret.key"
	keyHex, err := os.ReadFile(keyPath)
	if err == nil {
		key, err = hex.DecodeString(strings.TrimSpace(string(keyHex)))
		if err != nil || len(key) < 32 {
			log.Fatal("Invalid signing key file: ", keyPath)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		key = make([]byte, 64)
		_, err = crypto_rand.Read(key)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600)
		if err != nil {
			log.Fatal(err)
		}
	} else {
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

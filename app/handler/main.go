package handler

import (
  "crypto/rand"
  "database/sql"
  "fmt"
  "log"

  _ "github.com/mattn/go-sqlite3"

  "gitlab.com/ocvt/dolabra/utils"
)

var db *sql.DB

// Initialize global variables
func Initialize() {
  log.SetFlags(log.Lshortfile)

  // Setup db
  config := utils.GetConfig()
  dbURI := fmt.Sprintf("./data/%s.sqlite3", config.DBName)
  var err error
  db, err = sql.Open("sqlite3", dbURI)
  if err != nil {
    log.Fatal("Error opening database: ", err)
  }
  err = db.Ping()
  if err != nil{
    log.Fatal(err)
  }
  utils.DBMigrate(db)

  // Generate cookie encryption key
  _, err = rand.Read(key[:])
  if err != nil {
    log.Fatal(err)
  }
}

// Allow db to beclosed from app package
func DBClose() error {
  return db.Close()
}

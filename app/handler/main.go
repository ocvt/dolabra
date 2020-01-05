package handler

import (
  "crypto/rand"
  "database/sql"
  "fmt"
  "log"
  "os"

  _ "github.com/mattn/go-sqlite3"

  "gitlab.com/ocvt/dolabra/utils"
)

var db *sql.DB

// Initialize global variables
func Initialize() {
  log.SetFlags(log.Lshortfile)

  if len(os.Getenv("DEV")) > 0 {
    googleOAuthConfig.RedirectURL = "http://localhost:3000/auth/google/callback"
  }

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

  // Load envs
  TRIPS_FOLDER_ID = os.Getenv("GDRIVE_TRIPS_FOLDER_ID")
  HOME_PHOTOS_FOLDER_ID = os.Getenv("GDRIVE_HOME_PHOTOS_FOLDER_ID")
  SMTP_USERNAME = os.Getenv("SMTP_USERNAME")
  SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")
  SMTP_HOSTNAME = os.Getenv("SMTP_HOSTNAME")
  SMTP_PORT = os.Getenv("SMTP_PORT")
  SMTP_FROM_FIRST_NAME_DEFAULT = os.Getenv("SMTP_FROM_FIRST_NAME_DEFAULT")
  SMTP_FROM_LAST_NAME_DEFAULT = os.Getenv("SMTP_FROM_LAST_NAME_DEFAULT")
  SMTP_FROM_EMAIL_DEFAULT = os.Getenv("SMTP_FROM_EMAIL_DEFAULT")
  STRIPE_SECRET_KEY = os.Getenv("STRIPE_SECRET_KEY")

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

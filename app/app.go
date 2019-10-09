package app

import (
//  "fmt"
  "log"
  "net/http"

  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
//  _ "github.com/lib/pq"

  "gitlab.com/ocvt/api/app/handler"
  "gitlab.com/ocvt/api/config"
)

var r *chi.Mux
//var db *sql.DB

// Initialize with predefined configuration
func Initialize(config *config.Config) {
//  dbURI := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True",
//    config.DB.Username,
//    config.DB.Password,
//    config.DB.Host,
//    config.DB.Port,
//    config.DB.Name,
//    config.DB.Charset)
//
//  db, err := gorm.Open(config.DB.Dialect, dbURI)
//  if err != nil {
//    log.Fatal("err")
//  }
//
//  db = model.DBMigrate(db)
  r = chi.NewRouter()
  setRouters()
}

// Set routes
func setRouters() {
  // Set middleware
  r.Use(middleware.RequestID)
  r.Use(middleware.RealIP)
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)
  r.Use(handler.ProcessClientAuth)

  r.Route("/myaccount", func(r chi.Router) {
    //r.Get("/", handler.GetMyAccount)
    r.Get("/summary", handler.GetMyAccountSummary)
    //r.Post("/register", handler.PostMyAccountRegister)
  })
}

func Run(host string) {
  log.Fatal(http.ListenAndServe(host, r))
}

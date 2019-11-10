package app

import (
  "context"
  "log"
  "net/http"
  "os"
  "os/signal"
  "time"

  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/rs/cors"

  "gitlab.com/ocvt/api/app/handler"
)

var r *chi.Mux

// Initialize api
func Initialize() {
  log.SetFlags(log.Lshortfile)

  // Setup handlers
  handler.Initialize()

  // Finally, configure routes
  r = chi.NewRouter()
  setRouters()
}

// Set routes
func setRouters() {
  // TODO configure CORS
  // Set middleware
  r.Use(cors.Default().Handler)
  r.Use(middleware.RequestID)
  r.Use(middleware.RealIP)
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)
  // Process JWT is present
  r.Use(handler.ProcessClientAuth)

  r.Route("/auth", func(r chi.Router) {
    r.Get("/google", handler.GoogleLogin)
    r.Get("/google/callback", handler.GoogleCallback)
  })

  r.Route("/myaccount", func(r chi.Router) {
    r.Delete("/delete", handler.DeleteMyAccountDelete)
    r.Get("/", handler.GetMyAccount)
    r.Get("/name", handler.GetMyAccountName)
    r.Patch("/deactivate", handler.PatchMyAccountDeactivate)
    r.Patch("/reactivate", handler.PatchMyAccountReactivate)
    r.Post("/register", handler.PostMyAccountRegister)
  })
}

func Run(host string) {
  server := &http.Server{Addr: host, Handler: r}

  // Start server in separate goroutine
  go func() {
    log.Printf("Server starting on %s", host)
    log.Fatal(server.ListenAndServe())
  }()

  // Wait for SIGINT
  stop := make(chan os.Signal, 1)
  signal.Notify(stop, os.Interrupt)
  <-stop

  // Stop any remaining tasks
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  // Shutdown server
  err := server.Shutdown(ctx)
  if err != nil {
    log.Printf("Server failed to shutdown: %s", err.Error())
  } else  {
    log.Printf("Server successfully shutdown")
  }

  // Close DB
  err = handler.DBClose()
  if err != nil {
    log.Printf("DB failed to close: %s", err.Error())
  } else {
    log.Printf("DB successfully closed")
  }
}

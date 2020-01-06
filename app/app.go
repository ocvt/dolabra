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

  "gitlab.com/ocvt/dolabra/app/handler"
)

var r *chi.Mux

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

  r.Get("/homephotos", handler.GetHomePhotos)
  r.Get("/news", handler.GetNews)
  r.Get("/newsArchive", handler.GetNewsArchive)
  r.Get("/payment/{paymentOption}", handler.GetPayment)
  r.Post("/payment/paymentSucceeded", handler.PostPaymentSucceeded)
  r.Post("/quicksignup", handler.PostQuicksignup)
  r.Post("/unsubscribe/all", handler.PostUnsubscribeAll)

  r.Route("/auth", func(r chi.Router) {
    r.Get("/google", handler.GoogleLogin)
    r.Get("/google/callback", handler.GoogleCallback)
    if len(os.Getenv("DEV")) > 0 {
      r.Get("/dev/{subject}", handler.DevLogin)
    }
  })

  r.Route("/myaccount", func(r chi.Router) {
    r.Delete("/", handler.DeleteMyAccountDelete)
    r.Get("/", handler.GetMyAccount)
    r.Get("/name", handler.GetMyAccountName)
    r.Get("/notifications", handler.GetMyAccountNotifications)
    r.Patch("/deactivate", handler.PatchMyAccountDeactivate)
    r.Patch("/notifications", handler.PatchMyAccountNotifications)
    r.Patch("/reactivate", handler.PatchMyAccountReactivate)
    r.Post("/", handler.PostMyAccount)
  })

  r.Route("/trips", func(r chi.Router) {
    r.Get("/", handler.GetTrips)
    r.Get("/archive", handler.GetTripsArchive)
    r.Get("/archive/*", handler.GetTripsArchive)
    r.Get("/mytrips", handler.GetTripsMyTrips)
    r.Get("/photos", handler.GetAllTripsPhotos)
    r.Get("/types", handler.GetTripsTypes)
    r.Get("/{tripId}/admin", handler.GetTripsAdmin)
    r.Get("/{tripId}/photos", handler.GetTripsPhotos)
    r.Patch("/{tripId}/cancel", handler.PatchTripsCancel)
    r.Patch("/{tripId}/publish", handler.PatchTripsPublish)
    r.Post("/{tripId}/notify/signup/{signupId}", handler.PostTripsNotifySignup)
    r.Post("/{tripId}/notify/{groupId}", handler.PostTripsNotifyGroup)
    r.Post("/", handler.PostTrips)
    r.Post("/{tripId}/mainphoto", handler.PostTripsMainphoto)
    r.Post("/{tripId}/photos", handler.PostTripsPhotos)
    r.Route("/{tripId}/signup", func(r chi.Router) {
      r.Get("/", handler.GetTripsSignup)
      r.Patch("/cancel", handler.PatchTripsSignupCancel)
      r.Patch("/{signupId}/absent", handler.PatchTripsSignupAbsent)
      r.Patch("/{signupId}/boot", handler.PatchTripsSignupBoot)
      r.Patch("/{signupId}/forceadd", handler.PatchTripsSignupForceadd)
      r.Patch("/{signupId}/tripleader/{promote}", handler.PatchTripsSignupTripLeaderPromote)
      r.Post("/", handler.PostTripsSignup)
    })
  })

  r.Route("/webtools", func(r chi.Router) {
    r.Use(handler.EnsureOfficer)
    r.Delete("/news/{tripId}", handler.DeleteWebtoolsNews)
    r.Get("/emails", handler.GetWebtoolsEmails)
    r.Route("/members", func(r chi.Router) {
      r.Get("/", handler.GetWebtoolsMembers)
      r.Get("/{memberId}/attendance", handler.GetWebtoolsMembersAttendance)
      r.Get("/{memberId}/trips", handler.GetWebtoolsMembersTrips)
      r.Patch("/{memberId}/dues/grant", handler.PatchWebtoolsDuesGrant)
      r.Patch("/{memberId}/dues/revoke", handler.PatchWebtoolsDuesRevoke)
    })
    r.Get("/officers", handler.GetWebtoolsOfficers)
    r.Get("/payments", handler.GetWebtoolsPayments)
    r.Delete("/officers/{memberId}", handler.DeleteWebtoolsOfficers)
    r.Patch("/officers/{memberId}/{action}", handler.PatchWebtoolsOfficers)
    r.Post("/news", handler.PostWebtoolsNews)
    r.Post("/officers/", handler.PostWebtoolsOfficers)
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

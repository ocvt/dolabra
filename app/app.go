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
	r.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowedMethods:   []string{"DELETE", "GET", "PATCH", "POST"},
		AllowedOrigins:   []string{"http://cabinet.seaturtle.pw:4000"},
		Debug:            true,
	}).Handler)
	//r.Use(c.Handler)
	//r.Use(cors.Default().Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/homephotos", handler.GetHomePhotos)
	r.Get("/news", handler.GetNews)
	r.Get("/newsArchive", handler.GetNewsArchive)
	r.Post("/quicksignup", handler.PostQuicksignup)
	r.Post("/unsubscribe/all", handler.PostUnsubscribeAll)
	r.Route("/noauth", func(r chi.Router) {
		r.Get("/trips", handler.GetTripsSummary)
		r.Get("/trips/{tripId}", handler.GetTripSummary)
		r.Get("/trips/archive", handler.GetTripsArchiveDefault)
		r.Get("/trips/archive/{startId}/{perPage}", handler.GetTripsArchive)
		r.Get("/trips/photos", handler.GetAllTripsPhotos)
		r.Get("/trips/{tripId}/photos", handler.GetTripsPhotos)
		r.Get("/trips/types", handler.GetTripsTypes)
	})

	r.Route("/payment", func(r chi.Router) {
		r.Use(handler.ProcessClientAuth)
		r.Get("/{paymentOption}", handler.GetPayment)
		r.Post("/redeem", handler.PostPaymentRedeem)
		r.Post("/paymentSucceeded", handler.PostPaymentSucceeded)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/google", handler.GoogleLogin)
		r.Get("/google/callback", handler.GoogleCallback)
		r.Get("/logout", handler.GetLogout)
		if len(os.Getenv("DEV")) > 0 {
			r.Get("/dev/{sub}", handler.DevLogin)
		}
	})

	r.Route("/myaccount", func(r chi.Router) {
		r.Use(handler.ProcessClientAuth)
		r.Delete("/", handler.DeleteMyAccountDelete)
		r.Get("/", handler.GetMyAccount)
		r.Get("/name", handler.GetMyAccountName)
		r.Get("/notifications", handler.GetMyAccountNotifications)
		r.Patch("/", handler.PatchMyAccount)
		r.Patch("/deactivate", handler.PatchMyAccountDeactivate)
		r.Patch("/emergency", handler.PatchMyAccountEmergency)
		r.Patch("/notifications", handler.PatchMyAccountNotifications)
		r.Patch("/reactivate", handler.PatchMyAccountReactivate)
		r.Post("/", handler.PostMyAccount)
	})

	r.Route("/trips", func(r chi.Router) {
		r.Use(handler.ProcessClientAuth)
		r.Get("/{tripId}", handler.GetTrip)
		r.Get("/mytrips", handler.GetTripsMyTrips)
		//		r.Get("/{tripId}/mystatus", handler.GetTripMyStatus) // TODO may use in the future to get statuses from trip overview page
		r.Get("/{tripId}/admin", handler.GetTripsAdmin)
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
		r.Use(handler.ProcessClientAuth)
		r.Use(handler.EnsureOfficer)
		r.Delete("/news/{tripId}", handler.DeleteWebtoolsNews)
		r.Delete("/officers/{memberId}", handler.DeleteWebtoolsOfficers)
		r.Get("/emails", handler.GetWebtoolsEmails)
		r.Get("/equipment", handler.GetWebtoolsEquipment)
		r.Get("/officers", handler.GetWebtoolsOfficers)
		r.Get("/payments", handler.GetWebtoolsPayments)
		r.Patch("/equipment/{equipmentId}/{count}", handler.PatchWebtoolsEquipment)
		r.Patch("/officers/{memberId}/{action}", handler.PatchWebtoolsOfficers)
		r.Patch("/payments/{paymentRowId}/completed", handler.PatchWebtoolsPaymentsCompleted)
		r.Post("/equipment", handler.PostWebtoolsEquipment)
		r.Post("/payments", handler.PostWebtoolsPayments)
		r.Post("/payments/generateCode", handler.PostWebtoolsGenerateCode)
		r.Post("/news", handler.PostWebtoolsNews)
		r.Post("/officers/", handler.PostWebtoolsOfficers)
		r.Route("/members", func(r chi.Router) {
			r.Get("/", handler.GetWebtoolsMembers)
			r.Get("/{memberId}/attendance", handler.GetWebtoolsMembersAttendance)
			r.Get("/{memberId}/trips", handler.GetWebtoolsMembersTrips)
			r.Patch("/{memberId}/dues/grant", handler.PatchWebtoolsDuesGrant)
			r.Patch("/{memberId}/dues/revoke", handler.PatchWebtoolsDuesRevoke)
		})
	})
}

func Run(host string) {
	server := &http.Server{Addr: host, Handler: r}

	// Start server in separate goroutine
	go func() {
		log.Printf("Server starting on %s", host)
		log.Fatal(server.ListenAndServe())
	}()

	// Run tasks every 5 minutes
	//	ticker := time.NewTicker(5 * time.Minute)
	//	tickerQuit := make(chan struct{})
	//	go func() {
	//		log.Printf("Task ticker started, running every 5 minutes")
	//		for {
	//			select {
	//			case <-ticker.C:
	//				handler.DoTasks()
	//			case <-tickerQuit:
	//				log.Printf("Task ticker stopped")
	//				ticker.Stop()
	//				return
	//			}
	//		}
	//	}()

	// Wait for SIGINT
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	// Stop any remaining tasks
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown ticker
	//	close(tickerQuit)

	// Shutdown server
	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Server failed to shutdown: %s", err.Error())
	} else {
		log.Printf("Server successfully shutdown")
	}

	// Close DB TODO use transactions
	err = handler.DBClose()
	if err != nil {
		log.Printf("DB failed to close: %s", err.Error())
	} else {
		log.Printf("DB successfully closed")
	}
}

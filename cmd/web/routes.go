package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/handlers"
	"net/http"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Middleware Setup
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Routing
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	// Fileserver to go get static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Protected routes.
	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		// src highlights whether or not the users comes from the all or new reservations part of the layout.
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)

		mux.Get("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}", handlers.Repo.AdminDeleteReservation)

	})

	return mux
}

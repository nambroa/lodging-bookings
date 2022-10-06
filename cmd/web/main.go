package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/handlers"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"log"
	"net/http"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting application on port", portNumber)
	// Start a webserver and listen to a specific port.
	serve := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	// Types that will be stored in the session object (encoded in the session object).
	gob.Register(models.Reservation{})

	// Change this to true when in production.
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)
	return nil
}

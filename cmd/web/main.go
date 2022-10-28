package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/driver"
	"github.com/nambroa/lodging-bookings/internal/handlers"
	"github.com/nambroa/lodging-bookings/internal/helpers"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"log"
	"net/http"
	"os"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	// DB connection and mail channel cannot be closed in run() function since that function only runs once at the
	// beginning of the app. This means that the email channel would be created and immediately closed afterwards,
	// instead of being alive for the duration of the application.
	defer db.SQL.Close()
	defer close(app.Mailchan)
	listenForMail()
	fmt.Println("Starting application on port", portNumber)
	// Start a webserver and listen to a specific port.
	serve := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// Types that will be stored in the session object (encoded in the session object).
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	// Create channel to send emails.
	log.Printf("Start email listener..")
	mailChan := make(chan models.MailData)
	app.Mailchan = mailChan
	// Change this to true when in production.
	app.InProduction = false

	// Adding logs to the app config.
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Creating session info and adding it to the app config.
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// Connect to DB.
	log.Println("Connecting to database..")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=lodging-bookings user=postgres password=123")
	if err != nil {
		log.Fatal("Cannot connect to database. Error:", err)
	}
	log.Println("Connection to the database was successful.")

	// Create template cache.
	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Println(err.Error())
		log.Fatal("cannot create template cache")
	}
	app.TemplateCache = templateCache
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)
	return db, nil
}

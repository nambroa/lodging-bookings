package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var app config.AppConfig
var session *scs.SessionManager
var functions = template.FuncMap{}
var pathToTemplates = "../../templates" // Changed from base definition since tests are executed in a different package.

func TestMain(m *testing.M) {
	// Types that will be stored in the session object (encoded in the session object).
	gob.Register(models.Reservation{})

	// Change this to true when in production.
	app.InProduction = false

	// Adding logs to the app config.
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	templateCache, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = true // otherwise it will create a templateCache on every run overriding the template path variable.

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {

	mux := chi.NewRouter()

	// Middleware Setup
	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	// Routing
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/contact", Repo.Contact)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	// Fileserver to go get static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux

}

// SessionLoad loads and saves the session on every request.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template, error) {

	templateCache := map[string]*template.Template{} // Empty map, same as the make function.

	// Get all the files name *.fileName.gohtml from ./templates (FULL PATH from root project)
	// Ex: templates/home.page.gohtml
	fileNames, err := filepath.Glob(fmt.Sprintf("%s/*.page.gohtml", pathToTemplates))
	if err != nil {
		return templateCache, err
	}

	// Range through the received files.
	for _, fileName := range fileNames {
		name := filepath.Base(fileName) // Filename MINUS the full path. Ex home.fileName.gohtml.

		// Create template from the current fileName being iterated by the loop.
		// The line below creates a template with the name I want (home.filename.gohtml for example).
		// And it also populates it with the corresponding file.
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(fileName) // Page is the FULL PATH.
		if err != nil {
			return templateCache, err
		}

		// Search for layouts, that were not searched in the initial filepath.Glob and need to be applied in
		// each template (since they are a layout).
		fileNames, err := filepath.Glob(fmt.Sprintf("%s/*layout.gohtml", pathToTemplates))
		if err != nil {
			return templateCache, err
		}
		// If any layout fileNames were found, add the layout to the existing templateSet. So it mixes layout + page.
		if len(fileNames) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return templateCache, err
			}
		}

		templateCache[name] = templateSet
	}
	return templateCache, nil
}

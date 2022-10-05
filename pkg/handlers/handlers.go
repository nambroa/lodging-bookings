package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nambroa/lodging-bookings/pkg/config"
	"github.com/nambroa/lodging-bookings/pkg/models"
	"github.com/nambroa/lodging-bookings/pkg/render"
	"log"
	"net/http"
)

// Repository pattern used to share the appConfig with the handlers.

// Repo is used by the handlers.
var Repo *Repository

// Repository is the repository type.
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository with the content being the app config instance.
func NewRepo(appConfig *config.AppConfig) *Repository {
	return &Repository{
		App: appConfig,
	}
}

// NewHandlers sets the Repository for the handlers.
func NewHandlers(repo *Repository) {
	Repo = repo
}

// Home is the home page handler.
// By adding repo as a receiver, the Home handler has access to all the repository's content.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.gohtml", &models.TemplateData{})
}

// About is the about page handler.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, r, "about.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

// Majors is the Majors Suite page handler. Renders the room page.
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.gohtml", &models.TemplateData{})
}

// Generals is the Generals Quarters page handler. Renders the room page.
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.gohtml", &models.TemplateData{})
}

// Contact is the contact page handler.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.gohtml", &models.TemplateData{})
}

// Availability is the search availability page handler.
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.gohtml", &models.TemplateData{})
}

// PostAvailability is the search availability form handler.
// PostAvailability renders the search availability page after the form has been processed.
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start") // matches input from form with name attribute start. Default type is string.
	end := r.Form.Get("end")
	fmt.Println(start, end)
	w.Write([]byte("Posted to search availability"))
}

type jsonResponse struct {
	OK      bool   `json:"ok"` // `json` part means the name of the attribute in json.
	Message string `json:"message"`
}

// AvailabilityJSON is the search availability form handler. It sends back a JSON response.
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	response := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(response, "", "     ")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

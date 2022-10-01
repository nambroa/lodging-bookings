package handlers

import (
	"github.com/nambroa/lodging-bookings/pkg/config"
	"github.com/nambroa/lodging-bookings/pkg/models"
	"github.com/nambroa/lodging-bookings/pkg/render"
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
	render.RenderTemplate(w, "home.page.gohtml", &models.TemplateData{})
}

// About is the about page handler.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, "about.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

// Majors is the Majors Suite page handler. Renders the room page.
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, "majors.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

// Generals is the Generals Quarters page handler. Renders the room page.
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, "generals.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

// Availability is the search availability page handler.
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, "search-availability.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

// Contact is the contact page handler.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello, Again."}
	render.RenderTemplate(w, "contact.page.gohtml", &models.TemplateData{StringMap: stringMap})
}

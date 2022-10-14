package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/driver"
	"github.com/nambroa/lodging-bookings/internal/forms"
	"github.com/nambroa/lodging-bookings/internal/helpers"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"github.com/nambroa/lodging-bookings/internal/repository"
	"github.com/nambroa/lodging-bookings/internal/repository/dbrepo"
	"net/http"
)

// Repository pattern used to share the appConfig and the DB with the handlers.
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// Repo is used by the handlers.
var Repo *Repository

// NewRepo creates a new repository with the content being the app config instance.
func NewRepo(appConfig *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: appConfig,
		DB:  dbrepo.NewPostgresRepo(db.SQL, appConfig),
	}
}

// NewHandlers sets the Repository for the handlers.
func NewHandlers(repo *Repository) {
	Repo = repo
}

// Home is the home page handler.
// By adding repo as a receiver, the Home handler has access to all the repository's content.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.gohtml", &models.TemplateData{})
}

// About is the about page handler.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.gohtml", &models.TemplateData{})
}

// Majors is the Majors Suite page handler. Renders the room page.
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.gohtml", &models.TemplateData{})
}

// Generals is the Generals Quarters page handler. Renders the room page.
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.gohtml", &models.TemplateData{})
}

// Contact is the contact page handler.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.gohtml", &models.TemplateData{})
}

// Reservation is the Make Reservation page handler.
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := map[string]interface{}{"reservation": emptyReservation} // must be same key as reservation is called in PostReservation.

	render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form.
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	// If form is invalid, repopulate the fields of the reservation so the user only needs to type the errored fields.
	if !form.Valid() {
		data := map[string]interface{}{"reservation": reservation}

		render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Put the reservation info into the session to show later in the summary.
	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// Take reservation info from the session.
	// Last part of the line is type assertion aka casting, since Get method returns interface.
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	// Check if reservation was obtained properly. Otherwise display error to the user.
	if !ok {
		m.App.ErrorLog.Println("Can't get reservation from session")
		// We put an error message to be extracted in render method AddDefaultData to show to the user.
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Store reservation in TemplateData and delete it from the session to improve privacy.
	m.App.Session.Remove(r.Context(), "reservation")
	data := map[string]interface{}{"reservation": reservation}
	render.Template(w, r, "reservation-summary.page.gohtml", &models.TemplateData{Data: data})

}

// Availability is the search availability page handler.
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.gohtml", &models.TemplateData{})
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
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

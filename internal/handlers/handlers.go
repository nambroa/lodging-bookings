package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/driver"
	"github.com/nambroa/lodging-bookings/internal/forms"
	"github.com/nambroa/lodging-bookings/internal/helpers"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"github.com/nambroa/lodging-bookings/internal/repository"
	"github.com/nambroa/lodging-bookings/internal/repository/dbrepo"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("error obtaining reservation from session while choosing a room"))
		return
	}

	room, err := m.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	reservation.Room.RoomName = room.RoomName

	startDateFormatted := reservation.StartDate.Format("2006-01-02") // Formats time.Time in that specific layout.
	endDateFormatted := reservation.EndDate.Format("2006-01-02")     // Formats time.Time in that specific layout.

	data := map[string]interface{}{"reservation": reservation} // must be same key as reservation is called in PostReservation.
	stringMap := map[string]string{"start_date": startDateFormatted, "end_date": endDateFormatted}
	render.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func parseDateFromForm(form url.Values, dateString string) (time.Time, error) {
	// Declare the layout that matches how the date is extracted from the form
	layout := "2006-01-02"
	// Parse it in go-friendly date.
	// Example form date: 2006-01-02
	// Parsed date: 2016-01-02 0:00:00 +0000 UTC
	parsedDate, err := time.Parse(layout, form.Get(dateString))
	return parsedDate, err
}

// PostReservation handles the posting of a reservation form.
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	startDate, err := parseDateFromForm(r.Form, "start_date")
	endDate, err := parseDateFromForm(r.Form, "end_date")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id")) // Parses string to int.
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
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

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // Temporary, RestrictionID 1 = Restriction of type reservation
	}
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
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
	// matches input from form with name attribute start. Default type is string
	startDate, err := parseDateFromForm(r.Form, "start")
	endDate, err := parseDateFromForm(r.Form, "end")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability for selected dates. Please select another date.")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	// Storing info in the session to pass to the make reservation page later.
	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	m.App.Session.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.gohtml", &models.TemplateData{Data: data})
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

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id")) // key is the same as the key in routes.go
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("error obtaining reservation from session while choosing a room"))
		return
	}

	reservation.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", reservation)

	// Redirect user to make a reservation for the room they chose.
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/driver"
	"github.com/nambroa/lodging-bookings/internal/forms"
	"github.com/nambroa/lodging-bookings/internal/helpers"
	"github.com/nambroa/lodging-bookings/internal/models"
	"github.com/nambroa/lodging-bookings/internal/render"
	"github.com/nambroa/lodging-bookings/internal/repository"
	"github.com/nambroa/lodging-bookings/internal/repository/dbrepo"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// NewTestRepo creates a new repository without a DB connection for the unit tests.
func NewTestRepo(appConfig *config.AppConfig) *Repository {
	return &Repository{
		App: appConfig,
		DB:  dbrepo.NewTestingRepo(appConfig),
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
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reservation.Room.RoomName = room.RoomName
	m.App.Session.Put(r.Context(), "reservation", reservation) // Adding reservation with room info to the session.
	// Parsing go date format to user readable in order to show it in the reservation html.
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

// PostReservation handles the posting of a reservation form.
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// Get reservation from session context
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session for PostReservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form for PostReservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Update reservation stored in the session.
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

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
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into DB for PostReservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // Temporary, RestrictionID 1 = Restriction of type reservation
	}
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction into DB for PostReservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// Send email notification to the guest.
	htmlMessage := fmt.Sprintf(`
		<strong> Reservation Confirmation </strong><br>
		Dear %s: <br>
		Your reservation for the %s from the %s to the %s is now confirmed.
`, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:      reservation.Email,
		From:    "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}
	m.App.Mailchan <- msg

	// Send email notification to the owner.
	htmlMessage = fmt.Sprintf(`
		<strong> Reservation Confirmation </strong><br>
		Dear %s: <br>
		A reservation has been made for your property %s from the %s to the %s.
`, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:      "owner-email@here.com",
		From:    "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}
	m.App.Mailchan <- msg

	// Put the reservation info into the session to show later in the summary.
	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// ReservationSummary displays information about the reservation after confirming it.
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

	// Create maps to pass it to the html.
	data := map[string]interface{}{"reservation": reservation}
	stringMap := map[string]string{"start_date": reservation.StartDate.Format("2006-01-02"),
		"end_date": reservation.EndDate.Format("2006-01-02")}
	render.Template(w, r, "reservation-summary.page.gohtml", &models.TemplateData{Data: data, StringMap: stringMap})

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
	OK        bool   `json:"ok"` // `json` part means the name of the attribute in json.
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON is the search availability form handler. It sends back a JSON response.
// Used for the popup modal to search availability for a specific room.
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm() // Not needed for functionality, but needed for tests and its good practice.
	if err != nil {
		resp := jsonResponse{OK: false, Message: "Internal server error"}
		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	startDate, _ := parseDateFromForm(r.Form, "start")
	endDate, _ := parseDateFromForm(r.Form, "end")
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	response := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: r.Form.Get("start"),
		EndDate:   r.Form.Get("end"),
		RoomID:    strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(response, "", "     ")
	if err != nil {
		resp := jsonResponse{OK: false, Message: "Internal server error"}
		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// ChooseRoom displays a list of rooms that the user can make a reservation of.
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

// BookRoom takes URL params, builds a reservation in the session and takes the user to the make reservation page.
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// Get room ID, Reservation start and end date from url params.
	roomId, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	var reservation models.Reservation
	reservation.RoomID = roomId
	parsedStartDate, _ := time.Parse("2006-01-02", startDate)
	parsedEndDate, _ := time.Parse("2006-01-02", endDate)
	reservation.StartDate = parsedStartDate
	reservation.EndDate = parsedEndDate
	room, err := m.DB.GetRoomByID(roomId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	reservation.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) ShowLogin(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "login.page.gohtml", &models.TemplateData{Form: forms.New(nil)})
}

// PostShowLogin handles the user logging in.
func (m *Repository) PostShowLogin(writer http.ResponseWriter, request *http.Request) {
	// Prevents session fixation attack.
	_ = m.App.Session.RenewToken(request.Context())

	err := request.ParseForm()
	if err != nil {
		log.Println("Cannot parse postshowlogin form:", err)
		m.App.Session.Put(request.Context(), "error", "Cannot parse user information")
		http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
		return
	}

	form := forms.New(request.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		log.Println("Postshowlogin form is not valid:", form.Errors)
		render.Template(writer, request, "login.page.gohtml", &models.TemplateData{Form: form})
		return
	}

	id, _, err := m.DB.Authenticate(request.Form.Get("email"), request.Form.Get("password"))
	if err != nil {
		log.Println("Cannot authenticate user in postshowlogin:", err)
		m.App.Session.Put(request.Context(), "error", "Invalid login credentials")
		http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(request.Context(), "user_id", id)
	m.App.Session.Put(request.Context(), "flash", "Logged in successfully")
	http.Redirect(writer, request, "/", http.StatusSeeOther)

}

// Logout logs the user out.
func (m *Repository) Logout(writer http.ResponseWriter, request *http.Request) {
	// Destroy the user session.
	_ = m.App.Session.Destroy(request.Context())
	// Prevents session fixation attack.
	_ = m.App.Session.RenewToken(request.Context())
	m.App.Session.Put(request.Context(), "flash", "Logged out successfully")
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(writer http.ResponseWriter, request *http.Request) {
	render.Template(writer, request, "admin-dashboard.page.gohtml", &models.TemplateData{})
}

// AdminNewReservations shows all the new reservations in the admin layout.
func (m *Repository) AdminNewReservations(writer http.ResponseWriter, request *http.Request) {
	reservations, err := m.DB.GetNewReservations()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	data := map[string]interface{}{"reservations": reservations}

	render.Template(writer, request, "admin-new-reservations.page.gohtml", &models.TemplateData{Data: data})
}

// AdminAllReservations shows all the reservations in the admin layout.
func (m *Repository) AdminAllReservations(writer http.ResponseWriter, request *http.Request) {
	reservations, err := m.DB.GetAllReservations()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	data := map[string]interface{}{"reservations": reservations}
	render.Template(writer, request, "admin-all-reservations.page.gohtml", &models.TemplateData{Data: data})
}

// AdminReservationsCalendar displays the reservation calendar in the admin layout.
func (m *Repository) AdminReservationsCalendar(writer http.ResponseWriter, request *http.Request) {
	// Assume that there is no month or year specified.
	now := time.Now()

	// If the user specified a year and a month, convert the calendar to use those instead of the current year/month.
	if request.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(request.URL.Query().Get("y"))
		month, _ := strconv.Atoi(request.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}
	data := map[string]interface{}{"now": now}

	// Calculate next year/month and previous year/month. These are used to navigate the calendar.
	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")
	stringMap := map[string]string{"next_month": nextMonth, "next_month_year": nextMonthYear, "last_month": lastMonth,
		"last_month_year": lastMonthYear, "this_month": now.Format("01"), "this_month_year": now.Format("2006")}

	// Get the first and last days of the month.
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// Store number of days in a month.
	intMap := map[string]int{"days_in_month": lastOfMonth.Day()}

	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	data["rooms"] = rooms

	// Create reservation map and block map to be used in the calendar. For each day in the map marked as 0, it means
	// the day is available for reservation (and will not appear checked on the calendar).
	for _, room := range rooms {
		reservationMap := map[string]int{}
		blockMap := map[string]int{}
		// Iterate the whole month, one day at a time. Initialize the map as fully available.
		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0 // Room available
			blockMap[d.Format("2006-01-2")] = 0
		}
		// Get the restrictions for the rooms for the current month and add them to the map.
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(room.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(writer, err)
			return
		}
		// If the restriction is from a reservation, add it to the reservation map. Otherwise add it to the block map.
		for _, restriction := range restrictions {
			if restriction.ReservationID > 0 {
				// It's a reservation
				for d := restriction.StartDate; d.After(restriction.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = restriction.ReservationID
				}
			} else {
				// It's a block from the owner.
				blockMap[restriction.StartDate.Format("2006-01-2")] = restriction.ID
			}
		}
		// Add it to the data map in order to pass it on to the template.
		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap
		// Store the blockMap in the session, so that when the user makes changes in the calendar and posts
		// that form, I can take this block map out of the session and compare it to the changes made by the user.
		m.App.Session.Put(request.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.Template(writer, request, "admin-reservations-calendar.page.gohtml", &models.TemplateData{StringMap: stringMap,
		Data: data, IntMap: intMap})
}

// AdminShowReservation shows the content of a single reservation in the admin layout.
func (m *Repository) AdminShowReservation(writer http.ResponseWriter, request *http.Request) {
	// User can come from the all reservations or new reservations list.
	exploded := strings.Split(request.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4]) // actual URL is for ex. localhost:8080/admin/reservations/all/{ID}
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	src := exploded[3] // could be "all" or "new".
	stringMap := map[string]string{"src": src}

	// Get reservation from the DB.
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	render.Template(writer, request, "admin-reservations-show.page.gohtml", &models.TemplateData{
		Data: map[string]interface{}{"reservation": res}, StringMap: stringMap, Form: forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	// User can come from the all reservations or new reservations list.
	exploded := strings.Split(request.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4]) // actual URL is for ex. localhost:8080/admin/reservations/all/{ID}
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	src := exploded[3] // could be "all" or "new".

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}

	res.FirstName = request.Form.Get("first_name")
	res.LastName = request.Form.Get("last_name")
	res.Email = request.Form.Get("email")
	res.Phone = request.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}
	m.App.Session.Put(request.Context(), "flash", "Reservation Saved")
	http.Redirect(writer, request, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(request, "id"))
	src := chi.URLParam(request, "src")

	_ = m.DB.UpdateProcessedForReservation(id, 1)
	m.App.Session.Put(request.Context(), "flash", "Reservation marked as processed")

	http.Redirect(writer, request, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

// AdminDeleteReservation deletes a reservation.
func (m *Repository) AdminDeleteReservation(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(request, "id"))
	src := chi.URLParam(request, "src")

	_ = m.DB.DeleteReservation(id)
	m.App.Session.Put(request.Context(), "flash", "Reservation deleted")

	http.Redirect(writer, request, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

// AdminPostReservationsCalendar handles post requests coming from the reservation calendar in the admin layout.
func (m *Repository) AdminPostReservationsCalendar(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		helpers.ServerError(writer, err)
		return
	}

	year, _ := strconv.Atoi(request.Form.Get("y"))
	month, _ := strconv.Atoi(request.Form.Get("m"))

	m.App.Session.Put(request.Context(), "flash", "Changes saved")
	http.Redirect(writer, request, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)

}

// parseDateFromForm converts a date extracted from an html form to a Go friendly format (usually used to query).
func parseDateFromForm(form url.Values, dateString string) (time.Time, error) {
	// Declare the layout that matches how the date is extracted from the form
	layout := "2006-01-02"
	// Parse it in go-friendly date.
	// Example form date: 2006-01-02
	// Parsed date: 2016-01-02 0:00:00 +0000 UTC
	parsedDate, err := time.Parse(layout, form.Get(dateString))
	return parsedDate, err
}

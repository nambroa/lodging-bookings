package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nambroa/lodging-bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close() // What goes after defer only gets executed after the current function finishes.
	var resp *http.Response
	var err error

	for _, test := range theTests {

		if test.method == "GET" {
			resp, err = testServer.Client().Get(testServer.URL + test.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("For %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation_WithReservationInSession(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room:   models.Room{ID: 1, RoomName: "General's Quarters"},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx) // The request is now aware of the X-Session context.

	// ResponseRecorder simulates an http writer after the whole request-response lifecycle.
	// So in this case it would mean something like "after serveHTTP line, the user has entered the reservation page".
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_Reservation_WithoutReservationInSession(t *testing.T) {
	// Test Case where reservation is not in session (reset everything).
	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_Reservation_WithNonExistentRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1234,
		Room:   models.Room{ID: 1, RoomName: "General's Quarters"},
	}
	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostReservation_ValidReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID:    1,
		Room:      models.Room{ID: 1, RoomName: "General's Quarters"},
		StartDate: time.Now(),
		EndDate:   time.Now(),
	}

	postedData := url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "123456789")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Telling the webserver its a form post.

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation_WithNoReservationInSession(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "123456789")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Telling the webserver its a form post.

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation_WithInvalidFormFormat(t *testing.T) {
	reqBody := "first_nameeeeJohn"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_nameasdSmith")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Telling the webserver its a form post.

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation_WithInvalidRoomID(t *testing.T) {
	reservation := models.Reservation{
		RoomID:    1321,
		Room:      models.Room{ID: 1, RoomName: "General's Quarters"},
		StartDate: time.Now(),
		EndDate:   time.Now(),
	}

	postedData := url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "123456789")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Telling the webserver its a form post.

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_AvailabilityJSON_InvalidRoom(t *testing.T) {
	reqBody := "start=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1") // Room does not exist

	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "x-www-form-urlencoded")
	handler := http.HandlerFunc(Repo.AvailabilityJSON)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Transform response into json format
	var j jsonResponse
	err := json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Error("AvailabilityJSON test returned OK for an invalid room. Response: ", j)
	}

}

// getCtx gets the context from the session in the request.
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

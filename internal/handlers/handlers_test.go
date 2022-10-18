package handlers

import (
	"context"
	"github.com/nambroa/lodging-bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
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
		t.Errorf("Reservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusOK)
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
		t.Errorf("Reservation handler returned wrong response code. Got %d, wanted %d", rr.Code, http.StatusOK)
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

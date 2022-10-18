package handlers

import (
	"context"
	"github.com/nambroa/lodging-bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	params             []postData
	expectedStatusCode int
}{
	//{"home", "/", "GET", []postData{}, http.StatusOK},
	//{"about", "/about", "GET", []postData{}, http.StatusOK},
	//{"gq", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	//{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	//{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	//{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	//{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	//
	//{"post-search-avail", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "start", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"post-search-avail-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "start", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"post-make-reservation", "/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "John"},
	//	{key: "last_name", value: "Smith"},
	//	{key: "email", value: "me@here.com"},
	//	{key: "phone", value: "1337-3093"},
	//}, http.StatusOK},
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

		if test.method == "POST" {
			values := url.Values{}
			// Populate payload to send in post request (phone, email, dates, etc).
			for _, param := range test.params {
				values.Add(param.key, param.value)
			}
			resp, err = testServer.Client().PostForm(testServer.URL+test.url, values)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("For %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
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

// getCtx gets the context from the session in the request.
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

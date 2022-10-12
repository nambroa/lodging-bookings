package render

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/models"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var session *scs.SessionManager
var testApp config.AppConfig

// Gets called before any of our tests run. Before it closes, it runs the tests.
func TestMain(m *testing.M) {
	// Types that will be stored in the session object (encoded in the session object).
	gob.Register(models.Reservation{})

	// Change this to true when in production.
	testApp.InProduction = false

	// Adding logs to the app config.
	testApp.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false
	testApp.Session = session
	app = &testApp // The app used in render.go is now pointing to our test copy.

	os.Exit(m.Run())
}

type TestResponseWriter struct{}

func (tw *TestResponseWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *TestResponseWriter) WriteHeader(i int) {

}

func (tw *TestResponseWriter) Write(b []byte) (int, error) {
	length := len(b) // What a custom response writer is expected to return (read the docs).
	return length, nil
}

package helpers

import (
	"fmt"
	"github.com/nambroa/lodging-bookings/internal/config"
	"net/http"
	"runtime/debug"
)

var app *config.AppConfig

// NewHelpers passes an existing appConfig to the helpers.
func NewHelpers(a *config.AppConfig) {
	app = a
}

// ClientError raises an http error and writes it to the InfoLog in the App Config.
func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of: ", status)
	http.Error(w, http.StatusText(status), status)
}

// ServerError writes trace info (stack, error) to the ErrorLog in the App Config.
func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

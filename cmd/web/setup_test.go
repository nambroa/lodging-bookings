package main

import (
	"net/http"
	"os"
	"testing"
)

// Gets called before any of our tests run. Before it closes, it runs the tests.
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

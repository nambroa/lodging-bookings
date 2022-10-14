package render

import (
	"github.com/nambroa/lodging-bookings/internal/models"
	"net/http"
	"testing"
)

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r := getSession()
	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Flash value of 123 not found in session.")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	templateCache, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = templateCache
	r := getSession()
	var ww TestResponseWriter

	err = Template(&ww, r, "home.page.gohtml", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser", err)
	}

	err = Template(&ww, r, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist")
	}
}
func getSession() *http.Request {
	r, _ := http.NewRequest("GET", "/some-url", nil)

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session")) // Adds session info to context.
	r = r.WithContext(ctx)

	return r

}

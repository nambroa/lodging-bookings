package render

// All the files for a package must exist on the same directory
import (
	"bytes"
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// Specify functions that are available to the golang templates.
var functions = template.FuncMap{"humanDate": HumanDate}

var app *config.AppConfig

var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package.
func NewRenderer(aConfig *config.AppConfig) {
	app = aConfig
}

// HumanDate returns time in YYYY-MM-DD format.
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}
func AddDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	templateData.CSRFToken = nosurf.Token(r)
	// PopString since you want to show these messages only once to the user.
	// PopString removes it from the session after returning.
	// Example line that adds data = m.App.Session.Put(r.Context(), "error", "Can't get reservation from session.")
	templateData.Flash = app.Session.PopString(r.Context(), "flash")
	templateData.Error = app.Session.PopString(r.Context(), "error")
	templateData.Warning = app.Session.PopString(r.Context(), "warning")
	if app.Session.Exists(r.Context(), "user_id") {
		templateData.IsAuthenticated = 1
	}
	return templateData
}

// Template renders a specific html template to the writer w with filename ending in tmpl.
func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	// Get cache from the app config.
	// I want to rebuild the cache if useCache is false, for example when I'm developing the app (aka "dev mode")
	var templateCache map[string]*template.Template
	if app.UseCache {
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	// Get template from cache.
	templ, templateFound := templateCache[tmpl]
	if !templateFound {
		return errors.New("cannot get template from cache")
	}

	// This buffer is arbitrary code to add more error checking below. Not needed.
	// Without this, you can only error check when writing to w.
	buf := new(bytes.Buffer)
	templateData = AddDefaultData(templateData, r)
	err := templ.Execute(buf, templateData)
	if err != nil {
		fmt.Println("error executing template:", err)
		return err
	}

	// Render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {

	templateCache := map[string]*template.Template{} // Empty map, same as the make function.

	// Get all the files name *.fileName.gohtml from ./templates (FULL PATH from root project)
	// Ex: templates/home.page.gohtml
	fileNames, err := filepath.Glob(fmt.Sprintf("%s/*.page.gohtml", pathToTemplates))
	if err != nil {
		return templateCache, err
	}

	// Range through the received files.
	for _, fileName := range fileNames {
		name := filepath.Base(fileName) // Filename MINUS the full path. Ex home.fileName.gohtml.

		// Create template from the current fileName being iterated by the loop.
		// The line below creates a template with the name I want (home.filename.gohtml for example).
		// And it also populates it with the corresponding file.
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(fileName) // Page is the FULL PATH.
		if err != nil {
			return templateCache, err
		}

		// Search for layouts, that were not searched in the initial filepath.Glob and need to be applied in
		// each template (since they are a layout).
		fileNames, err := filepath.Glob(fmt.Sprintf("%s/*layout.gohtml", pathToTemplates))
		if err != nil {
			return templateCache, err
		}
		// If any layout fileNames were found, add the layout to the existing templateSet. So it mixes layout + page.
		if len(fileNames) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return templateCache, err
			}
		}

		templateCache[name] = templateSet
	}
	return templateCache, nil
}

package render

// All the files for a package must exist on the same directory
import (
	"bytes"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/nambroa/lodging-bookings/pkg/config"
	"github.com/nambroa/lodging-bookings/pkg/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var functions = template.FuncMap{}

var app *config.AppConfig

// NewTemplates sets the config for the template package.
func NewTemplates(aConfig *config.AppConfig) {
	app = aConfig
}
func AddDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	templateData.CSRFToken = nosurf.Token(r)
	return templateData
}

// RenderTemplate renders a specific html template to the writer w with filename ending in tmpl.
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) {
	// Get cache from the app config.
	// I want to rebuild the cache if useCache is false, for example when I'm developing the app (aka "dev mode")
	var templateCache map[string]*template.Template
	if app.UseCache {
		templateCache = map[string]*template.Template{} // Empty map, same as the make function.
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	// Get template from cache.
	templ, templateFound := templateCache[tmpl]
	if !templateFound {
		log.Fatal("Template not found in cache")
	}

	// This buffer is arbitrary code to add more error checking below. Not needed.
	// Without this, you can only error check when writing to w.
	buf := new(bytes.Buffer)
	templateData = AddDefaultData(templateData, r)
	err := templ.Execute(buf, templateData)
	if err != nil {
		fmt.Println("error executing template:", err)
	}

	// Render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {

	templateCache := map[string]*template.Template{} // Empty map, same as the make function.

	// Get all the files name *.fileName.gohtml from ./templates (FULL PATH from root project)
	// Ex: templates/home.page.gohtml
	fileNames, err := filepath.Glob("./templates/*.page.gohtml")
	if err != nil {
		return templateCache, err
	}

	// Range through the received files.
	for _, fileName := range fileNames {
		name := filepath.Base(fileName) // Filename MINUS the full path. Ex home.fileName.gohtml.

		// Create template from the current fileName being iterated by the loop.
		// The line below creates a template with the name I want (home.filename.gohtml for example).
		// And it also populates it with the corresponding file.
		templateSet, err := template.New(name).ParseFiles(fileName) // Page is the FULL PATH.
		if err != nil {
			return templateCache, err
		}

		// Search for layouts, that were not searched in the initial filepath.Glob and need to be applied in
		// each template (since they are a layout).
		fileNames, err := filepath.Glob("./templates/*layout.gohtml")
		if err != nil {
			return templateCache, err
		}
		// If any layout fileNames were found, add the layout to the existing templateSet. So it mixes layout + page.
		if len(fileNames) > 0 {
			templateSet, err = templateSet.ParseGlob("./templates/*.layout.gohtml")
			if err != nil {
				return templateCache, err
			}
		}

		templateCache[name] = templateSet
	}
	return templateCache, nil
}

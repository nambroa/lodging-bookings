package models

import "github.com/nambroa/lodging-bookings/internal/forms"

// TemplateData holds data sent from handlers to templates.
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{} // Map of string to "anything" (think T types).
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated int // Default value is 0 (int default value) meaning not authed.
}

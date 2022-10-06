package forms

import (
	"net/http"
	"net/url"
)

// Form creates a custom form struct and embeds a custom url.Values object.
type Form struct {
	url.Values
	Errors errors
}

// Valid returns true if there are no errors, otherwise false.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initializes a form struct and returns a pointer to make sure we are always handling the same form.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Has(field string, r *http.Request) bool {
	formField := r.Form.Get(field)
	if formField == "" {
		f.Errors.Add(field, "This field cannot be blank")
	}
	return formField != ""
}

package forms

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/url"
	"strings"
)

// Form creates a custom form struct and embeds a custom url.Values object.
type Form struct {
	url.Values
	Errors errors
}

// New initializes a form struct and returns a pointer to make sure we are always handling the same form.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Valid returns true if there are no errors, otherwise false.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Required checks for required fields.
func (f *Form) Required(fieldNames ...string) {
	for _, fieldName := range fieldNames {
		value := f.Get(fieldName)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(fieldName, "This field cannot be blank")
		}
	}
}

// Has checks if the form field in the POST request is not empty.
func (f *Form) Has(fieldName string) bool {
	formField := f.Get(fieldName)
	return formField != ""
}

// MinLength checks that the field is at least length characters long.
func (f *Form) MinLength(fieldName string, length int) bool {
	formField := f.Get(fieldName)
	if len(formField) < length {
		f.Errors.Add(fieldName, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail checks for valid email address.
func (f *Form) IsEmail(fieldName string) {
	if !govalidator.IsEmail(f.Get(fieldName)) {
		f.Errors.Add(fieldName, "Invalid email address")
	}
}

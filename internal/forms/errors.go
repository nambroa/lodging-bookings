package forms

type errors map[string][]string

// Add adds an error message for a given form field.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message.
func (e errors) Get(field string) string {
	errorsForField := e[field]
	if len(errorsForField) == 0 {
		return ""
	}
	return errorsForField[0]
}

package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_IsEmail_ShouldReturnTrueForValidEmail(t *testing.T) {
	postData := url.Values{}
	postData.Add("valid_email", "cleaver@gmail.com")
	form := New(postData)

	form.IsEmail("valid_email")
	if !form.Valid() {
		t.Error("Form returns invalid email for a valid email")
	}
}

func TestForm_IsEmail_ShouldReturnFalseForInvalidEmail(t *testing.T) {
	postData := url.Values{}
	postData.Add("invalid_email", "whisk.com")
	form := New(postData)

	form.IsEmail("invalid_email")
	if form.Valid() {
		t.Error("Form returns valid email for a invalid email")
	}
}

func TestForm_MinLength_ShouldReturnTrueForFieldWithAppropriateLength(t *testing.T) {
	postData := url.Values{}
	postData.Add("test_field", "test_value")
	form := New(postData)

	lengthRequirement := form.MinLength("test_field", 5)
	if !lengthRequirement {
		t.Error("Form returns insufficient length of 5 for value test_value when it should be sufficient")
	}
}

func TestForm_MinLength_ShouldReturnFalseForFieldWithInappropriateLength(t *testing.T) {
	postData := url.Values{}
	postData.Add("test_field", "test_value")
	form := New(postData)

	lengthRequirement := form.MinLength("test_field", 66)
	if lengthRequirement {
		t.Error("Form returns sufficient length of 66 for value test_value when it should be insufficient")
	}
}

func TestForm_Has_ShouldReturnFalseWithNoFields(t *testing.T) {
	form := New(url.Values{})

	hasField := form.Has("field_that_does_not_exist")
	if hasField {
		t.Error("The form has field_that_does_not_exist when it should not have it")
	}
}

func TestForm_Has_ShouldReturnTrueForExistingField(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("field_that_exists", "test_value")
	r := httptest.NewRequest("POST", "/test", nil)
	r.PostForm = postedData
	form := New(r.PostForm)

	hasField := form.Has("field_that_exists")
	if !hasField {
		t.Error("The form does not have field_that_exists when it should have it")
	}
}

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("Form is invalid when it should have been valid.")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", nil)
	form := New(r.PostForm)

	form.Required("a", "b")
	if form.Valid() {
		t.Error("Form is valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")

	r, _ = http.NewRequest("POST", "/test", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b")
	if !form.Valid() {
		t.Error("Form is invalid when required fields are present")
	}
}

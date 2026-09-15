package validator

import "testing"

type sample struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required"`
}

func TestValidate_ValidInput(t *testing.T) {
	errs := Validate(sample{Email: "a@b.com", Name: "Ali"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidate_InvalidInput(t *testing.T) {
	errs := Validate(sample{Email: "not-an-email", Name: ""})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

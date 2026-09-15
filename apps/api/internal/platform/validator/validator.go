package validator

import (
	"fmt"

	govalidator "github.com/go-playground/validator/v10"
)

var validate = govalidator.New()

func Validate(s any) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	verrs, ok := err.(govalidator.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "", Message: err.Error()}}
	}

	var fieldErrors []FieldError
	for _, verr := range verrs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   verr.Field(),
			Message: fmt.Sprintf("%s failed on %s", verr.Field(), verr.Tag()),
		})
	}
	return fieldErrors
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

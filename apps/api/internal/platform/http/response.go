package httpx

import (
	"encoding/json"
	"net/http"

	"basecode/api/internal/platform/validator"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// kirim response JSON dengan status HTTP yang diberikan.
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

// bikin format error standar lalu kirim sebagai JSON.
func WriteError(w http.ResponseWriter, status int, code, message string) error {
	errResp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	return WriteJSON(w, status, errResp)
}

type ValidationErrorResponse struct {
	Error ValidationErrorDetail `json:"error"`
}

type ValidationErrorDetail struct {
	Code   string                 `json:"code"` // always "validation_error"
	Fields []validator.FieldError `json:"fields"`
}

func WriteValidationError(w http.ResponseWriter, fields []validator.FieldError) error {
	return WriteJSON(w, http.StatusUnprocessableEntity, ValidationErrorResponse{
		Error: ValidationErrorDetail{Code: "validation_error", Fields: fields},
	})
}

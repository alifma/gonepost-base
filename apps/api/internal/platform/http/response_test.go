package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"basecode/api/internal/platform/validator"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %s", got["status"])
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "invalid_input", "email is required")

	var got ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if got.Error.Code != "invalid_input" {
		t.Fatalf("expected error code 'invalid_input', got %s", got.Error.Code)
	}
	if got.Error.Message != "email is required" {
		t.Fatalf("expected error message 'email is required', got %s", got.Error.Message)
	}
}

func TestWriteValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	fields := []validator.FieldError{{Field: "Email", Message: "required"}}
	WriteValidationError(rec, fields)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, rec.Code)
	}

	var got ValidationErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if got.Error.Code != "validation_error" {
		t.Fatalf("expected error code 'validation_error', got %s", got.Error.Code)
	}
	if len(got.Error.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(got.Error.Fields))
	}
}

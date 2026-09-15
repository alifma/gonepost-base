package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"basecode/api/internal/platform/logger"
)

func TestRecover_CatchesPanic(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.NewWithWriter(&buf)
	panicky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	handler := Recover(testLogger)(panicky)
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "internal_error") {
		t.Errorf("expected response body to contain %q, got %q", "internal_error", rec.Body.String())
	}

	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("expected log to contain %q, got %q", "panic recovered", buf.String())
	}

}

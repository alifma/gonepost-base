package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var seenID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = FromContext(r.Context())
	})
	handler := RequestID(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seenID == "" {
		t.Fatal("expected non-empty request ID in context")
	}
	if rec.Header().Get(RequestIDHeader) != seenID {
		t.Fatalf("expected response header %s to be %s, got %s", RequestIDHeader, seenID, rec.Header().Get(RequestIDHeader))
	}
}

func TestRequestID_PreservesIncoming(t *testing.T) {
	var seenID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = FromContext(r.Context())
	})
	handler := RequestID(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "fixed-test-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seenID != "fixed-test-id" {
		t.Fatalf("expected request ID in context to be 'fixed-test-id', got %s", seenID)
	}
	if rec.Header().Get(RequestIDHeader) != "fixed-test-id" {
		t.Fatalf("expected response header %s to be 'fixed-test-id', got %s", RequestIDHeader, rec.Header().Get(RequestIDHeader))
	}
}

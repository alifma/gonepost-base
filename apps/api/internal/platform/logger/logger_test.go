package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewWithWriter(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithWriter(&buf)
	l.Info("hello")

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if got["msg"] != "hello" {
		t.Errorf("expected msg %q, got %q", "hello", got["msg"])
	}
	if got["service"] != "api" {
		t.Errorf("expected service %q, got %q", "api", got["service"])
	}
}

package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRun_ShutsDownOnCancel(t *testing.T) {
	handler := http.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, "127.0.0.1:18080", handler)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server to shut down")
	}
}

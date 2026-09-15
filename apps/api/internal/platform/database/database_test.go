//go:build integration

package database

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	dsn := "postgres://basecode:basecode@localhost:5432/basecode?sslmode=disable"

	pool, err := New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTestDatabaseURL = "postgres://basecode:basecode@localhost:5433/basecode_test?sslmode=disable"

func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultTestDatabaseURL
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func UniqueEmail(t *testing.T) string {
	t.Helper()
	return strings.ToLower(fmt.Sprintf("test-%s@example.com", t.Name()))
}

func CreateUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		"INSERT INTO users (email, password_hash, status) VALUES (lower($1), 'x', 'active') RETURNING id", email).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	})
	return id
}

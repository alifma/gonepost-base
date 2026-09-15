//go:build integration

package users

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://basecode:basecode@localhost:5432/basecode?sslmode=disable")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// uniqueEmail returns a lowercase email — Create stores email as
// lower(email), so tests must compare against the same casing.
func uniqueEmail(t *testing.T) string {
	return strings.ToLower(fmt.Sprintf("test-%s@example.com", t.Name()))
}

func TestRepository_CreateAndGet(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := uniqueEmail(t)

	created, err := repo.Create(ctx, User{
		Email:        email,
		PasswordHash: "fake-hash",
		Status:       StatusActive,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", created.ID)
	})

	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	byEmail, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if byEmail.ID != created.ID {
		t.Errorf("expected same user by email, got different ID")
	}

	byID, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if byID.Email != email {
		t.Errorf("expected email %s, got %s", email, byID.Email)
	}
}

func TestRepository_CreateDuplicateEmail(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := uniqueEmail(t)

	created, err := repo.Create(ctx, User{Email: email, PasswordHash: "fake-hash", Status: StatusActive})
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", created.ID)
	})

	_, err = repo.Create(ctx, User{Email: email, PasswordHash: "fake-hash", Status: StatusActive})
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRepository_GetByEmail_NotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "does-not-exist@example.com")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_UpdateAndList(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := uniqueEmail(t)

	created, err := repo.Create(ctx, User{Email: email, PasswordHash: "fake-hash", Status: StatusActive})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", created.ID)
	})

	newName := "Updated Name"
	updated, err := repo.Update(ctx, created.ID, nil, &newName, StatusInactive)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.FullName == nil || *updated.FullName != newName {
		t.Errorf("expected full name %q, got %v", newName, updated.FullName)
	}
	if updated.Status != StatusInactive {
		t.Errorf("expected status inactive, got %s", updated.Status)
	}

	list, total, err := repo.List(ctx, ListFilter{Status: StatusInactive, Limit: 50})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	found := false
	for _, u := range list {
		if u.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Error("expected created user to appear in inactive-status list")
	}
	if total < 1 {
		t.Errorf("expected total >= 1, got %d", total)
	}
}

//go:build integration

package users

import (
	"context"
	"testing"

	"basecode/api/tests/testutil"
)

func TestRepository_CreateAndGet(t *testing.T) {
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := testutil.UniqueEmail(t)

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
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := testutil.UniqueEmail(t)

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
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "does-not-exist@example.com")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_UpdateAndList(t *testing.T) {
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	email := testutil.UniqueEmail(t)

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

func TestRepository_ListSearchAndPagination(t *testing.T) {
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	firstEmail := testutil.UniqueEmail(t) + "-first@example.com"
	secondEmail := testutil.UniqueEmail(t) + "-second@example.com"

	first, err := repo.Create(ctx, User{Email: firstEmail, FullName: stringPtr("Searchable Alpha"), PasswordHash: "fake-hash", Status: StatusActive})
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	second, err := repo.Create(ctx, User{Email: secondEmail, FullName: stringPtr("Searchable Beta"), PasswordHash: "fake-hash", Status: StatusActive})
	if err != nil {
		t.Fatalf("second Create failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id IN ($1, $2)", first.ID, second.ID)
	})

	list, total, err := repo.List(ctx, ListFilter{Search: "Searchable", Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 2 || len(list) != 1 {
		t.Fatalf("expected 2 matches and one paginated row, got total=%d rows=%d", total, len(list))
	}

	list, _, err = repo.List(ctx, ListFilter{Search: "Beta", Limit: 10})
	if err != nil {
		t.Fatalf("second search failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != second.ID {
		t.Fatalf("expected beta user, got %+v", list)
	}
}

func stringPtr(value string) *string { return &value }

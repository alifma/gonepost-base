//go:build integration

package items

import (
	"context"
	"errors"
	"testing"

	"basecode/api/tests/testutil"
)

func TestRepository_CRUDAndOwnerScoping(t *testing.T) {
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	alice := testutil.CreateUser(t, pool, "alice-"+testutil.UniqueEmail(t))
	bob := testutil.CreateUser(t, pool, "bob-"+testutil.UniqueEmail(t))

	desc := "first"
	created, err := repo.Create(ctx, alice, Input{Name: "Laptop", Description: &desc, Status: StatusActive})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.OwnerID != alice || created.Name != "Laptop" {
		t.Errorf("unexpected item: %+v", created)
	}

	// another user can neither see nor change nor delete it
	if _, err := repo.Get(ctx, bob, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get as other owner: expected ErrNotFound, got %v", err)
	}
	if _, err := repo.Update(ctx, bob, created.ID, Input{Name: "Hacked", Status: StatusActive}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update as other owner: expected ErrNotFound, got %v", err)
	}
	if err := repo.Delete(ctx, bob, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete as other owner: expected ErrNotFound, got %v", err)
	}
	if _, err := repo.Get(ctx, alice, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get with junk id: expected ErrNotFound, got %v", err)
	}

	updated, err := repo.Update(ctx, alice, created.ID, Input{Name: "Laptop 2", Status: StatusArchived})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Laptop 2" || updated.Status != StatusArchived || updated.Description != nil {
		t.Errorf("unexpected updated item: %+v", updated)
	}

	list, total, err := repo.List(ctx, alice, ListFilter{Status: StatusArchived, Search: "laptop"})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("List: expected 1 item, got total=%d len=%d err=%v", total, len(list), err)
	}
	if _, total, _ := repo.List(ctx, bob, ListFilter{}); total != 0 {
		t.Errorf("List as other owner: expected 0, got %d", total)
	}

	if err := repo.Delete(ctx, alice, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := repo.Get(ctx, alice, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after delete: expected ErrNotFound, got %v", err)
	}
}

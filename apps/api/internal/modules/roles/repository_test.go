//go:build integration

package roles

import (
	"context"
	"fmt"
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

func createTestUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		"INSERT INTO users (email, password_hash, status) VALUES (lower($1), 'x', 'active') RETURNING id", email).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	})
	return id
}

func TestRepository_RoleAndPermissionLifecycle(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	roleName := fmt.Sprintf("TEST_ROLE_%s", t.Name())

	role, err := repo.Create(ctx, roleName, nil)
	if err != nil {
		t.Fatalf("Create role failed: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM roles WHERE id = $1", role.ID) })

	permCode := fmt.Sprintf("test:%s", t.Name())
	perm, err := repo.EnsurePermission(ctx, permCode, nil)
	if err != nil {
		t.Fatalf("EnsurePermission failed: %v", err)
	}
	t.Cleanup(func() { pool.Exec(context.Background(), "DELETE FROM permissions WHERE id = $1", perm.ID) })

	if err := repo.GrantPermission(ctx, role.ID, perm.ID); err != nil {
		t.Fatalf("GrantPermission failed: %v", err)
	}
	// idempotent — granting twice must not error
	if err := repo.GrantPermission(ctx, role.ID, perm.ID); err != nil {
		t.Fatalf("GrantPermission (repeat) failed: %v", err)
	}

	codes, err := repo.ListRolePermissions(ctx, role.ID)
	if err != nil {
		t.Fatalf("ListRolePermissions failed: %v", err)
	}
	if len(codes) != 1 || codes[0] != permCode {
		t.Errorf("expected [%s], got %v", permCode, codes)
	}

	userID := createTestUser(t, pool, fmt.Sprintf("test-%s@example.com", t.Name()))

	has, err := repo.UserHasPermission(ctx, userID, permCode)
	if err != nil {
		t.Fatalf("UserHasPermission failed: %v", err)
	}
	if has {
		t.Error("expected user without role to not have permission")
	}

	if err := repo.AssignToUser(ctx, userID, role.ID); err != nil {
		t.Fatalf("AssignToUser failed: %v", err)
	}

	has, err = repo.UserHasPermission(ctx, userID, permCode)
	if err != nil {
		t.Fatalf("UserHasPermission failed: %v", err)
	}
	if !has {
		t.Error("expected user with role to have permission")
	}

	count, err := repo.CountUsersWithRole(ctx, role.ID)
	if err != nil {
		t.Fatalf("CountUsersWithRole failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if err := repo.RemoveFromUser(ctx, userID, role.ID); err != nil {
		t.Fatalf("RemoveFromUser failed: %v", err)
	}

	has, err = repo.UserHasPermission(ctx, userID, permCode)
	if err != nil {
		t.Fatalf("UserHasPermission failed: %v", err)
	}
	if has {
		t.Error("expected permission to be gone after role removed")
	}
}

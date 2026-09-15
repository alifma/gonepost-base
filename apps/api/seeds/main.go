// Command seeds populates local/dev Postgres with:
//   - every known permission (from internal/modules/permissions)
//   - the SUPER_ADMIN system role, granted all permissions
//   - a starter user, made SUPER_ADMIN
//
// Idempotent — safe to run multiple times.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"basecode/api/internal/modules/permissions"
	"basecode/api/internal/modules/roles"
	"basecode/api/internal/platform/security"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	seedEmail    = "admin@example.com"
	seedPassword = "changeme123"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	rolesRepo := roles.NewRepository(pool)

	// 1. every known permission exists
	for _, code := range permissions.All {
		if _, err := rolesRepo.EnsurePermission(ctx, code, nil); err != nil {
			log.Fatalf("failed to ensure permission %s: %v", code, err)
		}
	}
	fmt.Printf("seed: ensured %d permissions\n", len(permissions.All))

	// 2. SUPER_ADMIN role exists, granted every permission
	superAdmin, err := rolesRepo.GetByName(ctx, roles.SuperAdminRole)
	if err != nil {
		desc := "full access to every permission"
		superAdmin, err = rolesRepo.CreateSystemRole(ctx, roles.SuperAdminRole, &desc)
		if err != nil {
			log.Fatalf("failed to create SUPER_ADMIN role: %v", err)
		}
		fmt.Println("seed: created role", roles.SuperAdminRole)
	}
	for _, code := range permissions.All {
		perm, err := rolesRepo.GetPermissionByCode(ctx, code)
		if err != nil {
			log.Fatalf("failed to look up permission %s: %v", code, err)
		}
		if err := rolesRepo.GrantPermission(ctx, superAdmin.ID, perm.ID); err != nil {
			log.Fatalf("failed to grant %s to SUPER_ADMIN: %v", code, err)
		}
	}

	// 3. starter user exists and is SUPER_ADMIN
	var userID string
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE lower(email) = lower($1)", seedEmail).Scan(&userID)
	if err != nil {
		hash, err := security.Hash(seedPassword)
		if err != nil {
			log.Fatalf("failed to hash password: %v", err)
		}
		err = pool.QueryRow(ctx,
			"INSERT INTO users (email, full_name, password_hash, status) VALUES (lower($1), $2, $3, 'active') RETURNING id",
			seedEmail, "Seed Admin", hash).Scan(&userID)
		if err != nil {
			log.Fatalf("failed to insert seed user: %v", err)
		}
		fmt.Printf("seed: created user %s / %s\n", seedEmail, seedPassword)
	} else {
		fmt.Println("seed: user", seedEmail, "already exists")
	}

	if err := rolesRepo.AssignToUser(ctx, userID, superAdmin.ID); err != nil {
		log.Fatalf("failed to assign SUPER_ADMIN to seed user: %v", err)
	}
	fmt.Printf("seed: %s is SUPER_ADMIN\n", seedEmail)
}

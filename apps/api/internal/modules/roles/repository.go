package roles

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound           = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name string, description *string) (Role, error) {
	const q = `
		INSERT INTO roles (name, description) VALUES ($1, $2)
		RETURNING id, name, description, is_system, created_at`
	row := r.pool.QueryRow(ctx, q, name, description)
	return scanRole(row)
}

// CreateSystemRole is for seed-only use (e.g. SUPER_ADMIN) — marks the role
// is_system so it's distinguishable from roles created via the API.
func (r *Repository) CreateSystemRole(ctx context.Context, name string, description *string) (Role, error) {
	const q = `
		INSERT INTO roles (name, description, is_system) VALUES ($1, $2, true)
		RETURNING id, name, description, is_system, created_at`
	row := r.pool.QueryRow(ctx, q, name, description)
	return scanRole(row)
}

func (r *Repository) GetByID(ctx context.Context, id string) (Role, error) {
	const q = `SELECT id, name, description, is_system, created_at FROM roles WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	role, err := scanRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	return role, err
}

func (r *Repository) GetByName(ctx context.Context, name string) (Role, error) {
	const q = `SELECT id, name, description, is_system, created_at FROM roles WHERE upper(name) = upper($1)`
	row := r.pool.QueryRow(ctx, q, name)
	role, err := scanRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	return role, err
}

func (r *Repository) List(ctx context.Context) ([]Role, error) {
	const q = `SELECT id, name, description, is_system, created_at FROM roles ORDER BY created_at`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

// EnsurePermission creates the permission if it doesn't exist yet
// (idempotent — used by the seed script).
func (r *Repository) EnsurePermission(ctx context.Context, code string, description *string) (Permission, error) {
	const q = `
		INSERT INTO permissions (code, description) VALUES ($1, $2)
		ON CONFLICT (code) DO UPDATE SET description = EXCLUDED.description
		RETURNING id, code, description`
	row := r.pool.QueryRow(ctx, q, code, description)
	return scanPermission(row)
}

func (r *Repository) GetPermissionByCode(ctx context.Context, code string) (Permission, error) {
	const q = `SELECT id, code, description FROM permissions WHERE code = $1`
	row := r.pool.QueryRow(ctx, q, code)
	p, err := scanPermission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Permission{}, ErrPermissionNotFound
	}
	return p, err
}

// GrantPermission is idempotent — granting an already-granted permission is a no-op.
func (r *Repository) GrantPermission(ctx context.Context, roleID, permissionID string) error {
	const q = `
		INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, roleID, permissionID)
	return err
}

func (r *Repository) ListRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	const q = `
		SELECT p.code FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.code`
	rows, err := r.pool.Query(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// AssignToUser is idempotent — assigning an already-held role is a no-op.
func (r *Repository) AssignToUser(ctx context.Context, userID, roleID string) error {
	const q = `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, userID, roleID)
	return err
}

func (r *Repository) RemoveFromUser(ctx context.Context, userID, roleID string) error {
	const q = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.pool.Exec(ctx, q, userID, roleID)
	return err
}

func (r *Repository) ListUserRoles(ctx context.Context, userID string) ([]Role, error) {
	const q = `
		SELECT r.id, r.name, r.description, r.is_system, r.created_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

// UserHasPermission is the core authorization query: does this user hold
// any role that grants this permission code?
func (r *Repository) UserHasPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON rp.role_id = ur.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE ur.user_id = $1 AND p.code = $2
		)`
	var exists bool
	err := r.pool.QueryRow(ctx, q, userID, permissionCode).Scan(&exists)
	return exists, err
}

// CountUsersWithRole is used to protect the last active SUPER_ADMIN from
// being demoted or deleted.
func (r *Repository) CountUsersWithRole(ctx context.Context, roleID string) (int, error) {
	const q = `SELECT count(*) FROM user_roles WHERE role_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, q, roleID).Scan(&count)
	return count, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRole(row scanner) (Role, error) {
	var r Role
	err := row.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem, &r.CreatedAt)
	return r, err
}

func scanPermission(row scanner) (Permission, error) {
	var p Permission
	err := row.Scan(&p.ID, &p.Code, &p.Description)
	return p, err
}

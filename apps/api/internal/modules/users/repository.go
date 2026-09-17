package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already registered")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type ListFilter struct {
	Status Status // empty means no filter
	Search string // matches email, username, or full name
	Limit  int
	Offset int
}

func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	const q = `
		INSERT INTO users (email, username, full_name, password_hash, status)
		VALUES (lower($1), $2, $3, $4, $5)
		RETURNING id, email, username, full_name, password_hash, status, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, u.Email, u.Username, u.FullName, u.PasswordHash, u.Status)
	created, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	return created, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `
		SELECT id, email, username, full_name, password_hash, status, created_at, updated_at
		FROM users WHERE lower(email) = lower($1)`

	row := r.pool.QueryRow(ctx, q, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	const q = `
		SELECT id, email, username, full_name, password_hash, status, created_at, updated_at
		FROM users WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]User, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	const countQuery = `
		SELECT count(*) FROM users
		WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%' OR full_name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR status = $2)`
	const listQuery = `
		SELECT id, email, username, full_name, password_hash, status, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%' OR full_name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, f.Search, f.Status).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, listQuery, f.Search, f.Status, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, u)
	}
	return result, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id string, username, fullName *string, status Status) (User, error) {
	const q = `
		UPDATE users
		SET username = $2, full_name = $3, status = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, email, username, full_name, password_hash, status, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, id, username, fullName, status)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.FullName, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

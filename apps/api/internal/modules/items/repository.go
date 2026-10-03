package items

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("item not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type ListFilter struct {
	Status Status // empty means no filter
	Search string // matches name or description
	Limit  int
	Offset int
}

const itemCols = `id, owner_id, name, description, status, created_at, updated_at`

// Every query below filters on owner_id: a caller can never see or touch
// another user's rows, and a wrong owner looks exactly like "not found".

func (r *Repository) Create(ctx context.Context, ownerID string, in Input) (Item, error) {
	const q = `
		INSERT INTO items (owner_id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + itemCols

	return scanItem(r.pool.QueryRow(ctx, q, ownerID, in.Name, in.Description, in.Status))
}

func (r *Repository) Get(ctx context.Context, ownerID, id string) (Item, error) {
	const q = `SELECT ` + itemCols + ` FROM items WHERE owner_id = $1 AND id = $2`

	it, err := scanItem(r.pool.QueryRow(ctx, q, ownerID, id))
	return it, notFound(err)
}

func (r *Repository) List(ctx context.Context, ownerID string, f ListFilter) ([]Item, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	const where = `
		WHERE owner_id = $1
		AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
		AND ($3 = '' OR status = $3)`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM items`+where, ownerID, f.Search, f.Status).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+itemCols+` FROM items`+where+` ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
		ownerID, f.Search, f.Status, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, it)
	}
	return result, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, ownerID, id string, in Input) (Item, error) {
	const q = `
		UPDATE items
		SET name = $3, description = $4, status = $5, updated_at = now()
		WHERE owner_id = $1 AND id = $2
		RETURNING ` + itemCols

	it, err := scanItem(r.pool.QueryRow(ctx, q, ownerID, id, in.Name, in.Description, in.Status))
	return it, notFound(err)
}

func (r *Repository) Delete(ctx context.Context, ownerID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM items WHERE owner_id = $1 AND id = $2`, ownerID, id)
	if err != nil {
		return notFound(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(row scanner) (Item, error) {
	var it Item
	err := row.Scan(&it.ID, &it.OwnerID, &it.Name, &it.Description, &it.Status, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}

// notFound maps "no row" and "malformed uuid in the path" (SQLSTATE 22P02)
// to ErrNotFound, so a junk id answers 404 instead of 500.
func notFound(err error) error {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgx.ErrNoRows) || (errors.As(err, &pgErr) && pgErr.Code == "22P02") {
		return ErrNotFound
	}
	return err
}

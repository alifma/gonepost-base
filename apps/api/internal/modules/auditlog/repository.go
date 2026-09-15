package auditlog

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Insert(ctx context.Context, e Event) error {
	metadata := e.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO audit_logs (actor_user_id, action, resource, resource_id, result, request_id, trace_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = r.pool.Exec(ctx, q, e.ActorUserID, e.Action, e.Resource, e.ResourceID, e.Result, e.RequestID, e.TraceID, metadataJSON)
	return err
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Event, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	where := ""
	args := []any{}
	addClause := func(clause string, val any) {
		args = append(args, val)
		if where == "" {
			where = "WHERE " + clause + " = $" + strconv.Itoa(len(args))
		} else {
			where += " AND " + clause + " = $" + strconv.Itoa(len(args))
		}
	}
	if f.Action != "" {
		addClause("action", f.Action)
	}
	if f.ActorUserID != "" {
		addClause("actor_user_id", f.ActorUserID)
	}

	countQ := "SELECT count(*) FROM audit_logs " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), f.Limit, f.Offset)
	listQ := `
		SELECT id, actor_user_id, action, resource, resource_id, result, request_id, trace_id, metadata, created_at
		FROM audit_logs ` + where + `
		ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, e)
	}
	return result, total, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(row scanner) (Event, error) {
	var e Event
	var metadataJSON []byte
	err := row.Scan(&e.ID, &e.ActorUserID, &e.Action, &e.Resource, &e.ResourceID, &e.Result, &e.RequestID, &e.TraceID, &metadataJSON, &e.CreatedAt)
	if err != nil {
		return Event{}, err
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &e.Metadata); err != nil {
			return Event{}, err
		}
	}
	return e, nil
}

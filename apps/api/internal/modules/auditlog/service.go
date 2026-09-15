package auditlog

import (
	"context"
	"log/slog"
)

// Inserter is the subset of Repository the service needs — lets tests
// supply a fake without a real database.
type Inserter interface {
	Insert(ctx context.Context, e Event) error
}

// Service is the single place audited actions go through.
//
// Design decision (PLAN.md requires this be explicit): audit writes are
// FAIL-CLOSED, not best-effort. If the audit write fails, Record returns
// the error and callers are expected to fail the whole request (typically
// a 500), even though the underlying mutation already happened. This is a
// simplification — a true fix would wrap the business mutation and the
// audit write in one database transaction, which the current module split
// (separate repositories, separate pool calls) doesn't do. Reliable retry
// via a queue is the alternative PLAN.md mentions, but that needs a worker,
// which is explicitly deferred (see PLAN.md Deferred Work). Fail-closed was
// chosen over silently-best-effort because an audit trail that can silently
// go missing defeats the point of having one.
type Service struct {
	repo   Inserter
	logger *slog.Logger
}

func NewService(repo Inserter, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) Record(ctx context.Context, e Event) error {
	if err := s.repo.Insert(ctx, e); err != nil {
		s.logger.Error("audit write failed",
			"action", e.Action,
			"resource", e.Resource,
			"request_id", e.RequestID,
			"err", err,
		)
		return err
	}
	return nil
}

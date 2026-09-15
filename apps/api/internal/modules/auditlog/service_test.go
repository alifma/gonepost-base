package auditlog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"basecode/api/internal/platform/logger"
)

type fakeInserter struct {
	err     error
	inserts []Event
}

func (f *fakeInserter) Insert(_ context.Context, e Event) error {
	if f.err != nil {
		return f.err
	}
	f.inserts = append(f.inserts, e)
	return nil
}

func testLogger(buf *bytes.Buffer) *slog.Logger {
	return logger.NewWithWriter(buf)
}

func TestService_Record_Success(t *testing.T) {
	repo := &fakeInserter{}
	svc := NewService(repo, testLogger(&bytes.Buffer{}))

	err := svc.Record(context.Background(), Event{Action: ActionLogin, Resource: "auth", Result: ResultSuccess})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.inserts) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(repo.inserts))
	}
}

func TestService_Record_FailClosed(t *testing.T) {
	repo := &fakeInserter{err: errors.New("db unreachable")}
	var logBuf bytes.Buffer
	svc := NewService(repo, testLogger(&logBuf))

	err := svc.Record(context.Background(), Event{Action: ActionLogin, Resource: "auth", Result: ResultSuccess})
	if err == nil {
		t.Fatal("expected Record to return the insert error (fail-closed), got nil")
	}
	if logBuf.Len() == 0 {
		t.Error("expected the failure to also be logged, log output was empty")
	}
}

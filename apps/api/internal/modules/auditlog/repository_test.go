//go:build integration

package auditlog

import (
	"context"
	"testing"

	"basecode/api/tests/testutil"
)

func TestRepository_InsertAndList(t *testing.T) {
	pool := testutil.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	resourceID := "resource-1"
	event := Event{
		Action:     "test.action",
		Resource:   "test",
		ResourceID: &resourceID,
		Result:     ResultSuccess,
		RequestID:  "req-1",
		Metadata:   map[string]any{"foo": "bar"},
	}

	if err := repo.Insert(ctx, event); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM audit_logs WHERE action = 'test.action'")
	})

	list, total, err := repo.List(ctx, ListFilter{Action: "test.action", Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total < 1 {
		t.Fatalf("expected total >= 1, got %d", total)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one event")
	}

	got := list[0]
	if got.Resource != "test" {
		t.Errorf("expected resource 'test', got %s", got.Resource)
	}
	if got.ResourceID == nil || *got.ResourceID != resourceID {
		t.Errorf("expected resource_id %s, got %v", resourceID, got.ResourceID)
	}
	if got.Metadata["foo"] != "bar" {
		t.Errorf("expected metadata foo=bar, got %v", got.Metadata)
	}
}

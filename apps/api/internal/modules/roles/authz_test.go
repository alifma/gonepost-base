package roles

import (
	"context"
	"testing"
)

type fakeChecker struct {
	grants map[string]map[string]bool // userID -> permissionCode -> allowed
}

func (f *fakeChecker) UserHasPermission(_ context.Context, userID, code string) (bool, error) {
	return f.grants[userID][code], nil
}

func TestAuthzService_Can(t *testing.T) {
	checker := &fakeChecker{grants: map[string]map[string]bool{
		"u1": {"users:read": true},
	}}
	svc := NewAuthzService(checker)

	ok, err := svc.Can(context.Background(), "u1", "users:read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected u1 to have users:read")
	}

	ok, err = svc.Can(context.Background(), "u1", "users:write")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected u1 to NOT have users:write")
	}

	ok, err = svc.Can(context.Background(), "unknown-user", "users:read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected unknown user to have no permissions")
	}
}

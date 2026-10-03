package items

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/platform/authctx"
	"basecode/api/internal/platform/openapigen"
)

type fakeStore struct {
	created   Input
	createdBy string
	err       error
}

func (f *fakeStore) Create(_ context.Context, ownerID string, in Input) (Item, error) {
	f.created, f.createdBy = in, ownerID
	if f.err != nil {
		return Item{}, f.err
	}
	return Item{ID: "item-1", OwnerID: ownerID, Name: in.Name, Description: in.Description, Status: in.Status, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (f *fakeStore) Get(_ context.Context, _, _ string) (Item, error) { return Item{}, f.err }
func (f *fakeStore) List(_ context.Context, _ string, _ ListFilter) ([]Item, int, error) {
	return nil, 0, f.err
}
func (f *fakeStore) Update(_ context.Context, _, _ string, _ Input) (Item, error) {
	return Item{}, f.err
}
func (f *fakeStore) Delete(_ context.Context, _, _ string) error { return f.err }

type fakeAudit struct {
	events []auditlog.Event
	err    error
}

func (f *fakeAudit) Record(_ context.Context, e auditlog.Event) error {
	f.events = append(f.events, e)
	return f.err
}

func do(h http.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1/items", strings.NewReader(body))
	req = req.WithContext(authctx.WithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestCreate_ValidatesAndScopesToCaller(t *testing.T) {
	store, audit := &fakeStore{}, &fakeAudit{}
	h := NewHandler(store, audit)

	rec := do(h.Create, http.MethodPost, `{"name":"  Laptop  ","description":"work"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body)
	}
	if store.createdBy != "user-1" {
		t.Errorf("item must be created for the caller, got owner %q", store.createdBy)
	}
	if store.created.Name != "Laptop" || store.created.Status != StatusActive {
		t.Errorf("expected trimmed name and default status, got %+v", store.created)
	}
	if len(audit.events) != 1 || audit.events[0].Action != auditlog.ActionItemCreate {
		t.Errorf("expected one item.create audit event, got %+v", audit.events)
	}

	var got openapigen.Item
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got.Id != "item-1" {
		t.Errorf("unexpected body: %s", rec.Body)
	}
}

func TestCreate_Rejects(t *testing.T) {
	cases := map[string]struct {
		body string
		code int
	}{
		"bad json":       {`{`, http.StatusBadRequest},
		"missing name":   {`{"name":"   "}`, http.StatusUnprocessableEntity},
		"unknown status": {`{"name":"x","status":"deleted"}`, http.StatusUnprocessableEntity},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			store := &fakeStore{}
			rec := do(NewHandler(store, &fakeAudit{}).Create, http.MethodPost, tc.body)
			if rec.Code != tc.code {
				t.Errorf("expected %d, got %d: %s", tc.code, rec.Code, rec.Body)
			}
			if store.createdBy != "" {
				t.Error("store must not be touched on invalid input")
			}
		})
	}
}

func TestCreate_AuditFailureIs500(t *testing.T) {
	h := NewHandler(&fakeStore{}, &fakeAudit{err: errors.New("db down")})
	if rec := do(h.Create, http.MethodPost, `{"name":"x"}`); rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when the audit write fails, got %d", rec.Code)
	}
}

func TestStoreNotFoundIs404(t *testing.T) {
	h := NewHandler(&fakeStore{err: ErrNotFound}, &fakeAudit{})

	get := func(w http.ResponseWriter, r *http.Request) { h.Get(w, r, "x") }
	del := func(w http.ResponseWriter, r *http.Request) { h.Delete(w, r, "x") }
	for name, fn := range map[string]http.HandlerFunc{"get": get, "delete": del} {
		if rec := do(fn, http.MethodGet, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", name, rec.Code)
		}
	}
}

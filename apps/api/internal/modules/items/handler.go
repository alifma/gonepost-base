package items

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/platform/authctx"
	httpx "basecode/api/internal/platform/http"
	"basecode/api/internal/platform/openapigen"
	"basecode/api/internal/platform/validator"
)

// Store is what the handler needs from the repository — defined here so
// handler tests can use a fake without a database.
type Store interface {
	Create(ctx context.Context, ownerID string, in Input) (Item, error)
	Get(ctx context.Context, ownerID, id string) (Item, error)
	List(ctx context.Context, ownerID string, f ListFilter) ([]Item, int, error)
	Update(ctx context.Context, ownerID, id string, in Input) (Item, error)
	Delete(ctx context.Context, ownerID, id string) error
}

// Recorder is the subset of auditlog.Service the handler needs.
type Recorder interface {
	Record(ctx context.Context, e auditlog.Event) error
}

type Handler struct {
	store Store
	audit Recorder
}

func NewHandler(store Store, audit Recorder) *Handler {
	return &Handler{store: store, audit: audit}
}

// itemRequest mirrors the OpenAPI ItemInput schema; the validate tags are
// the rules. Keep both in sync when a field changes.
type itemRequest struct {
	Name        string `json:"name" validate:"required,max=120"`
	Description string `json:"description" validate:"max=1000"`
	Status      string `json:"status" validate:"omitempty,oneof=active archived"`
}

// parse decodes and validates the body. On failure it has already written
// the error response and returns false.
func parse(w http.ResponseWriter, r *http.Request) (Input, bool) {
	var req itemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return Input{}, false
	}
	req.Name = strings.TrimSpace(req.Name)
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return Input{}, false
	}

	in := Input{Name: req.Name, Status: Status(req.Status)}
	if in.Status == "" {
		in.Status = StatusActive
	}
	if req.Description != "" {
		in.Description = &req.Description
	}
	return in, true
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := parse(w, r)
	if !ok {
		return
	}

	created, err := h.store.Create(r.Context(), owner(r), in)
	if err != nil {
		internalError(w)
		return
	}
	if !h.record(w, r, auditlog.ActionItemCreate, created.ID) {
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, created.Public())
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request, id string) {
	it, err := h.store.Get(r.Context(), owner(r), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, it.Public())
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{
		Status: Status(q.Get("status")),
		Search: q.Get("search"),
		Limit:  atoiDefault(q.Get("limit"), 20),
		Offset: atoiDefault(q.Get("offset"), 0),
	}

	list, total, err := h.store.List(r.Context(), owner(r), filter)
	if err != nil {
		internalError(w)
		return
	}

	resp := openapigen.ItemListResponse{Data: make([]openapigen.Item, 0, len(list)), Total: total, Limit: filter.Limit}
	for _, it := range list {
		resp.Data = append(resp.Data, it.Public())
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request, id string) {
	in, ok := parse(w, r)
	if !ok {
		return
	}

	updated, err := h.store.Update(r.Context(), owner(r), id, in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !h.record(w, r, auditlog.ActionItemUpdate, updated.ID) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, updated.Public())
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.store.Delete(r.Context(), owner(r), id); err != nil {
		writeStoreError(w, err)
		return
	}
	if !h.record(w, r, auditlog.ActionItemDelete, id) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// record writes an audit event for the caller's action. Fail-closed like
// the other modules: on failure it writes the 500 itself and returns false.
func (h *Handler) record(w http.ResponseWriter, r *http.Request, action, resourceID string) bool {
	actorID := owner(r)
	err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: &actorID,
		Action:      action,
		Resource:    "item",
		ResourceID:  &resourceID,
		Result:      auditlog.ResultSuccess,
		RequestID:   middleware.FromContext(r.Context()),
	})
	if err != nil {
		internalError(w)
		return false
	}
	return true
}

// owner is the authenticated caller. Routes are always mounted behind
// requireAuth, so the ID is present.
func owner(r *http.Request) string {
	id, _ := authctx.UserIDFromContext(r.Context())
	return id
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "item not found")
		return
	}
	internalError(w)
}

func internalError(w http.ResponseWriter) {
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if s == "" || err != nil {
		return def
	}
	return n
}

package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/platform/authctx"
	httpx "basecode/api/internal/platform/http"
	"basecode/api/internal/platform/openapigen"
	"basecode/api/internal/platform/security"
	"basecode/api/internal/platform/validator"
)

type Handler struct {
	repo  *Repository
	audit *auditlog.Service
}

func NewHandler(repo *Repository, audit *auditlog.Service) *Handler {
	return &Handler{repo: repo, audit: audit}
}

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	passwordHash, err := security.Hash(req.Password)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	u := User{
		Email:        req.Email,
		Username:     ptrOrNil(req.Username),
		FullName:     ptrOrNil(req.FullName),
		PasswordHash: passwordHash,
		Status:       StatusActive,
	}

	created, err := h.repo.Create(r.Context(), u)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httpx.WriteError(w, http.StatusConflict, "email_taken", "email is already registered")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionUserCreate, created.ID) {
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, created.Public())
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request, id string) {
	u, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u.Public())
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{
		Status: Status(q.Get("status")),
		Limit:  atoiDefault(q.Get("limit"), 20),
		Offset: atoiDefault(q.Get("offset"), 0),
	}

	list, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	resp := openapigen.UserListResponse{Data: make([]openapigen.User, 0, len(list)), Total: total, Limit: filter.Limit}
	for _, u := range list {
		resp.Data = append(resp.Data, u.Public())
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type updateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Status   string `json:"status" validate:"required,oneof=active inactive suspended"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request, id string) {
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	updated, err := h.repo.Update(r.Context(), id, ptrOrNil(req.Username), ptrOrNil(req.FullName), Status(req.Status))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionUserUpdate, updated.ID) {
		return
	}

	httpx.WriteJSON(w, http.StatusOK, updated.Public())
}

// recordActorEvent writes an audit event attributed to the currently
// authenticated caller (the admin doing the action), targeting resource
// "user" with the given resourceID. Fail-closed: on write failure, writes
// a 500 response itself and returns false so the caller stops.
func (h *Handler) recordActorEvent(w http.ResponseWriter, r *http.Request, action, resourceID string) bool {
	actorID, _ := authctx.UserIDFromContext(r.Context())
	var actorPtr *string
	if actorID != "" {
		actorPtr = &actorID
	}

	err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: actorPtr,
		Action:      action,
		Resource:    "user",
		ResourceID:  &resourceID,
		Result:      auditlog.ResultSuccess,
		RequestID:   middleware.FromContext(r.Context()),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return false
	}
	return true
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

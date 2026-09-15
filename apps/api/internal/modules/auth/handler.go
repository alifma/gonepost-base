package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	httpx "basecode/api/internal/platform/http"
	"basecode/api/internal/platform/validator"

	"basecode/api/internal/modules/users"
)

type Handler struct {
	service      *Service
	audit        *auditlog.Service
	cookieSecure bool
}

func NewHandler(service *Service, audit *auditlog.Service, cookieSecure bool) *Handler {
	return &Handler{service: service, audit: audit, cookieSecure: cookieSecure}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	rawToken, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.recordLogin(w, r, nil, req.Email, auditlog.ResultFailure)
		writeAuthError(w, err)
		return
	}

	if !h.recordLogin(w, r, &user.ID, req.Email, auditlog.ResultSuccess) {
		return
	}

	SetSessionCookie(w, rawToken, h.cookieSecure)
	httpx.WriteJSON(w, http.StatusOK, user.Public())
}

// recordLogin writes the audit event and, on a failed write, responds 500
// and returns false (fail-closed — see auditlog.Service doc). Returns true
// if the caller should proceed.
func (h *Handler) recordLogin(w http.ResponseWriter, r *http.Request, actorUserID *string, email string, result auditlog.Result) bool {
	err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: actorUserID,
		Action:      auditlog.ActionLogin,
		Resource:    "auth",
		Result:      result,
		RequestID:   middleware.FromContext(r.Context()),
		Metadata:    map[string]any{"email": email},
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return false
	}
	return true
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	rawToken, ok := ReadSessionCookie(r)
	if !ok {
		ClearSessionCookie(w, h.cookieSecure)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var actorUserID *string
	if user, err := h.service.CurrentUser(r.Context(), rawToken); err == nil {
		actorUserID = &user.ID
	}

	_ = h.service.Logout(r.Context(), rawToken)

	if err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: actorUserID,
		Action:      auditlog.ActionLogout,
		Resource:    "auth",
		Result:      auditlog.ResultSuccess,
		RequestID:   middleware.FromContext(r.Context()),
	}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	ClearSessionCookie(w, h.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	rawToken, ok := ReadSessionCookie(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "login required")
		return
	}

	user, err := h.service.CurrentUser(r.Context(), rawToken)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "session is invalid or expired")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, user.Public())
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "login required")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	result := auditlog.ResultSuccess
	changeErr := h.service.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if changeErr != nil {
		result = auditlog.ResultFailure
	}

	if err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: &userID,
		Action:      auditlog.ActionPasswordChange,
		Resource:    "user",
		ResourceID:  &userID,
		Result:      result,
		RequestID:   middleware.FromContext(r.Context()),
	}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if changeErr != nil {
		writeAuthError(w, changeErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	case errors.Is(err, ErrAccountNotActive):
		httpx.WriteError(w, http.StatusForbidden, "account_not_active", "account is not active")
	case errors.Is(err, ErrRateLimited):
		httpx.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
	case errors.Is(err, users.ErrNotFound):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
	}
}

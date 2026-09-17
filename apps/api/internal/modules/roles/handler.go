package roles

import (
	"encoding/json"
	"errors"
	"net/http"

	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/platform/authctx"
	httpx "basecode/api/internal/platform/http"
	"basecode/api/internal/platform/openapigen"
	"basecode/api/internal/platform/validator"
)

var ErrCannotRemoveLastSuperAdmin = errors.New("cannot remove the last active super admin")

type Handler struct {
	repo  *Repository
	audit *auditlog.Service
}

func NewHandler(repo *Repository, audit *auditlog.Service) *Handler {
	return &Handler{repo: repo, audit: audit}
}

func toRoleResponse(r Role) openapigen.Role {
	return openapigen.Role{Id: r.ID, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem}
}

// recordActorEvent writes an audit event attributed to the currently
// authenticated caller. Fail-closed: on write failure, responds 500 itself
// and returns false so the caller stops.
func (h *Handler) recordActorEvent(w http.ResponseWriter, r *http.Request, action, resource string, resourceID *string) bool {
	actorID, _ := authctx.UserIDFromContext(r.Context())
	var actorPtr *string
	if actorID != "" {
		actorPtr = &actorID
	}

	err := h.audit.Record(r.Context(), auditlog.Event{
		ActorUserID: actorPtr,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Result:      auditlog.ResultSuccess,
		RequestID:   middleware.FromContext(r.Context()),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return false
	}
	return true
}

type createRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	role, err := h.repo.Create(r.Context(), req.Name, ptrOrNil(req.Description))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionRoleCreate, "role", &role.ID) {
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toRoleResponse(role))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.List(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	resp := make([]openapigen.Role, 0, len(list))
	for _, role := range list {
		resp = append(resp, toRoleResponse(role))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type grantPermissionRequest struct {
	Code string `json:"code" validate:"required"`
}

func (h *Handler) GrantPermission(w http.ResponseWriter, r *http.Request, roleID string) {
	var req grantPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	perm, err := h.repo.GetPermissionByCode(r.Context(), req.Code)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "permission not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if err := h.repo.GrantPermission(r.Context(), roleID, perm.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionPermissionGrant, "role", &roleID) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request, roleID string) {
	codes, err := h.repo.ListRolePermissions(r.Context(), roleID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, codes)
}

func (h *Handler) RevokePermission(w http.ResponseWriter, r *http.Request, roleID, code string) {
	perm, err := h.repo.GetPermissionByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "permission not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	if err := h.repo.RevokePermission(r.Context(), roleID, perm.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	if !h.recordActorEvent(w, r, auditlog.ActionPermissionRevoke, "role", &roleID) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type assignRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}

func (h *Handler) AssignToUser(w http.ResponseWriter, r *http.Request, userID string) {
	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		httpx.WriteValidationError(w, errs)
		return
	}

	if err := h.repo.AssignToUser(r.Context(), userID, req.RoleID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionUserRoleAssign, "user", &userID) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveFromUser(w http.ResponseWriter, r *http.Request, userID, roleID string) {
	role, err := h.repo.GetByID(r.Context(), roleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "role not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if role.Name == SuperAdminRole {
		count, err := h.repo.CountUsersWithRole(r.Context(), role.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
			return
		}
		if count <= 1 {
			httpx.WriteError(w, http.StatusConflict, "last_super_admin", ErrCannotRemoveLastSuperAdmin.Error())
			return
		}
	}

	if err := h.repo.RemoveFromUser(r.Context(), userID, roleID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionUserRoleRemove, "user", &userID) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	list, err := h.repo.ListUserRoles(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	resp := make([]openapigen.Role, 0, len(list))
	for _, role := range list {
		resp = append(resp, toRoleResponse(role))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// BootstrapSuperAdmin grants SUPER_ADMIN to the calling (already
// authenticated) user, but only if no one holds it yet. This is the
// "protected bootstrap path" — protected by both requiring a logged-in
// user (RequireAuth) and by the one-time-only condition below, not by a
// permission check (there's no super-admin yet to have granted one).
func (h *Handler) BootstrapSuperAdmin(w http.ResponseWriter, r *http.Request, userID string) {
	role, err := h.repo.GetByName(r.Context(), SuperAdminRole)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "super admin role is not seeded")
		return
	}

	count, err := h.repo.CountUsersWithRole(r.Context(), role.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}
	if count > 0 {
		httpx.WriteError(w, http.StatusForbidden, "already_bootstrapped", "a super admin already exists")
		return
	}

	if err := h.repo.AssignToUser(r.Context(), userID, role.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	if !h.recordActorEvent(w, r, auditlog.ActionBootstrapAdmin, "user", &userID) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

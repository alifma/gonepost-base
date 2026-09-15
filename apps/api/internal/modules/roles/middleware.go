package roles

import (
	"net/http"

	"basecode/api/internal/middleware"
	"basecode/api/internal/modules/auditlog"
	"basecode/api/internal/platform/authctx"
	httpx "basecode/api/internal/platform/http"
)

// RequirePermission is a middleware factory. It must run AFTER
// auth.RequireAuth in the chain — it reads the user ID that middleware put
// in the request context, so mount it as the inner wrapper:
//
//	requireAuth(requirePermission(handler))
func RequirePermission(authz *AuthzService, audit *auditlog.Service, permissionCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := authctx.UserIDFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "login required")
				return
			}

			allowed, err := authz.Can(r.Context(), userID, permissionCode)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
				return
			}
			if !allowed {
				// Best-effort audit — a denied-permission event failing to
				// write must not block the 403 the caller already deserves,
				// so this one is NOT fail-closed like mutation audits are.
				_ = audit.Record(r.Context(), auditlog.Event{
					ActorUserID: &userID,
					Action:      auditlog.ActionAuthzDenied,
					Resource:    "authz",
					Result:      auditlog.ResultFailure,
					RequestID:   middleware.FromContext(r.Context()),
					Metadata:    map[string]any{"permission": permissionCode, "path": r.URL.Path},
				})
				httpx.WriteError(w, http.StatusForbidden, "forbidden", "you don't have permission to do this")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

package auth

import (
	"context"
	"net/http"

	"basecode/api/internal/platform/authctx"
	httpx "basecode/api/internal/platform/http"
)

// RequireAuth is a middleware factory: it needs a *Service to validate the
// session cookie, so (like middleware.Recover/CORS) it's a function that
// returns a middleware rather than being one directly.
func RequireAuth(service *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := ReadSessionCookie(r)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "login required")
				return
			}

			user, err := service.CurrentUser(r.Context(), rawToken)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "session is invalid or expired")
				return
			}

			ctx := authctx.WithUserID(r.Context(), user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext reads the authenticated user's ID set by RequireAuth.
// Thin re-export of authctx so existing call sites don't need to import
// two packages.
func UserIDFromContext(ctx context.Context) (string, bool) {
	return authctx.UserIDFromContext(ctx)
}

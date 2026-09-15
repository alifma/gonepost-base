package middleware

import (
	httpx "basecode/api/internal/platform/http"
	"log/slog"
	"net/http"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "error", rec, "request_id", FromContext(r.Context()))
					httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

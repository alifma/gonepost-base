// Package authctx holds the request-context key for "who is the
// authenticated caller" — shared infra, not domain logic, so any module
// (auth, users, roles, auditlog) can read it without depending on each
// other and risking an import cycle (auth already depends on users for
// its repository interfaces; users depending back on auth would cycle).
package authctx

import "context"

type ctxKey string

const userIDKey ctxKey = "authUserID"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

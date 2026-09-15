package users

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"basecode/api/internal/platform/openapigen"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

// User is the full internal representation, including PasswordHash.
// Never serialize this directly to JSON — use Public() instead.
type User struct {
	ID           string
	Email        string
	Username     *string
	FullName     *string
	PasswordHash string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Public converts to the OpenAPI-generated User type — the same shape the
// TypeScript client expects, generated from the same spec. Never a
// hand-duplicated response struct: if the contract changes, this fails to
// compile until it's updated to match.
func (u User) Public() openapigen.User {
	return openapigen.User{
		Id:        u.ID,
		Email:     openapi_types.Email(u.Email),
		Username:  u.Username,
		FullName:  u.FullName,
		Status:    openapigen.UserStatus(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

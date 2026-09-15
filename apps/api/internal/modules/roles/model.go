package roles

import "time"

const SuperAdminRole = "SUPER_ADMIN"

type Role struct {
	ID          string
	Name        string
	Description *string
	IsSystem    bool
	CreatedAt   time.Time
}

type Permission struct {
	ID          string
	Code        string
	Description *string
}

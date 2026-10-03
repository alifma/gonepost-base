package roles

import "time"

const SuperAdminRole = "SUPER_ADMIN"

// MemberRole is the everyday role: use the app's features on the member's
// own data, no user or role administration. Seeded with the feature
// permissions (see seeds/main.go).
const MemberRole = "MEMBER"

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

package permissions

// Permission codes follow "resource:action". This list is the source of
// truth — the seed script inserts these into the `permissions` table, and
// handlers reference these constants (never raw strings) when calling
// authz.HasPermission, so a typo fails to compile instead of silently
// always denying/allowing access.
const (
	UsersRead  = "users:read"
	UsersWrite = "users:write"
	RolesRead  = "roles:read"
	RolesWrite = "roles:write"
	AuditRead  = "audit:read"
)

// All is every known permission code — used by the seed script to make sure
// every permission constant above actually exists as a row, and by the
// SUPER_ADMIN role seed to grant all of them.
var All = []string{
	UsersRead,
	UsersWrite,
	RolesRead,
	RolesWrite,
	AuditRead,
}

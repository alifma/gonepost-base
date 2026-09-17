package auditlog

import "time"

type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
)

// Action codes: "resource.verb", parallel to permission codes but for
// events, not access checks.
const (
	ActionLogin            = "auth.login"
	ActionLogout           = "auth.logout"
	ActionPasswordChange   = "auth.password_change"
	ActionBootstrapAdmin   = "auth.bootstrap_admin"
	ActionUserCreate       = "user.create"
	ActionUserUpdate       = "user.update"
	ActionRoleCreate       = "role.create"
	ActionPermissionGrant  = "role.permission_grant"
	ActionPermissionRevoke = "role.permission_revoke"
	ActionUserRoleAssign   = "role.user_assign"
	ActionUserRoleRemove   = "role.user_remove"
	ActionAuthzDenied      = "authz.denied"
)

type Event struct {
	ID          string
	ActorUserID *string
	Action      string
	Resource    string
	ResourceID  *string
	Result      Result
	RequestID   string
	TraceID     string
	Metadata    map[string]any
	CreatedAt   time.Time
}

type ListFilter struct {
	Action      string
	ActorUserID string
	Limit       int
	Offset      int
}

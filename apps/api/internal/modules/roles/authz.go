package roles

import "context"

// PermissionChecker is the subset of Repository the authz service needs —
// defined here so tests can supply a fake without a real database.
type PermissionChecker interface {
	UserHasPermission(ctx context.Context, userID, permissionCode string) (bool, error)
}

// AuthzService is the single place permission decisions go through.
// Handlers/services never query role_permissions/user_roles directly.
type AuthzService struct {
	checker PermissionChecker
}

func NewAuthzService(checker PermissionChecker) *AuthzService {
	return &AuthzService{checker: checker}
}

func (s *AuthzService) Can(ctx context.Context, userID, permissionCode string) (bool, error) {
	return s.checker.UserHasPermission(ctx, userID, permissionCode)
}

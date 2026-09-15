package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"basecode/api/internal/modules/users"
	"basecode/api/internal/platform/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountNotActive   = errors.New("account is not active")
	ErrRateLimited        = errors.New("too many login attempts, try again later")
)

// UserRepository is the subset of users.Repository the auth service needs.
// Defined here (not in users package) so tests can supply a fake without
// touching a real database.
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (users.User, error)
	GetByID(ctx context.Context, id string) (users.User, error)
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
}

type SessionStore interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error)
	GetActiveByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
}

type Service struct {
	users      UserRepository
	sessions   SessionStore
	loginLimit *RateLimiter
	now        func() time.Time
	sessionTTL time.Duration
}

func NewService(users UserRepository, sessions SessionStore, loginLimit *RateLimiter) *Service {
	return &Service{
		users:      users,
		sessions:   sessions,
		loginLimit: loginLimit,
		now:        time.Now,
		sessionTTL: SessionDuration,
	}
}

// Login verifies credentials and, on success, creates a session and returns
// the raw session token (to be set as a cookie by the HTTP handler).
func (s *Service) Login(ctx context.Context, email, password string) (rawToken string, user users.User, err error) {
	normalizedEmail := normalizeEmail(email)
	if s.loginLimit != nil && !s.loginLimit.Allowed(normalizedEmail) {
		return "", users.User{}, ErrRateLimited
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			s.recordLoginFailure(normalizedEmail)
			return "", users.User{}, ErrInvalidCredentials
		}
		return "", users.User{}, err
	}

	ok, err := security.Verify(password, u.PasswordHash)
	if err != nil {
		return "", users.User{}, err
	}
	if !ok {
		s.recordLoginFailure(normalizedEmail)
		return "", users.User{}, ErrInvalidCredentials
	}

	if u.Status != users.StatusActive {
		s.recordLoginFailure(normalizedEmail)
		return "", users.User{}, ErrAccountNotActive
	}

	raw, hash, err := newSessionToken()
	if err != nil {
		return "", users.User{}, err
	}

	if _, err := s.sessions.Create(ctx, u.ID, hash, s.now().Add(s.sessionTTL)); err != nil {
		return "", users.User{}, err
	}

	return raw, u, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	return s.sessions.RevokeByTokenHash(ctx, hashToken(rawToken))
}

// CurrentUser resolves a raw session token into the user it belongs to.
func (s *Service) CurrentUser(ctx context.Context, rawToken string) (users.User, error) {
	sess, err := s.sessions.GetActiveByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		return users.User{}, err
	}
	return s.users.GetByID(ctx, sess.UserID)
}

func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	ok, err := security.Verify(oldPassword, u.PasswordHash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidCredentials
	}

	newHash, err := security.Hash(newPassword)
	if err != nil {
		return err
	}

	return s.users.UpdatePasswordHash(ctx, userID, newHash)
}

func (s *Service) recordLoginFailure(normalizedEmail string) {
	if s.loginLimit != nil {
		s.loginLimit.RecordFailure(normalizedEmail)
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

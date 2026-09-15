package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"basecode/api/internal/modules/users"
	"basecode/api/internal/platform/security"
)

type fakeUserRepo struct {
	byEmail map[string]users.User
	byID    map[string]users.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]users.User{}, byID: map[string]users.User{}}
}

func (f *fakeUserRepo) add(u users.User) {
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (users.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return users.User{}, users.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id string) (users.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return users.User{}, users.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) UpdatePasswordHash(_ context.Context, id, hash string) error {
	u, ok := f.byID[id]
	if !ok {
		return users.ErrNotFound
	}
	u.PasswordHash = hash
	f.add(u)
	return nil
}

type fakeSessionStore struct {
	byHash map[string]Session
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{byHash: map[string]Session{}}
}

func (f *fakeSessionStore) Create(_ context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error) {
	s := Session{ID: "sess-1", UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	f.byHash[tokenHash] = s
	return s, nil
}

func (f *fakeSessionStore) GetActiveByTokenHash(_ context.Context, tokenHash string) (Session, error) {
	s, ok := f.byHash[tokenHash]
	if !ok || (s.RevokedAt != nil) || time.Now().After(s.ExpiresAt) {
		return Session{}, ErrSessionNotFound
	}
	return s, nil
}

func (f *fakeSessionStore) RevokeByTokenHash(_ context.Context, tokenHash string) error {
	s, ok := f.byHash[tokenHash]
	if !ok {
		return nil
	}
	now := time.Now()
	s.RevokedAt = &now
	f.byHash[tokenHash] = s
	return nil
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	h, err := security.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return h
}

func TestService_Login_Success(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusActive})
	svc := NewService(userRepo, newFakeSessionStore(), nil)

	rawToken, u, err := svc.Login(context.Background(), "a@b.com", "correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rawToken == "" {
		t.Error("expected non-empty raw token")
	}
	if u.ID != "u1" {
		t.Errorf("expected user u1, got %s", u.ID)
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusActive})
	svc := NewService(userRepo, newFakeSessionStore(), nil)

	_, _, err := svc.Login(context.Background(), "a@b.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Login_UnknownEmail(t *testing.T) {
	svc := NewService(newFakeUserRepo(), newFakeSessionStore(), nil)

	_, _, err := svc.Login(context.Background(), "nobody@b.com", "whatever")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials (not leaking whether email exists), got %v", err)
	}
}

func TestService_Login_InactiveAccount(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusSuspended})
	svc := NewService(userRepo, newFakeSessionStore(), nil)

	_, _, err := svc.Login(context.Background(), "a@b.com", "correct-password")
	if !errors.Is(err, ErrAccountNotActive) {
		t.Errorf("expected ErrAccountNotActive, got %v", err)
	}
}

func TestService_Login_RateLimited(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusActive})
	limiter := NewRateLimiter(2, time.Minute)
	svc := NewService(userRepo, newFakeSessionStore(), limiter)

	svc.Login(context.Background(), "a@b.com", "wrong")
	svc.Login(context.Background(), "a@b.com", "wrong")
	_, _, err := svc.Login(context.Background(), "a@b.com", "correct-password")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited on 3rd attempt, got %v", err)
	}
}

// Regression: successful logins must NOT consume rate-limit quota — only
// failures should. A limit of 2 should allow far more than 2 successful
// logins in a row (e.g. repeated testing, multiple tabs/devices).
func TestService_Login_SuccessDoesNotConsumeRateLimit(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusActive})
	limiter := NewRateLimiter(2, time.Minute)
	svc := NewService(userRepo, newFakeSessionStore(), limiter)

	for i := 0; i < 5; i++ {
		if _, _, err := svc.Login(context.Background(), "a@b.com", "correct-password"); err != nil {
			t.Fatalf("login %d: expected success, got %v", i+1, err)
		}
	}
}

func TestService_LoginLogoutCurrentUser(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "correct-password"), Status: users.StatusActive})
	svc := NewService(userRepo, newFakeSessionStore(), nil)

	rawToken, _, err := svc.Login(context.Background(), "a@b.com", "correct-password")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	u, err := svc.CurrentUser(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("CurrentUser failed: %v", err)
	}
	if u.ID != "u1" {
		t.Errorf("expected u1, got %s", u.ID)
	}

	if err := svc.Logout(context.Background(), rawToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if _, err := svc.CurrentUser(context.Background(), rawToken); err == nil {
		t.Error("expected CurrentUser to fail after logout")
	}
}

func TestService_ChangePassword(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.add(users.User{ID: "u1", Email: "a@b.com", PasswordHash: mustHash(t, "old-password"), Status: users.StatusActive})
	svc := NewService(userRepo, newFakeSessionStore(), nil)

	if err := svc.ChangePassword(context.Background(), "u1", "old-password", "new-password"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, _, err := svc.Login(context.Background(), "a@b.com", "new-password"); err != nil {
		t.Errorf("expected login with new password to succeed, got %v", err)
	}
	if _, _, err := svc.Login(context.Background(), "a@b.com", "old-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected old password to no longer work, got %v", err)
	}
}

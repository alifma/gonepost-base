package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const SessionDuration = 7 * 24 * time.Hour

// Session is the DB-side record. TokenHash is stored, never the raw token.
type Session struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// newSessionToken generates a random opaque token (returned to the client)
// and its SHA-256 hash (stored in the database).
func newSessionToken() (rawToken, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(buf)
	tokenHash = hashToken(rawToken)
	return rawToken, tokenHash, nil
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

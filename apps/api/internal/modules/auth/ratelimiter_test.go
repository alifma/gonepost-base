package auth

import (
	"testing"
	"time"
)

// helper matching the real usage pattern: check Allowed(), and only
// RecordFailure() when simulating a failed attempt.
func attemptFailure(rl *RateLimiter, key string) bool {
	if !rl.Allowed(key) {
		return false
	}
	rl.RecordFailure(key)
	return true
}

func TestRateLimiter_AllowsUpToLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !attemptFailure(rl, "key") {
			t.Fatalf("expected attempt %d to be allowed", i+1)
		}
	}
	if attemptFailure(rl, "key") {
		t.Error("expected 4th attempt to be denied")
	}
}

func TestRateLimiter_SeparateKeysIndependent(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !attemptFailure(rl, "a") {
		t.Error("expected first attempt for key a to be allowed")
	}
	if !attemptFailure(rl, "b") {
		t.Error("expected first attempt for key b to be allowed (independent of a)")
	}
	if attemptFailure(rl, "a") {
		t.Error("expected second attempt for key a to be denied")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !attemptFailure(rl, "key") {
		t.Fatal("expected first attempt to be allowed")
	}
	if attemptFailure(rl, "key") {
		t.Fatal("expected second immediate attempt to be denied")
	}

	time.Sleep(60 * time.Millisecond)

	if !attemptFailure(rl, "key") {
		t.Error("expected attempt after window to be allowed again")
	}
}

func TestRateLimiter_AllowedDoesNotRecord(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	// Checking Allowed() repeatedly must not itself consume the quota —
	// only RecordFailure() should.
	for i := 0; i < 5; i++ {
		if !rl.Allowed("key") {
			t.Fatalf("Allowed() call %d should not consume quota by itself", i+1)
		}
	}
}

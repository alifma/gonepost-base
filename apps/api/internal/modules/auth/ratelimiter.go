package auth

import (
	"sync"
	"time"
)

// RateLimiter is a simple in-memory fixed-window limiter, keyed by a string
// (e.g. normalized email or client IP). Good enough for a single-instance
// deployment; once the API runs behind more than one instance, this needs
// to move to a shared store (Redis — see PLAN.md Deferred Work).
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allowed reports whether key is currently within the limit, WITHOUT
// recording a new attempt. Call this before doing the thing being
// rate-limited (e.g. before checking a password).
func (rl *RateLimiter) Allowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	kept := rl.prune(key)
	rl.attempts[key] = kept
	return len(kept) < rl.limit
}

// RecordFailure counts one failed attempt against key. Only call this on
// failure — a rate limiter that also counts successes would penalize
// legitimate repeated logins (e.g. multiple tabs/devices, or just testing),
// not just brute-force attempts.
func (rl *RateLimiter) RecordFailure(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	kept := rl.prune(key)
	rl.attempts[key] = append(kept, time.Now())
}

// prune drops attempts older than the window. Caller must hold rl.mu.
func (rl *RateLimiter) prune(key string) []time.Time {
	cutoff := time.Now().Add(-rl.window)
	kept := rl.attempts[key][:0]
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}

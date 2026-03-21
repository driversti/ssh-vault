package hub

import (
	"sync"
	"time"
)

// RateLimiter implements a per-IP sliding window rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

// NewRateLimiter creates a rate limiter with the given window duration and request limit.
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

// Allow checks whether a request from the given IP is allowed.
// It prunes expired entries and adds the current request if within the limit.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Prune expired entries for this IP
	timestamps := rl.requests[ip]
	valid := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)

	// Lazy cleanup: remove stale IPs (no requests in >5 minutes)
	staleThreshold := now.Add(-5 * time.Minute)
	for key, ts := range rl.requests {
		if key == ip {
			continue
		}
		if len(ts) == 0 || ts[len(ts)-1].Before(staleThreshold) {
			delete(rl.requests, key)
		}
	}

	return true
}

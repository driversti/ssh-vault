package hub

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 3)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("192.168.1.1") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiter_IPIsolation(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 2)

	// Exhaust limit for IP A
	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Error("IP A should be rate-limited")
	}

	// IP B should still be allowed
	if !rl.Allow("10.0.0.2") {
		t.Error("IP B should not be affected by IP A's limit")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 2)

	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Error("should be rate-limited before window expires")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("10.0.0.1") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimiter_StaleCleanup(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 10)

	// Add entries for a stale IP by manipulating the internal state
	rl.mu.Lock()
	staleTime := time.Now().Add(-10 * time.Minute)
	rl.requests["stale-ip"] = []time.Time{staleTime}
	rl.mu.Unlock()

	// A request from a different IP should trigger cleanup
	rl.Allow("fresh-ip")

	rl.mu.Lock()
	_, exists := rl.requests["stale-ip"]
	rl.mu.Unlock()

	if exists {
		t.Error("stale IP entry should have been cleaned up")
	}
}

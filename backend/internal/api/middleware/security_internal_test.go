package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimiter_CleanupOnce_RemovesStaleEntries exercises the sweep logic that
// the background cleanup goroutine runs periodically, without waiting on the
// real 5-minute ticker interval.
func TestRateLimiter_CleanupOnce_RemovesStaleEntries(t *testing.T) {
	rl := &RateLimiter{limit: 5, window: 10 * time.Millisecond}

	// Stale entry: window started long enough ago to exceed window*2.
	rl.entries.Store("stale-key", &rateLimitEntry{
		count:       1,
		windowStart: time.Now().Add(-time.Hour),
	})
	// Fresh entry: should survive the sweep.
	rl.entries.Store("fresh-key", &rateLimitEntry{
		count:       1,
		windowStart: time.Now(),
	})

	rl.cleanupOnce()

	_, staleStillPresent := rl.entries.Load("stale-key")
	_, freshStillPresent := rl.entries.Load("fresh-key")

	assert.False(t, staleStillPresent, "stale entry should have been removed")
	assert.True(t, freshStillPresent, "fresh entry should not have been removed")
}

package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_CleanupOnce_RemovesStaleEntries(t *testing.T) {
	rl := NewRateLimiter(5, 10*time.Millisecond)

	rl.Allow("stale-key")
	rl.Allow("fresh-key")

	// Backdate the "stale-key" entry's window so it's well past 2x the window.
	val, _ := rl.entries.Load("stale-key")
	entry := val.(*rateLimitEntry)
	entry.mu.Lock()
	entry.windowStart = time.Now().Add(-time.Hour)
	entry.mu.Unlock()

	rl.cleanupOnce()

	_, staleStillPresent := rl.entries.Load("stale-key")
	_, freshStillPresent := rl.entries.Load("fresh-key")
	assert.False(t, staleStillPresent, "stale entry should have been removed")
	assert.True(t, freshStillPresent, "fresh entry should be kept")
}

func TestRateLimiter_CleanupOnce_NoEntries(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	assert.NotPanics(t, func() {
		rl.cleanupOnce()
	})
}

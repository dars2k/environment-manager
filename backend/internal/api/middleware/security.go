package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// SecurityHeadersMiddleware adds security-related HTTP headers to every response.
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitEntry tracks the request count for a single key in a fixed window.
type rateLimitEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

// RateLimiter implements a simple fixed-window rate limiter keyed by client IP.
type RateLimiter struct {
	entries sync.Map
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a RateLimiter that allows at most limit requests per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{limit: limit, window: window}
	go rl.cleanup()
	return rl
}

// Allow returns true when the key is within the rate limit for the current window.
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	val, _ := rl.entries.LoadOrStore(key, &rateLimitEntry{windowStart: now})
	entry := val.(*rateLimitEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if now.Sub(entry.windowStart) >= rl.window {
		entry.count = 0
		entry.windowStart = now
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	return true
}

// cleanup removes stale entries to prevent unbounded memory growth.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.entries.Range(func(k, v interface{}) bool {
			e := v.(*rateLimitEntry)
			e.mu.Lock()
			stale := now.Sub(e.windowStart) > rl.window*2
			e.mu.Unlock()
			if stale {
				rl.entries.Delete(k)
			}
			return true
		})
	}
}

// RateLimitMiddleware wraps a handler with IP-based rate limiting.
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			if !rl.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too many requests. Please try again later."}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP returns the client IP from RemoteAddr, ignoring forwarded headers
// to prevent spoofing. Normalises IPv6 loopback to 127.0.0.1.
func extractClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}

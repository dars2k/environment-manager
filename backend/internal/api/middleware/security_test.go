package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.SecurityHeadersMiddleware()(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := middleware.NewRateLimiter(3, time.Minute)

	// First 3 requests should be allowed
	assert.True(t, rl.Allow("1.2.3.4"))
	assert.True(t, rl.Allow("1.2.3.4"))
	assert.True(t, rl.Allow("1.2.3.4"))

	// 4th request should be denied
	assert.False(t, rl.Allow("1.2.3.4"))

	// Different IP should still be allowed
	assert.True(t, rl.Allow("5.6.7.8"))
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 10*time.Millisecond)

	assert.True(t, rl.Allow("10.0.0.1"))
	assert.True(t, rl.Allow("10.0.0.1"))
	assert.False(t, rl.Allow("10.0.0.1"))

	// Wait for window to expire
	time.Sleep(15 * time.Millisecond)

	// Should be allowed again in the new window
	assert.True(t, rl.Allow("10.0.0.1"))
}

func TestRateLimitMiddleware_Allows(t *testing.T) {
	rl := middleware.NewRateLimiter(10, time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	wrapped := middleware.RateLimitMiddleware(rl)(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRateLimitMiddleware_Blocks(t *testing.T) {
	// Use a limiter that blocks after 1 request
	rl := middleware.NewRateLimiter(1, time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.RateLimitMiddleware(rl)(handler)

	ip := "10.10.10.10:9999"

	// First request allowed
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = ip
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = ip
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Equal(t, "application/json", w2.Header().Get("Content-Type"))
	assert.Equal(t, "60", w2.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_IPv6Loopback(t *testing.T) {
	rl := middleware.NewRateLimiter(10, time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.RateLimitMiddleware(rl)(handler)

	// IPv6 loopback should be normalised to 127.0.0.1
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[::1]:1234"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_MalformedRemoteAddr(t *testing.T) {
	rl := middleware.NewRateLimiter(10, time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.RateLimitMiddleware(rl)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "not-an-ip" // no port, SplitHostPort will fail
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	// Should still work, just uses the raw addr as key
	assert.Equal(t, http.StatusOK, w.Code)
}

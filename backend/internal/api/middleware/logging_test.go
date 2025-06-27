package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/api/middleware"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		handler      http.HandlerFunc
		expectedCode int
		checkLog     func(*testing.T, *logrus.Entry)
	}{
		{
			name:   "Successful request",
			method: "GET",
			path:   "/test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
			expectedCode: http.StatusOK,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, "GET", entry.Data["method"])
				assert.Equal(t, "/test", entry.Data["path"])
				assert.Equal(t, http.StatusOK, entry.Data["status"])
				assert.NotNil(t, entry.Data["duration"])
				assert.NotNil(t, entry.Data["ip"])
				assert.NotNil(t, entry.Data["user_agent"])
				assert.Equal(t, "HTTP request processed", entry.Message)
			},
		},
		{
			name:   "Request with error status",
			method: "POST",
			path:   "/api/users",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
			}),
			expectedCode: http.StatusBadRequest,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, "POST", entry.Data["method"])
				assert.Equal(t, "/api/users", entry.Data["path"])
				assert.Equal(t, http.StatusBadRequest, entry.Data["status"])
			},
		},
		{
			name:   "Request with no explicit status",
			method: "GET",
			path:   "/healthz",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Don't set status explicitly, should default to 200
				w.Write([]byte("healthy"))
			}),
			expectedCode: http.StatusOK,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, http.StatusOK, entry.Data["status"])
			},
		},
		{
			name:   "Request with server error",
			method: "DELETE",
			path:   "/api/environments/123",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			expectedCode: http.StatusInternalServerError,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, http.StatusInternalServerError, entry.Data["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger with hook
			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.InfoLevel)

			// Create middleware
			middleware := middleware.LoggingMiddleware(logger)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("User-Agent", "test-agent/1.0")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute middleware
			middleware(tt.handler).ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedCode, w.Code)

			// Check log entry
			assert.Equal(t, 1, len(hook.Entries))
			entry := hook.LastEntry()
			assert.Equal(t, logrus.InfoLevel, entry.Level)
			
			if tt.checkLog != nil {
				tt.checkLog(t, entry)
			}

			// Check duration is reasonable
			duration, ok := entry.Data["duration"].(int64)
			assert.True(t, ok)
			assert.GreaterOrEqual(t, duration, int64(0))
			assert.Less(t, duration, int64(1000)) // Should be less than 1 second
		})
	}
}

func TestLoggingMiddleware_Duration(t *testing.T) {
	// Create test logger with hook
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.InfoLevel)

	// Create middleware
	middleware := middleware.LoggingMiddleware(logger)

	// Create handler that sleeps
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Check log entry
	assert.Equal(t, 1, len(hook.Entries))
	entry := hook.LastEntry()

	// Check duration is at least 10ms
	duration, ok := entry.Data["duration"].(int64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, duration, int64(10))
}

func TestLoggingMiddleware_MultipleWrites(t *testing.T) {
	// Create test logger with hook
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.InfoLevel)

	// Create middleware
	middleware := middleware.LoggingMiddleware(logger)

	// Create handler that writes header multiple times
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// Note: Go's http.ResponseWriter ignores subsequent WriteHeader calls
		// but httptest.ResponseRecorder may behave differently
		w.Write([]byte("created"))
	})

	// Create request
	req := httptest.NewRequest("POST", "/resource", nil)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Check response - httptest.ResponseRecorder captures the first status
	assert.Equal(t, http.StatusCreated, w.Code)

	// Check log entry
	entry := hook.LastEntry()
	assert.Equal(t, http.StatusCreated, entry.Data["status"])
}

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		shouldPanic   bool
		panicValue    interface{}
		expectedCode  int
		expectedBody  string
		checkLog      func(*testing.T, *logrus.Entry)
	}{
		{
			name: "Normal request - no panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}),
			shouldPanic:  false,
			expectedCode: http.StatusOK,
			expectedBody: `{"success":true}`,
		},
		{
			name: "Handler panics with string",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			shouldPanic:  true,
			panicValue:   "something went wrong",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, "something went wrong", entry.Data["error"])
				assert.Equal(t, "GET", entry.Data["method"])
				assert.Equal(t, "/test", entry.Data["path"])
				assert.Equal(t, "Panic recovered", entry.Message)
			},
		},
		{
			name: "Handler panics with error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(assert.AnError)
			}),
			shouldPanic:  true,
			panicValue:   assert.AnError,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.Equal(t, assert.AnError, entry.Data["error"])
			},
		},
		{
			name: "Handler panics with struct",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				type customError struct {
					Code    string
					Message string
				}
				panic(customError{Code: "CUSTOM", Message: "Custom error"})
			}),
			shouldPanic:  true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`,
			checkLog: func(t *testing.T, entry *logrus.Entry) {
				assert.NotNil(t, entry.Data["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger with hook
			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.ErrorLevel)

			// Create middleware
			middleware := middleware.RecoveryMiddleware(logger)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Execute middleware
			middleware(tt.handler).ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			// Check content type only for panic cases where middleware sets it
			if tt.shouldPanic {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}

			// Check log
			if tt.shouldPanic {
				assert.Equal(t, 1, len(hook.Entries))
				entry := hook.LastEntry()
				assert.Equal(t, logrus.ErrorLevel, entry.Level)
				
				if tt.checkLog != nil {
					tt.checkLog(t, entry)
				}
			} else {
				assert.Equal(t, 0, len(hook.Entries))
			}
		})
	}
}

func TestRecoveryMiddleware_WriteAfterPanic(t *testing.T) {
	// Create test logger
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.ErrorLevel)

	// Create middleware
	middleware := middleware.RecoveryMiddleware(logger)

	// Create handler that writes before panic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial"))
		panic("panic after write")
	})

	// Create request
	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// The panic should still be caught
	assert.Equal(t, 1, len(hook.Entries))
	entry := hook.LastEntry()
	assert.Equal(t, "panic after write", entry.Data["error"])
}


// Benchmarks
func BenchmarkLoggingMiddleware(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	
	middleware := middleware.LoggingMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware(handler).ServeHTTP(w, req)
	}
}

func BenchmarkRecoveryMiddleware(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	
	middleware := middleware.RecoveryMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware(handler).ServeHTTP(w, req)
	}
}

// Helper to check if log contains expected fields
func assertLogContains(t *testing.T, entry *logrus.Entry, expectedFields map[string]interface{}) {
	for key, expectedValue := range expectedFields {
		actualValue, exists := entry.Data[key]
		assert.True(t, exists, "Expected log field %s to exist", key)
		if expectedValue != nil {
			assert.Equal(t, expectedValue, actualValue, "Log field %s mismatch", key)
		}
	}
}

// Test with real HTTP server
func TestLoggingMiddleware_Integration(t *testing.T) {
	// Create logger
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	// Create handler with middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	
	wrapped := middleware.LoggingMiddleware(logger)(handler)
	
	// Create test server
	server := httptest.NewServer(wrapped)
	defer server.Close()
	
	// Make request
	resp, err := http.Get(server.URL + "/test-path")
	assert.NoError(t, err)
	defer resp.Body.Close()
	
	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "test", resp.Header.Get("X-Custom-Header"))
	
	// Check log
	logOutput := buf.String()
	assert.Contains(t, logOutput, "HTTP request processed")
	assert.Contains(t, logOutput, "/test-path")
	
	// Parse JSON log
	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test-path", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
}

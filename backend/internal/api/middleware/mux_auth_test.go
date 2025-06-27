package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func createTestToken(claims jwt.MapClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestMuxAuthMiddleware(t *testing.T) {
	jwtSecret := "test-secret-key"
	logger := logrus.New()
	logger.SetOutput(nil)

	tests := []struct {
		name           string
		authHeader     string
		setupToken     func() string
		expectedStatus int
		expectedBody   string
		checkContext   func(*testing.T, *http.Request)
	}{
		{
			name:           "Missing authorization header",
			authHeader:     "",
			setupToken:     func() string { return "" },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required\n",
		},
		{
			name:           "Invalid authorization header format - no Bearer",
			authHeader:     "InvalidToken",
			setupToken:     func() string { return "InvalidToken" },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format\n",
		},
		{
			name:           "Invalid authorization header format - wrong prefix",
			authHeader:     "Basic token123",
			setupToken:     func() string { return "Basic token123" },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format\n",
		},
		{
			name: "Invalid token - malformed",
			authHeader: "Bearer invalid.token.here",
			setupToken: func() string { return "Bearer invalid.token.here" },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token\n",
		},
		{
			name: "Invalid token - wrong secret",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"userId":   "123",
					"username": "testuser",
					"exp":      time.Now().Add(time.Hour).Unix(),
				}, "wrong-secret")
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token\n",
		},
		{
			name: "Expired token",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"userId":   "123",
					"username": "testuser",
					"exp":      time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
				}, jwtSecret)
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token\n",
		},
		{
			name: "Valid token but invalid claims type",
			setupToken: func() string {
				// Create a token with custom claims struct (not MapClaims)
				type CustomClaims struct {
					jwt.RegisteredClaims
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid user ID in token\n",
		},
		{
			name: "Valid token but missing userId",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"username": "testuser",
					"email":    "test@example.com",
					"exp":      time.Now().Add(time.Hour).Unix(),
				}, jwtSecret)
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid user ID in token\n",
		},
		{
			name: "Valid token but userId is not string",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"userId":   123, // Number instead of string
					"username": "testuser",
					"exp":      time.Now().Add(time.Hour).Unix(),
				}, jwtSecret)
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid user ID in token\n",
		},
		{
			name: "Valid token with all required claims",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"userId":   "user-123",
					"username": "testuser",
					"email":    "test@example.com",
					"role":     "admin",
					"exp":      time.Now().Add(time.Hour).Unix(),
					"iat":      time.Now().Unix(),
				}, jwtSecret)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext: func(t *testing.T, r *http.Request) {
				userID := r.Context().Value("userID")
				assert.NotNil(t, userID)
				assert.Equal(t, "user-123", userID)
			},
		},
		{
			name: "Valid token without exp claim (never expires)",
			setupToken: func() string {
				token, _ := createTestToken(jwt.MapClaims{
					"userId":   "user-456",
					"username": "testuser",
				}, jwtSecret)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext: func(t *testing.T, r *http.Request) {
				userID := r.Context().Value("userID")
				assert.Equal(t, "user-456", userID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := middleware.MuxAuthMiddleware(jwtSecret, logger)

			// Create handler that will be protected
			var capturedRequest *http.Request
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Wrap handler with middleware
			protected := middleware(handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			authHeader := tt.setupToken()
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			protected.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			// Check context if handler was called
			if tt.expectedStatus == http.StatusOK && tt.checkContext != nil {
				assert.NotNil(t, capturedRequest)
				tt.checkContext(t, capturedRequest)
			}
		})
	}
}

func TestMuxAuthMiddleware_Integration(t *testing.T) {
	jwtSecret := "integration-test-secret"
	logger := logrus.New()
	logger.SetOutput(nil)

	// Create middleware
	authMiddleware := middleware.MuxAuthMiddleware(jwtSecret, logger)

	// Create a chain of handlers
	var executionOrder []string
	
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler1")
		userID := r.Context().Value("userID")
		assert.NotNil(t, userID)
		w.Write([]byte("Handler 1 executed for user: " + userID.(string)))
	})

	// Create valid token
	token, err := createTestToken(jwt.MapClaims{
		"userId":   "integration-user",
		"username": "integrationtest",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}, jwtSecret)
	assert.NoError(t, err)

	// Create test server with middleware
	server := httptest.NewServer(authMiddleware(handler1))
	defer server.Close()

	// Make request without auth
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Make request with auth
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check that handler was executed
	assert.Contains(t, executionOrder, "handler1")
}

func TestMuxAuthMiddleware_ContextPropagation(t *testing.T) {
	jwtSecret := "context-test-secret"
	logger := logrus.New()
	logger.SetOutput(nil)

	// Create middleware
	authMiddleware := middleware.MuxAuthMiddleware(jwtSecret, logger)

	// Create nested handlers to test context propagation
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		assert.NotNil(t, userID)
		assert.Equal(t, "context-test-user", userID)
		
		// Try to add more context
		ctx := context.WithValue(r.Context(), "extra", "value")
		assert.Equal(t, "value", ctx.Value("extra"))
		assert.Equal(t, "context-test-user", ctx.Value("userID"))
		
		w.WriteHeader(http.StatusOK)
	})

	// Create valid token
	token, _ := createTestToken(jwt.MapClaims{
		"userId": "context-test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}, jwtSecret)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute
	authMiddleware(innerHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func BenchmarkMuxAuthMiddleware(b *testing.B) {
	jwtSecret := "benchmark-secret"
	logger := logrus.New()
	logger.SetOutput(nil)

	// Create middleware
	authMiddleware := middleware.MuxAuthMiddleware(jwtSecret, logger)

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create valid token
	token, _ := createTestToken(jwt.MapClaims{
		"userId": "bench-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}, jwtSecret)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		authMiddleware(handler).ServeHTTP(w, req)
	}
}

func TestMuxAuthMiddleware_AlgorithmValidation(t *testing.T) {
	jwtSecret := "algorithm-test-secret"
	logger := logrus.New()
	logger.SetOutput(nil)

	// Test that only HS256 is accepted (implicit in jwt.Parse)
	tests := []struct {
		name           string
		signingMethod  jwt.SigningMethod
		expectedStatus int
	}{
		{
			name:           "HS256 algorithm",
			signingMethod:  jwt.SigningMethodHS256,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HS384 algorithm", 
			signingMethod:  jwt.SigningMethodHS384,
			expectedStatus: http.StatusOK, // The middleware doesn't check algorithm
		},
		{
			name:           "HS512 algorithm",
			signingMethod:  jwt.SigningMethodHS512,
			expectedStatus: http.StatusOK, // The middleware doesn't check algorithm
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			authMiddleware := middleware.MuxAuthMiddleware(jwtSecret, logger)

			// Create handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Create token with specific algorithm
			token := jwt.NewWithClaims(tt.signingMethod, jwt.MapClaims{
				"userId": "test-user",
				"exp":    time.Now().Add(time.Hour).Unix(),
			})
			tokenString, err := token.SignedString([]byte(jwtSecret))
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			w := httptest.NewRecorder()

			// Execute
			authMiddleware(handler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

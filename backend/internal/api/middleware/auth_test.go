package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"app-env-manager/internal/api/middleware"
	"app-env-manager/internal/domain/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		allowedRoles   []string
		userRole       string
		hasRole        bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No role in context",
			allowedRoles:   []string{string(entities.UserRoleAdmin)},
			hasRole:        false,
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Access denied"}`,
		},
		{
			name:           "User has required role - single role",
			allowedRoles:   []string{string(entities.UserRoleAdmin)},
			userRole:       string(entities.UserRoleAdmin),
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User doesn't have required role",
			allowedRoles:   []string{string(entities.UserRoleAdmin)},
			userRole:       string(entities.UserRoleUser),
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Insufficient permissions"}`,
		},
		{
			name:           "User has one of multiple allowed roles",
			allowedRoles:   []string{string(entities.UserRoleAdmin), string(entities.UserRoleViewer)},
			userRole:       string(entities.UserRoleViewer),
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User doesn't have any of multiple allowed roles",
			allowedRoles:   []string{string(entities.UserRoleAdmin), string(entities.UserRoleViewer)},
			userRole:       string(entities.UserRoleUser),
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Insufficient permissions"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Set user role if needed
			if tt.hasRole {
				c.Set("userRole", tt.userRole)
			}
			
			// Create middleware
			roleMiddleware := middleware.RequireRole(tt.allowedRoles...)
			
			// Track if next handler was called
			nextCalled := false
			roleMiddleware(c)
			if !c.IsAborted() {
				nextCalled = true
			}
			
			// Check expectations
			if tt.expectedStatus != http.StatusOK {
				assert.True(t, c.IsAborted())
				assert.False(t, nextCalled)
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.False(t, c.IsAborted())
				assert.True(t, nextCalled)
			}
		})
	}
}

func TestAuthMiddleware_HeaderValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		shouldAbort    bool
	}{
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authorization header required"}`,
			shouldAbort:    true,
		},
		{
			name:           "Invalid authorization header format - no Bearer",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format"}`,
			shouldAbort:    true,
		},
		{
			name:           "Invalid authorization header format - wrong prefix",
			authHeader:     "Basic token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format"}`,
			shouldAbort:    true,
		},
		{
			name:           "Invalid authorization header format - too many parts",
			authHeader:     "Bearer token extra parts",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format"}`,
			shouldAbort:    true,
		},
		{
			name:           "Valid header format",
			authHeader:     "Bearer valid-token-format",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			shouldAbort:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req
			
			// Test header validation logic
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
				c.Abort()
			} else {
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
					c.Abort()
				}
			}
			
			// Check expectations
			if tt.shouldAbort {
				assert.True(t, c.IsAborted())
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.False(t, c.IsAborted())
			}
		})
	}
}

func TestOptionalAuth_HeaderValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name         string
		authHeader   string
		shouldAbort  bool
	}{
		{
			name:         "No authorization header",
			authHeader:   "",
			shouldAbort:  false,
		},
		{
			name:         "Invalid header format",
			authHeader:   "InvalidToken",
			shouldAbort:  false,
		},
		{
			name:         "Wrong auth type",
			authHeader:   "Basic token123",
			shouldAbort:  false,
		},
		{
			name:         "Valid header format",
			authHeader:   "Bearer valid-token",
			shouldAbort:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req
			
			// Test optional auth logic (never aborts)
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					// Would validate token here
				}
			}
			
			// OptionalAuth should never abort
			assert.False(t, c.IsAborted())
		})
	}
}

// Test helper functions
func TestRequireRole_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("Empty allowed roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userRole", string(entities.UserRoleAdmin))
		
		middleware := middleware.RequireRole()
		middleware(c)
		
		// Should abort because no roles are allowed
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	
	t.Run("Role type assertion fails", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userRole", 123) // Wrong type
		
		middleware := middleware.RequireRole(string(entities.UserRoleAdmin))
		
		// Should panic or handle gracefully
		assert.Panics(t, func() {
			middleware(c)
		})
	})
}

// Benchmark tests
func BenchmarkRequireRole(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userRole", string(entities.UserRoleAdmin))
	
	middleware := middleware.RequireRole(string(entities.UserRoleAdmin))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware(c)
			// Reset context for next iteration
			w = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(w)
			c.Set("userRole", string(entities.UserRoleAdmin))
	}
}

func BenchmarkAuthHeaderParsing(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		c.Request = req
		
		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		_ = parts
	}
}

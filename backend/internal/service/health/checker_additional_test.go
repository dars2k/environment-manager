package health_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/service/health"
	"github.com/stretchr/testify/assert"
)

func TestBuildRequest_VariousScenarios(t *testing.T) {
	checker := health.NewChecker(5 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name        string
		env         *entities.Environment
		expectedURL string
		expectError bool
	}{
		{
			name: "Full URL endpoint",
			env: &entities.Environment{
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "https://api.example.com/health",
					Method:   "GET",
				},
			},
			expectedURL: "https://api.example.com/health",
			expectError: false,
		},
		{
			name: "Environment URL with relative endpoint",
			env: &entities.Environment{
				EnvironmentURL: "https://app.example.com",
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "/api/health",
					Method:   "GET",
				},
			},
			expectedURL: "https://app.example.com/api/health",
			expectError: false,
		},
		{
			name: "Environment URL without endpoint",
			env: &entities.Environment{
				EnvironmentURL: "https://app.example.com/",
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "",
					Method:   "GET",
				},
			},
			expectedURL: "https://app.example.com",
			expectError: false,
		},
		{
			name: "Build from target - HTTPS port",
			env: &entities.Environment{
				Target: entities.Target{
					Host: "example.com",
					Port: 443,
				},
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "/health",
					Method:   "GET",
				},
			},
			expectedURL: "https://example.com:443/health",
			expectError: false,
		},
		{
			name: "Build from target with domain",
			env: &entities.Environment{
				Target: entities.Target{
					Host:   "192.168.1.1",
					Port:   8080,
					Domain: "app.local",
				},
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "/status",
					Method:   "POST",
				},
			},
			expectedURL: "http://app.local:8080/status",
			expectError: false,
		},
		{
			name: "With custom headers",
			env: &entities.Environment{
				Target: entities.Target{
					Host: "localhost",
					Port: 3000,
				},
				HealthCheck: entities.HealthCheckConfig{
					Endpoint: "/api/ping",
					Method:   "GET",
					Headers: map[string]string{
						"Authorization": "Bearer token",
						"X-API-Key":     "secret",
					},
				},
			},
			expectedURL: "http://localhost:3000/api/ping",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server to validate the request
			var capturedReq *http.Request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Override the endpoint to use test server
			if tt.env.HealthCheck.Endpoint != "" && !isFullURL(tt.env.HealthCheck.Endpoint) {
				tt.env.HealthCheck.Endpoint = server.URL + tt.env.HealthCheck.Endpoint
			} else if tt.env.EnvironmentURL != "" {
				tt.env.EnvironmentURL = server.URL
			} else {
				tt.env.Target.Host = "localhost"
				tt.env.Target.Port = getPort(server.URL)
			}

			result, err := checker.CheckHealth(ctx, tt.env)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				
				// Verify headers if set
				if capturedReq != nil && tt.env.HealthCheck.Headers != nil {
					for key, value := range tt.env.HealthCheck.Headers {
						assert.Equal(t, value, capturedReq.Header.Get(key))
					}
					// Check default User-Agent
					assert.NotEmpty(t, capturedReq.Header.Get("User-Agent"))
				}
			}
		})
	}
}

func TestIsFullURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"https://example.com:8080/path", true},
		{"ftp://example.com", true},
		{"ws://example.com", true},
		{"/api/health", false},
		{"api/health", false},
		{"localhost:8080", false},
		{"", false},
		{"http", false},
		{"https", false},
		{"http:", false},
		{"https:", false},
		{"http:/", false},
		{"https:/", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isFullURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	validator := health.NewValidator()

	tests := []struct {
		name       string
		config     entities.ValidationConfig
		statusCode int
		body       []byte
		expectOK   bool
		expectMsg  string
	}{
		{
			name: "Status code match - int value",
			config: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
			statusCode: 200,
			body:       []byte("OK"),
			expectOK:   true,
			expectMsg:  "Status code 200 matches expected 200",
		},
		{
			name: "Status code match - float64 value",
			config: entities.ValidationConfig{
				Type:  "statusCode",
				Value: float64(201),
			},
			statusCode: 201,
			body:       []byte("Created"),
			expectOK:   true,
			expectMsg:  "Status code 201 matches expected 201",
		},
		{
			name: "Status code mismatch",
			config: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
			statusCode: 404,
			body:       []byte("Not Found"),
			expectOK:   false,
			expectMsg:  "Status code 404 does not match expected 200",
		},
		{
			name: "Invalid status code configuration",
			config: entities.ValidationConfig{
				Type:  "statusCode",
				Value: "200", // string instead of number
			},
			statusCode: 200,
			body:       []byte("OK"),
			expectOK:   false,
			expectMsg:  "Invalid status code configuration: 200",
		},
		{
			name: "JSON regex match",
			config: entities.ValidationConfig{
				Type:  "jsonRegex",
				Value: `"status":\s*"healthy"`,
			},
			statusCode: 200,
			body:       []byte(`{"status": "healthy", "timestamp": "2024-01-01"}`),
			expectOK:   true,
			expectMsg:  "Response matches expected pattern",
		},
		{
			name: "JSON regex no match",
			config: entities.ValidationConfig{
				Type:  "jsonRegex",
				Value: `"status":\s*"healthy"`,
			},
			statusCode: 200,
			body:       []byte(`{"status": "unhealthy", "error": "database down"}`),
			expectOK:   false,
			expectMsg:  "Response does not match expected pattern",
		},
		{
			name: "Invalid regex pattern",
			config: entities.ValidationConfig{
				Type:  "jsonRegex",
				Value: "[invalid(regex",
			},
			statusCode: 200,
			body:       []byte("test"),
			expectOK:   false,
			expectMsg:  "Invalid regex pattern:",
		},
		{
			name: "Invalid regex configuration",
			config: entities.ValidationConfig{
				Type:  "jsonRegex",
				Value: 123, // not a string
			},
			statusCode: 200,
			body:       []byte("test"),
			expectOK:   false,
			expectMsg:  "Invalid regex pattern configuration",
		},
		{
			name: "Unknown validation type",
			config: entities.ValidationConfig{
				Type:  "unknown",
				Value: "test",
			},
			statusCode: 200,
			body:       []byte("test"),
			expectOK:   false,
			expectMsg:  "Unknown validation type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := validator.Validate(tt.config, tt.statusCode, tt.body)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectMsg != "" {
				assert.Contains(t, msg, tt.expectMsg)
			}
		})
	}
}

func TestCheckHealth_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		env            *entities.Environment
		expectedStatus entities.HealthStatus
		expectedMsg    string
	}{
		{
			name: "Large response body",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Write more than 1MB
					largeData := make([]byte, 2*1024*1024)
					w.Write(largeData)
				}))
			},
			env: &entities.Environment{
				HealthCheck: entities.HealthCheckConfig{
					Enabled:  true,
					Method:   "GET",
					Validation: entities.ValidationConfig{
						Type:  "statusCode",
						Value: 200,
					},
				},
			},
			expectedStatus: entities.HealthStatusHealthy,
		},
		{
			name: "Redirect handling",
			setupServer: func() *httptest.Server {
				redirectCount := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					redirectCount++
					if redirectCount <= 3 {
						http.Redirect(w, r, fmt.Sprintf("/redirect%d", redirectCount), http.StatusFound)
						return
					}
					w.WriteHeader(http.StatusOK)
				}))
			},
			env: &entities.Environment{
				HealthCheck: entities.HealthCheckConfig{
					Enabled:  true,
					Method:   "GET",
					Validation: entities.ValidationConfig{
						Type:  "statusCode",
						Value: 200,
					},
				},
			},
			expectedStatus: entities.HealthStatusHealthy,
		},
		{
			name: "Too many redirects",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Always redirect
					http.Redirect(w, r, "/infinite", http.StatusFound)
				}))
			},
			env: &entities.Environment{
				HealthCheck: entities.HealthCheckConfig{
					Enabled: true,
					Method:  "GET",
				},
			},
			expectedStatus: entities.HealthStatusUnhealthy,
			expectedMsg:    "too many redirects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			checker := health.NewChecker(5 * time.Second)
			tt.env.HealthCheck.Endpoint = server.URL

			ctx := context.Background()
			result, err := checker.CheckHealth(ctx, tt.env)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			if tt.expectedMsg != "" {
				assert.Contains(t, result.Message, tt.expectedMsg)
			}
		})
	}
}

// Helper functions
func isFullURL(s string) bool {
	if len(s) >= 7 && s[:7] == "http://" {
		return true
	}
	if len(s) >= 8 && s[:8] == "https://" {
		return true
	}
	if strings.Contains(s, "://") {
		return true
	}
	return false
}

func getPort(url string) int {
	// Simple port extraction from URL
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) >= 3 {
			portStr := strings.Split(parts[2], "/")[0]
			var port int
			fmt.Sscanf(portStr, "%d", &port)
			return port
		}
	}
	return 80
}

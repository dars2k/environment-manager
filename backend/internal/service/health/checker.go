package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"app-env-manager/internal/domain/entities"
)

// Checker performs health checks on environments
type Checker struct {
	httpClient *http.Client
	validator  *Validator
}

// CheckResult contains the result of a health check
type CheckResult struct {
	Status       entities.HealthStatus
	Message      string
	ResponseTime int64 // milliseconds
	StatusCode   int
}

// NewChecker creates a new health checker
func NewChecker(timeout time.Duration) *Checker {
	return &Checker{
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		validator: NewValidator(),
	}
}

// CheckHealth performs a health check on an environment
func (c *Checker) CheckHealth(ctx context.Context, env *entities.Environment) (*CheckResult, error) {
	if !env.HealthCheck.Enabled {
		return &CheckResult{
			Status:  entities.HealthStatusUnknown,
			Message: "Health check disabled",
		}, nil
	}

	start := time.Now()
	
	// Build request
	req, err := c.buildRequest(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	responseTime := time.Since(start).Milliseconds()
	
	if err != nil {
		return &CheckResult{
			Status:       entities.HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Request failed: %v", err),
			ResponseTime: responseTime,
		}, nil
	}
	defer resp.Body.Close()

	// Read response body for validation
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
	if err != nil {
		return &CheckResult{
			Status:       entities.HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Failed to read response: %v", err),
			ResponseTime: responseTime,
			StatusCode:   resp.StatusCode,
		}, nil
	}

	// Validate response
	valid, message := c.validator.Validate(env.HealthCheck.Validation, resp.StatusCode, body)
	
	status := entities.HealthStatusHealthy
	if !valid {
		status = entities.HealthStatusUnhealthy
	}

	return &CheckResult{
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
	}, nil
}

// buildRequest builds an HTTP request for health check
func (c *Checker) buildRequest(ctx context.Context, env *entities.Environment) (*http.Request, error) {
	// Build URL
	var url string
	
	// Check if the endpoint already contains a full URL
	if isFullURL(env.HealthCheck.Endpoint) {
		url = env.HealthCheck.Endpoint
	} else if env.EnvironmentURL != "" && isFullURL(env.EnvironmentURL) {
		// If environmentURL is set and is a full URL, use it as base
		baseURL := strings.TrimRight(env.EnvironmentURL, "/")
		endpoint := strings.TrimLeft(env.HealthCheck.Endpoint, "/")
		if endpoint != "" {
			url = fmt.Sprintf("%s/%s", baseURL, endpoint)
		} else {
			url = baseURL
		}
	} else {
		// Build URL from components
		scheme := "http"
		if env.Target.Port == 443 {
			scheme = "https"
		}
		
		host := env.Target.Host
		if env.Target.Domain != "" {
			host = env.Target.Domain
		}
		
		url = fmt.Sprintf("%s://%s:%d%s", scheme, host, env.Target.Port, env.HealthCheck.Endpoint)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, env.HealthCheck.Method, url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers
	for key, value := range env.HealthCheck.Headers {
		req.Header.Set(key, value)
	}
	
	// Set default User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "AppEnvManager/1.0")
	}
	
	return req, nil
}

// isFullURL checks if the string is a full URL (starts with http:// or https://)
func isFullURL(s string) bool {
	// Check if it starts with http:// or https://
	if len(s) >= 7 && s[:7] == "http://" {
		return true
	}
	if len(s) >= 8 && s[:8] == "https://" {
		return true
	}
	// Also check for URLs that might have been entered without the protocol
	// but contain common URL patterns
	if strings.Contains(s, "://") {
		return true
	}
	return false
}

// Validator validates health check responses
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a health check response
func (v *Validator) Validate(config entities.ValidationConfig, statusCode int, body []byte) (bool, string) {
	switch config.Type {
	case "statusCode":
		expectedCode, ok := config.Value.(float64)
		if !ok {
			if intCode, ok := config.Value.(int); ok {
				expectedCode = float64(intCode)
			} else {
				return false, fmt.Sprintf("Invalid status code configuration: %v", config.Value)
			}
		}
		
		if statusCode == int(expectedCode) {
			return true, fmt.Sprintf("Status code %d matches expected %d", statusCode, int(expectedCode))
		}
		return false, fmt.Sprintf("Status code %d does not match expected %d", statusCode, int(expectedCode))
		
	case "jsonRegex":
		pattern, ok := config.Value.(string)
		if !ok {
			return false, "Invalid regex pattern configuration"
		}
		
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Sprintf("Invalid regex pattern: %v", err)
		}
		
		if regex.Match(body) {
			return true, "Response matches expected pattern"
		}
		return false, "Response does not match expected pattern"
		
	default:
		return false, fmt.Sprintf("Unknown validation type: %s", config.Type)
	}
}

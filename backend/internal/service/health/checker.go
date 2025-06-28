package health

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"app-env-manager/internal/domain/entities"
)

// Checker performs health checks on environments
type Checker struct {
	httpClient *http.Client
	validator  *Validator
	allowLocalhost bool  // Allow localhost for testing
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
	checker := &Checker{
		validator: NewValidator(),
		allowLocalhost: false,
	}
	
	checker.httpClient = &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Validate each redirect URL
			if err := checker.validateURLWithOptions(req.URL.String()); err != nil {
				return fmt.Errorf("redirect to invalid URL: %w", err)
			}
			// Limit redirect chain
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	
	return checker
}

// NewCheckerForTesting creates a new health checker that allows localhost addresses
func NewCheckerForTesting(timeout time.Duration) *Checker {
	checker := NewChecker(timeout)
	checker.allowLocalhost = true
	return checker
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
	var checkURL string
	
	// Check if the endpoint already contains a full URL
	if isFullURL(env.HealthCheck.Endpoint) {
		checkURL = env.HealthCheck.Endpoint
	} else if env.EnvironmentURL != "" && isFullURL(env.EnvironmentURL) {
		// If environmentURL is set and is a full URL, use it as base
		baseURL := strings.TrimRight(env.EnvironmentURL, "/")
		endpoint := strings.TrimLeft(env.HealthCheck.Endpoint, "/")
		if endpoint != "" {
			checkURL = fmt.Sprintf("%s/%s", baseURL, endpoint)
		} else {
			checkURL = baseURL
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
		
		checkURL = fmt.Sprintf("%s://%s:%d%s", scheme, host, env.Target.Port, env.HealthCheck.Endpoint)
	}
	
	// Validate URL to prevent SSRF attacks
	if err := c.validateURLWithOptions(checkURL); err != nil {
		return nil, fmt.Errorf("invalid health check URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, env.HealthCheck.Method, checkURL, nil)
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
		
		// Validate regex pattern to prevent ReDoS attacks
		if err := validateRegexPattern(pattern); err != nil {
			return false, fmt.Sprintf("Invalid regex pattern: %v", err)
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

// validateURLWithOptions validates a URL with checker-specific options
func (c *Checker) validateURLWithOptions(rawURL string) error {
	return validateURLWithLocalhost(rawURL, c.allowLocalhost)
}

// validateURL validates a URL to prevent SSRF attacks
func validateURL(rawURL string) error {
	return validateURLWithLocalhost(rawURL, false)
}

// validateURLWithLocalhost validates a URL to prevent SSRF attacks with optional localhost allowance
func validateURLWithLocalhost(rawURL string, allowLocalhost bool) error {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme - only allow HTTP and HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS schemes are allowed")
	}

	// Extract hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	// Check for local/private addresses
	// Resolve the hostname to IP addresses
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If we can't resolve, it might be a non-existent host
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		// Check for localhost
		if ip.IsLoopback() && !allowLocalhost {
			return fmt.Errorf("localhost addresses are not allowed")
		}

		// Check for private networks (allow if localhost is allowed and it's a loopback)
		if ip.IsPrivate() && !(allowLocalhost && ip.IsLoopback()) {
			return fmt.Errorf("private network addresses are not allowed")
		}

		// Check for link-local addresses
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local addresses are not allowed")
		}

		// Check for multicast
		if ip.IsMulticast() {
			return fmt.Errorf("multicast addresses are not allowed")
		}

		// Check for unspecified addresses (0.0.0.0 or ::)
		if ip.IsUnspecified() {
			return fmt.Errorf("unspecified addresses are not allowed")
		}
	}

	// Additional checks for common internal hostnames
	lowerHostname := strings.ToLower(hostname)
	
	// Always block metadata endpoints
	metadataHostnames := []string{
		"metadata", "metadata.google.internal", // GCP metadata
		"169.254.169.254", // AWS/Azure metadata
	}
	
	for _, blocked := range metadataHostnames {
		if lowerHostname == blocked {
			return fmt.Errorf("hostname '%s' is not allowed", hostname)
		}
	}
	
	// Block localhost variants unless explicitly allowed
	if !allowLocalhost {
		localhostVariants := []string{
			"localhost", "127.0.0.1", "0.0.0.0", "::1",
		}
		
		for _, blocked := range localhostVariants {
			if lowerHostname == blocked {
				return fmt.Errorf("hostname '%s' is not allowed", hostname)
			}
		}
	}

	return nil
}

// validateRegexPattern validates a regex pattern to prevent ReDoS attacks
func validateRegexPattern(pattern string) error {
	// Check pattern length
	if len(pattern) > 1000 {
		return fmt.Errorf("regex pattern too long (max 1000 characters)")
	}
	
	// Check for dangerous regex patterns that could cause ReDoS
	dangerousPatterns := []string{
		`(\w+)*$`,           // Nested quantifiers
		`(a+)+`,             // Nested quantifiers
		`(.*)*`,             // Nested quantifiers
		`(.+)+`,             // Nested quantifiers
		`(\d+)*\d`,          // Overlapping patterns
		`(a|a)*`,            // Alternation with overlap
		`(a|ab)*`,           // Alternation with overlap
		`([^x]|x)*`,         // Alternation that matches everything
	}
	
	// Check for common ReDoS patterns
	for _, dangerous := range dangerousPatterns {
		if strings.Contains(pattern, dangerous) {
			return fmt.Errorf("potentially dangerous regex pattern detected")
		}
	}
	
	// Check for excessive backtracking indicators
	// Count nested quantifiers
	nestedCount := 0
	inGroup := false
	hasQuantifier := false
	
	for i, char := range pattern {
		switch char {
		case '(':
			if hasQuantifier {
				nestedCount++
			}
			inGroup = true
			hasQuantifier = false
		case ')':
			inGroup = false
			// Check if followed by quantifier
			if i+1 < len(pattern) {
				next := pattern[i+1]
				if next == '*' || next == '+' || next == '?' || next == '{' {
					hasQuantifier = true
				}
			}
		case '*', '+', '?':
			if inGroup {
				hasQuantifier = true
			}
		case '{':
			if inGroup {
				hasQuantifier = true
			}
		}
	}
	
	if nestedCount > 2 {
		return fmt.Errorf("too many nested quantifiers in regex pattern")
	}
	
	// Try to compile the regex with a timeout to catch problematic patterns
	done := make(chan error, 1)
	go func() {
		_, err := regexp.Compile(pattern)
		done <- err
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("regex compilation timeout - pattern may be too complex")
	}
	
	return nil
}

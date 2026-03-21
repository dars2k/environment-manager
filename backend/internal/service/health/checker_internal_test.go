package health

import (
	"context"
	"testing"

	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newTestChecker() *Checker {
	return NewChecker(0)
}

func baseEnv() *entities.Environment {
	return &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "test",
		Target: entities.Target{
			Host: "myhost.example.com",
			Port: 8080,
		},
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Method:   "GET",
			Endpoint: "/health",
		},
	}
}

// ---- isFullURL ----

func TestIsFullURL_HTTP(t *testing.T) {
	assert.True(t, isFullURL("http://example.com/health"))
}

func TestIsFullURL_HTTPS(t *testing.T) {
	assert.True(t, isFullURL("https://example.com/health"))
}

func TestIsFullURL_CustomScheme(t *testing.T) {
	assert.True(t, isFullURL("ftp://example.com/file"))
}

func TestIsFullURL_RelativePath(t *testing.T) {
	assert.False(t, isFullURL("/health"))
}

func TestIsFullURL_PlainPath(t *testing.T) {
	assert.False(t, isFullURL("health"))
}

func TestIsFullURL_Empty(t *testing.T) {
	assert.False(t, isFullURL(""))
}

// ---- buildRequest ----

func TestBuildRequest_EndpointIsFullURL(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.HealthCheck.Endpoint = "http://myhost.example.com:8080/health"

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, "http://myhost.example.com:8080/health", req.URL.String())
}

func TestBuildRequest_EnvironmentURLBase(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.EnvironmentURL = "http://staging.example.com"
	env.HealthCheck.Endpoint = "/ping"

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Contains(t, req.URL.String(), "staging.example.com")
	assert.Contains(t, req.URL.String(), "ping")
}

func TestBuildRequest_EnvironmentURLNoEndpoint(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.EnvironmentURL = "http://staging.example.com"
	env.HealthCheck.Endpoint = ""

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, "http://staging.example.com", req.URL.String())
}

func TestBuildRequest_BuiltFromComponents(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	// No EnvironmentURL, no full URL endpoint → build from host:port

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Contains(t, req.URL.String(), "myhost.example.com")
	assert.Contains(t, req.URL.String(), "8080")
}

func TestBuildRequest_Port443UsesHTTPS(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.Target.Port = 443

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.True(t, req.URL.Scheme == "https", "expected https scheme, got %s", req.URL.Scheme)
}

func TestBuildRequest_DomainOverridesHost(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.Target.Domain = "override.example.com"

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Contains(t, req.URL.String(), "override.example.com")
}

func TestBuildRequest_DefaultUserAgent(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, "AppEnvManager/1.0", req.Header.Get("User-Agent"))
}

func TestBuildRequest_CustomHeaderPreserved(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	env.HealthCheck.Headers = map[string]string{
		"X-Custom": "value",
	}

	req, err := c.buildRequest(context.Background(), env)
	require.NoError(t, err)
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
}

func TestBuildRequest_InvalidURLReturnsError(t *testing.T) {
	c := newTestChecker()
	env := baseEnv()
	// Passing a method that will cause http.NewRequest to fail
	env.HealthCheck.Method = "INVALID METHOD WITH SPACES"
	env.HealthCheck.Endpoint = "http://example.com/health"

	_, err := c.buildRequest(context.Background(), env)
	assert.Error(t, err)
}

package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/service/health"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewChecker(t *testing.T) {
	checker := health.NewChecker(5 * time.Second)
	assert.NotNil(t, checker)
}

func TestChecker_CheckHealth_HTTPSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		
		// Check for custom headers
		assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	checker := health.NewCheckerForTesting(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: server.URL + "/health",
			Method:   "GET",
			Timeout:  5,
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
			},
			Validation: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusHealthy, result.Status)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.GreaterOrEqual(t, result.ResponseTime, int64(0))
	assert.NotEmpty(t, result.Message)
}

func TestChecker_CheckHealth_HTTPFailure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	checker := health.NewCheckerForTesting(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: server.URL + "/health",
			Method:   "GET",
			Timeout:  5,
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusUnhealthy, result.Status)
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.NotEmpty(t, result.Message)
}

func TestChecker_CheckHealth_Timeout(t *testing.T) {
	// Create test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := health.NewCheckerForTesting(1 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: server.URL + "/health",
			Method:   "GET",
			Timeout:  1, // 1 second timeout
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "Request failed")
}

func TestChecker_CheckHealth_InvalidURL(t *testing.T) {
	checker := health.NewChecker(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: "://invalid-url",
			Method:   "GET",
			Timeout:  5,
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestChecker_CheckHealth_DisabledCheck(t *testing.T) {
	checker := health.NewChecker(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled: false,
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusUnknown, result.Status)
	assert.Equal(t, "Health check disabled", result.Message)
}

func TestChecker_CheckHealth_TCPSuccess(t *testing.T) {
	// Create a TCP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := health.NewCheckerForTesting(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: server.URL, // Use full URL
			Timeout:  5,
			Validation: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusHealthy, result.Status)
	assert.NotEmpty(t, result.Message)
}

func TestChecker_CheckHealth_HTTPSuccessDefaultStatusCode(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	checker := health.NewCheckerForTesting(5 * time.Second)
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: server.URL + "/health",
			Method:   "GET",
			Timeout:  5,
			Validation: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
		},
	}

	ctx := context.Background()
	result, err := checker.CheckHealth(ctx, env)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.HealthStatusHealthy, result.Status)
	assert.Equal(t, http.StatusOK, result.StatusCode)
}

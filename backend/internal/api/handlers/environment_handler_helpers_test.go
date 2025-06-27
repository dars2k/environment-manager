package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test helper functions that are easier to test in isolation

func TestRespondJSON(t *testing.T) {
	// Create a test handler
	handler := &EnvironmentHandler{}

	// Test data
	testData := map[string]string{
		"message": "test message",
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Call respondJSON
	handler.respondJSON(w, http.StatusOK, testData)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response dto.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.NotEmpty(t, response.Metadata.Timestamp)
	assert.Equal(t, "1.0.0", response.Metadata.Version)
}

func TestRespondError(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Validation error",
			err:            errors.NewValidationError("field", "invalid value"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name:           "Not found error",
			err:            errors.ErrEnvironmentNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "ENV_NOT_FOUND",
		},
		{
			name:           "Duplicate error",
			err:            errors.ErrEnvironmentAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "ENV_DUPLICATE",
		},
		{
			name:           "Unauthorized error",
			err:            errors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "AUTH_UNAUTHORIZED",
		},
		{
			name:           "Invalid credentials error",
			err:            errors.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "AUTH_INVALID",
		},
		{
			name:           "Generic error",
			err:            errors.NewInternalError(errors.NewInternalError(errors.ErrOperationFailed)),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := &EnvironmentHandler{}
			w := httptest.NewRecorder()

			handler.respondError(w, tc.err)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.Equal(t, tc.expectedCode, response.Error.Code)
			assert.NotEmpty(t, response.Error.Message)
			assert.NotEmpty(t, response.Metadata.Timestamp)
			assert.Equal(t, "1.0.0", response.Metadata.Version)
		})
	}
}

func TestParseListFilter(t *testing.T) {
	testCases := []struct {
		name           string
		queryString    string
		expectedPage   int
		expectedLimit  int
		expectedStatus *entities.HealthStatus
	}{
		{
			name:           "No parameters",
			queryString:    "",
			expectedPage:   1,
			expectedLimit:  200,
			expectedStatus: nil,
		},
		{
			name:           "Valid pagination",
			queryString:    "page=2&limit=50",
			expectedPage:   2,
			expectedLimit:  50,
			expectedStatus: nil,
		},
		{
			name:           "Invalid page",
			queryString:    "page=invalid",
			expectedPage:   1,
			expectedLimit:  200,
			expectedStatus: nil,
		},
		{
			name:           "Negative page",
			queryString:    "page=-5",
			expectedPage:   1,
			expectedLimit:  200,
			expectedStatus: nil,
		},
		{
			name:           "Zero page",
			queryString:    "page=0",
			expectedPage:   1,
			expectedLimit:  200,
			expectedStatus: nil,
		},
		{
			name:           "Limit exceeds max",
			queryString:    "limit=500",
			expectedPage:   1,
			expectedLimit:  200,
			expectedStatus: nil,
		},
		{
			name:           "Valid limit within bounds",
			queryString:    "limit=75",
			expectedPage:   1,
			expectedLimit:  75,
			expectedStatus: nil,
		},
		{
			name:          "All parameters",
			queryString:   "page=3&limit=25&status=healthy",
			expectedPage:  3,
			expectedLimit: 25,
			expectedStatus: func() *entities.HealthStatus { s := entities.HealthStatusHealthy; return &s }(),
		},
		{
			name:          "Status filter",
			queryString:   "status=unhealthy",
			expectedPage:  1,
			expectedLimit: 200,
			expectedStatus: func() *entities.HealthStatus { s := entities.HealthStatusUnhealthy; return &s }(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/environments?"+tc.queryString, nil)
			filter := parseListFilter(req)

			assert.Equal(t, tc.expectedPage, filter.Pagination.Page)
			assert.Equal(t, tc.expectedLimit, filter.Pagination.Limit)
			
			if tc.expectedStatus == nil {
				assert.Nil(t, filter.Status)
			} else {
				assert.NotNil(t, filter.Status)
				assert.Equal(t, *tc.expectedStatus, *filter.Status)
			}
		})
	}
}

func TestGenerateOperationID(t *testing.T) {
	id1 := generateOperationID()
	
	// IDs should have the correct format
	assert.Contains(t, id1, "op-")
	assert.Regexp(t, `^op-\d{14}$`, id1)

	// Wait 1 second to ensure different timestamp
	time.Sleep(1 * time.Second)
	id2 := generateOperationID()

	// IDs should be unique when generated at different times
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id2, "op-")
	assert.Regexp(t, `^op-\d{14}$`, id2)
}

func TestCurrentTimestamp(t *testing.T) {
	ts := currentTimestamp()

	// Should be parseable as RFC3339
	parsedTime, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err)

	// Should be recent (within last second)
	assert.WithinDuration(t, time.Now().UTC(), parsedTime, time.Second)
}

func TestRedactSensitiveFields(t *testing.T) {
	// Create an environment with sensitive data
	env := &entities.Environment{
		ID:             primitive.NewObjectID(),
		Name:           "test-env",
		Description:    "Test Environment",
		EnvironmentURL: "https://test.example.com",
		Credentials: entities.CredentialRef{
			Type:     "key",
			Username: "testuser",
			KeyID:    primitive.NewObjectID(), // This should be redacted
		},
		HealthCheck: entities.HealthCheckConfig{
			Headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"X-API-Key":     "secret-api-key",
				"X-Auth-Token":  "secret-auth-token",
				"Content-Type":  "application/json",
			},
		},
		Commands: entities.CommandConfig{
			Restart: entities.RestartConfig{
				Headers: map[string]string{
					"Authorization": "Basic secret",
					"X-API-Key":     "secret-key",
					"Accept":        "application/json",
				},
			},
		},
		UpgradeConfig: entities.UpgradeConfig{
			UpgradeCommand: entities.CommandDetails{
				Headers: map[string]string{
					"Authorization": "Bearer upgrade-token",
					"User-Agent":    "app-env-manager",
				},
			},
		},
		Metadata: map[string]interface{}{
			"password":    "secret-password",
			"privateKey":  "secret-private-key",
			"key":         "secret-key",
			"secret":      "secret-value",
			"environment": "production",
			"region":      "us-east-1",
		},
	}

	// Redact sensitive fields
	redacted := redactSensitiveFields(env)

	// Verify original is not modified
	assert.NotEqual(t, primitive.NilObjectID, env.Credentials.KeyID)

	// Verify redacted fields
	assert.Equal(t, primitive.NilObjectID, redacted.Credentials.KeyID)

	// Check health check headers
	assert.Equal(t, "[REDACTED]", redacted.HealthCheck.Headers["Authorization"])
	assert.Equal(t, "[REDACTED]", redacted.HealthCheck.Headers["X-API-Key"])
	assert.Equal(t, "[REDACTED]", redacted.HealthCheck.Headers["X-Auth-Token"])
	assert.Equal(t, "application/json", redacted.HealthCheck.Headers["Content-Type"])

	// Check command headers
	assert.Equal(t, "[REDACTED]", redacted.Commands.Restart.Headers["Authorization"])
	assert.Equal(t, "[REDACTED]", redacted.Commands.Restart.Headers["X-API-Key"])
	assert.Equal(t, "application/json", redacted.Commands.Restart.Headers["Accept"])

	// Check upgrade headers
	assert.Equal(t, "[REDACTED]", redacted.UpgradeConfig.UpgradeCommand.Headers["Authorization"])
	assert.Equal(t, "app-env-manager", redacted.UpgradeConfig.UpgradeCommand.Headers["User-Agent"])

	// Check metadata - sensitive fields should be removed
	_, hasPassword := redacted.Metadata["password"]
	_, hasPrivateKey := redacted.Metadata["privateKey"]
	_, hasKey := redacted.Metadata["key"]
	_, hasSecret := redacted.Metadata["secret"]
	
	assert.False(t, hasPassword)
	assert.False(t, hasPrivateKey)
	assert.False(t, hasKey)
	assert.False(t, hasSecret)
	assert.Equal(t, "production", redacted.Metadata["environment"])
	assert.Equal(t, "us-east-1", redacted.Metadata["region"])
}

func TestRedactSensitiveFields_NilInput(t *testing.T) {
	var env *entities.Environment
	redacted := redactSensitiveFields(env)
	assert.Nil(t, redacted)
}

func TestRedactSensitiveFields_EmptyHeaders(t *testing.T) {
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "test-env",
		// No headers set
	}

	redacted := redactSensitiveFields(env)
	assert.NotNil(t, redacted)
	assert.Equal(t, env.Name, redacted.Name)
}

func TestRedactSensitiveFieldsList(t *testing.T) {
	envs := []*entities.Environment{
		{
			ID:   primitive.NewObjectID(),
			Name: "env1",
			Credentials: entities.CredentialRef{
				KeyID: primitive.NewObjectID(),
			},
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "env2",
			Credentials: entities.CredentialRef{
				KeyID: primitive.NewObjectID(),
			},
		},
	}

	redacted := redactSensitiveFieldsList(envs)

	assert.Len(t, redacted, 2)
	for i, env := range redacted {
		assert.Equal(t, envs[i].Name, env.Name)
		assert.Equal(t, primitive.NilObjectID, env.Credentials.KeyID)
	}
}

func TestRedactSensitiveFieldsList_Empty(t *testing.T) {
	var envs []*entities.Environment
	redacted := redactSensitiveFieldsList(envs)
	assert.NotNil(t, redacted)
	assert.Empty(t, redacted)
}

// Test request/response types
func TestRestartRequest(t *testing.T) {
	req := RestartRequest{
		Force: true,
	}
	assert.True(t, req.Force)
}

func TestShutdownRequest(t *testing.T) {
	req := ShutdownRequest{
		GracefulTimeout: 30,
	}
	assert.Equal(t, 30, req.GracefulTimeout)
}

func TestOperationResponse(t *testing.T) {
	resp := OperationResponse{
		OperationID: "op-123",
		Status:      "in_progress",
	}
	assert.Equal(t, "op-123", resp.OperationID)
	assert.Equal(t, "in_progress", resp.Status)
}

func TestMessageResponse(t *testing.T) {
	resp := MessageResponse{
		Message: "Success",
	}
	assert.Equal(t, "Success", resp.Message)
}

func TestSuccessResponse(t *testing.T) {
	resp := SuccessResponse{
		Success: true,
		Data:    map[string]string{"key": "value"},
		Metadata: ResponseMetadata{
			Timestamp: "2024-01-01T00:00:00Z",
			Version:   "1.0.0",
		},
	}
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, "1.0.0", resp.Metadata.Version)
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    "TEST_ERROR",
			Message: "Test error",
			Details: map[string]interface{}{
				"field": "test",
			},
		},
		Metadata: ResponseMetadata{
			Timestamp: "2024-01-01T00:00:00Z",
			Version:   "1.0.0",
		},
	}
	assert.False(t, resp.Success)
	assert.Equal(t, "TEST_ERROR", resp.Error.Code)
	assert.NotNil(t, resp.Error.Details)
}

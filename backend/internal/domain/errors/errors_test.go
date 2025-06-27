package errors_test

import (
	"testing"

	"app-env-manager/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestDomainError_Error(t *testing.T) {
	err := errors.DomainError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Details: map[string]interface{}{
			"field": "test",
		},
	}

	assert.Equal(t, "Test error message", err.Error())
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  errors.DomainError
		code string
		msg  string
	}{
		{
			name: "ErrEnvironmentNotFound",
			err:  errors.ErrEnvironmentNotFound,
			code: "ENV_NOT_FOUND",
			msg:  "Environment not found",
		},
		{
			name: "ErrEnvironmentAlreadyExists",
			err:  errors.ErrEnvironmentAlreadyExists,
			code: "ENV_DUPLICATE",
			msg:  "Environment with this name already exists",
		},
		{
			name: "ErrSSHConnectionFailed",
			err:  errors.ErrSSHConnectionFailed,
			code: "SSH_CONNECTION_FAILED",
			msg:  "Failed to establish SSH connection",
		},
		{
			name: "ErrHealthCheckFailed",
			err:  errors.ErrHealthCheckFailed,
			code: "HEALTH_CHECK_FAILED",
			msg:  "Health check failed",
		},
		{
			name: "ErrOperationFailed",
			err:  errors.ErrOperationFailed,
			code: "OPERATION_FAILED",
			msg:  "Operation execution failed",
		},
		{
			name: "ErrInvalidCredentials",
			err:  errors.ErrInvalidCredentials,
			code: "AUTH_INVALID",
			msg:  "Invalid authentication credentials",
		},
		{
			name: "ErrUnauthorized",
			err:  errors.ErrUnauthorized,
			code: "AUTH_UNAUTHORIZED",
			msg:  "Unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.msg, tt.err.Message)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := errors.NewValidationError("username", "required field")
	
	domainErr, ok := err.(errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", domainErr.Code)
	assert.Equal(t, "Validation failed", domainErr.Message)
	assert.Equal(t, "username", domainErr.Details["field"])
	assert.Equal(t, "required field", domainErr.Details["reason"])
}

func TestNewInternalError(t *testing.T) {
	originalErr := assert.AnError
	err := errors.NewInternalError(originalErr)
	
	domainErr, ok := err.(errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", domainErr.Code)
	assert.Equal(t, "Internal server error", domainErr.Message)
	assert.Equal(t, originalErr.Error(), domainErr.Details["error"])
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Environment not found error",
			err:      errors.ErrEnvironmentNotFound,
			expected: true,
		},
		{
			name: "Other domain error",
			err: errors.DomainError{
				Code:    "OTHER_ERROR",
				Message: "Other error",
			},
			expected: false,
		},
		{
			name:     "Non-domain error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.IsNotFound(tt.err))
		})
	}
}

func TestIsDuplicate(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Environment duplicate error",
			err:      errors.ErrEnvironmentAlreadyExists,
			expected: true,
		},
		{
			name: "Other domain error",
			err: errors.DomainError{
				Code:    "OTHER_ERROR",
				Message: "Other error",
			},
			expected: false,
		},
		{
			name:     "Non-domain error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.IsDuplicate(tt.err))
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		context  string
		expected string
	}{
		{
			name:     "Wrap error with context",
			err:      assert.AnError,
			context:  "failed to connect",
			expected: "failed to connect: assert.AnError general error for testing",
		},
		{
			name:     "Nil error",
			err:      nil,
			context:  "some context",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := errors.WrapError(tt.err, tt.context)
			if tt.err == nil {
				assert.Nil(t, wrapped)
			} else {
				assert.NotNil(t, wrapped)
				assert.Equal(t, tt.expected, wrapped.Error())
			}
		})
	}
}

package errors

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e DomainError) Error() string {
	return e.Message
}

// Common domain errors
var (
	ErrEnvironmentNotFound = DomainError{
		Code:    "ENV_NOT_FOUND",
		Message: "Environment not found",
	}

	ErrEnvironmentAlreadyExists = DomainError{
		Code:    "ENV_DUPLICATE",
		Message: "Environment with this name already exists",
	}

	ErrSSHConnectionFailed = DomainError{
		Code:    "SSH_CONNECTION_FAILED",
		Message: "Failed to establish SSH connection",
	}

	ErrHealthCheckFailed = DomainError{
		Code:    "HEALTH_CHECK_FAILED",
		Message: "Health check failed",
	}

	ErrOperationFailed = DomainError{
		Code:    "OPERATION_FAILED",
		Message: "Operation execution failed",
	}

	ErrInvalidCredentials = DomainError{
		Code:    "AUTH_INVALID",
		Message: "Invalid authentication credentials",
	}

	ErrUnauthorized = DomainError{
		Code:    "AUTH_UNAUTHORIZED",
		Message: "Unauthorized access",
	}
)

// NewValidationError creates a new validation error
func NewValidationError(field string, reason string) error {
	return DomainError{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

// NewInternalError creates a new internal error
func NewInternalError(err error) error {
	return DomainError{
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == "ENV_NOT_FOUND"
	}
	return false
}

// IsDuplicate checks if the error is a duplicate error
func IsDuplicate(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == "ENV_DUPLICATE"
	}
	return false
}

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

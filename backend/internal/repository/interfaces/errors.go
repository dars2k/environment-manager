package interfaces

import "errors"

// Common repository errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidID    = errors.New("invalid ID format")
	ErrDuplicate    = errors.New("resource already exists")
)

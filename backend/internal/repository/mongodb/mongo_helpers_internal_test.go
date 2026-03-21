package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// validateStringInput tests

func TestValidateStringInput_Valid(t *testing.T) {
	result, err := validateStringInput("hello-world")
	assert.NoError(t, err)
	assert.Equal(t, "hello-world", result)
}

func TestValidateStringInput_NotString(t *testing.T) {
	_, err := validateStringInput(123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

func TestValidateStringInput_Empty(t *testing.T) {
	_, err := validateStringInput("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateStringInput_StripsInjectionChars(t *testing.T) {
	// Special chars like $ and . used in NoSQL injection should be stripped
	result, err := validateStringInput("env$name")
	assert.NoError(t, err)
	assert.Equal(t, "envname", result)
}

// validateUsername tests

func TestValidateUsername_Valid(t *testing.T) {
	result, err := validateUsername("alice123")
	assert.NoError(t, err)
	assert.Equal(t, "alice123", result)
}

func TestValidateUsername_NotString(t *testing.T) {
	_, err := validateUsername(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

func TestValidateUsername_Empty(t *testing.T) {
	_, err := validateUsername("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateUsername_InvalidChars(t *testing.T) {
	_, err := validateUsername("alice_smith")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestValidateUsername_AllAlphanumeric(t *testing.T) {
	result, err := validateUsername("Alice42")
	assert.NoError(t, err)
	assert.Equal(t, "Alice42", result)
}

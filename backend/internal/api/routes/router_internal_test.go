package routes

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOriginAllowed_NoOriginHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	assert.True(t, isOriginAllowed(req, []string{"https://example.com"}))
}

func TestIsOriginAllowed_MatchingOrigin(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	assert.True(t, isOriginAllowed(req, []string{"https://example.com", "https://other.com"}))
}

func TestIsOriginAllowed_NotInList(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	assert.False(t, isOriginAllowed(req, []string{"https://example.com"}))
}

func TestIsOriginAllowed_EmptyAllowedList(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	assert.False(t, isOriginAllowed(req, []string{}))
}

func TestGenerateClientID_Format(t *testing.T) {
	id := generateClientID()
	assert.True(t, strings.HasPrefix(id, "client-"), "should start with 'client-'")
	// hex encoding of 16 bytes = 32 hex chars
	assert.Len(t, id, len("client-")+32)
}

func TestGenerateClientID_Unique(t *testing.T) {
	id1 := generateClientID()
	id2 := generateClientID()
	assert.NotEqual(t, id1, id2)
}

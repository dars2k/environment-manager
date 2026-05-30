package database

// Tests targeting the mongo.Connect error path in NewMongoDB.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMongoDB_InvalidURIScheme covers the mongo.Connect error path by supplying
// a URI with an unsupported scheme. The MongoDB driver validates the scheme and
// returns an error before any network connection is attempted.
func TestNewMongoDB_InvalidURIScheme(t *testing.T) {
	_, err := NewMongoDB(
		"http://localhost:27017", // "http" is not a valid MongoDB URI scheme
		"testdb",
		5,
		5*time.Second,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to MongoDB")
}

// TestNewMongoDB_EmptyURI covers the mongo.Connect error path with an empty URI,
// which also fails scheme validation.
func TestNewMongoDB_EmptyURI(t *testing.T) {
	_, err := NewMongoDB(
		"", // empty URI → scheme error
		"testdb",
		5,
		5*time.Second,
	)
	require.Error(t, err)
}

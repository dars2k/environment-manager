package ctxutil_test

import (
	"context"
	"testing"

	"app-env-manager/internal/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestWithUser_And_UserFromContext(t *testing.T) {
	ctx := ctxutil.WithUser(context.Background(), "user-123", "alice")

	userID, username := ctxutil.UserFromContext(ctx)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "alice", username)
}

func TestUserFromContext_EmptyOnMissingValues(t *testing.T) {
	userID, username := ctxutil.UserFromContext(context.Background())
	assert.Empty(t, userID)
	assert.Empty(t, username)
}

func TestWithUserFull_And_RoleFromContext(t *testing.T) {
	ctx := ctxutil.WithUserFull(context.Background(), "user-456", "bob", "admin")

	userID, username := ctxutil.UserFromContext(ctx)
	assert.Equal(t, "user-456", userID)
	assert.Equal(t, "bob", username)
	assert.Equal(t, "admin", ctxutil.RoleFromContext(ctx))
}

func TestRoleFromContext_EmptyOnMissingRole(t *testing.T) {
	assert.Empty(t, ctxutil.RoleFromContext(context.Background()))
}

func TestWithUser_DoesNotSetRole(t *testing.T) {
	ctx := ctxutil.WithUser(context.Background(), "u1", "user1")
	assert.Empty(t, ctxutil.RoleFromContext(ctx))
}

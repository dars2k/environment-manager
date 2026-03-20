package ctxutil

import "context"

type contextKey string

const (
	keyUserID   contextKey = "userID"
	keyUsername contextKey = "username"
	keyRole     contextKey = "role"
)

// WithUser stores user identity (ID, username) in the context.
func WithUser(ctx context.Context, userID, username string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	return context.WithValue(ctx, keyUsername, username)
}

// WithUserFull stores user identity including role in the context.
func WithUserFull(ctx context.Context, userID, username, role string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	ctx = context.WithValue(ctx, keyUsername, username)
	return context.WithValue(ctx, keyRole, role)
}

// UserFromContext extracts userID and username from context.
// Returns empty strings when the caller is not a user (e.g. background jobs).
func UserFromContext(ctx context.Context) (userID, username string) {
	if v, ok := ctx.Value(keyUserID).(string); ok {
		userID = v
	}
	if v, ok := ctx.Value(keyUsername).(string); ok {
		username = v
	}
	return
}

// RoleFromContext extracts the user role from context.
func RoleFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(keyRole).(string); ok {
		return v
	}
	return ""
}

package ctxutil

import "context"

type contextKey string

const (
	keyUserID   contextKey = "userID"
	keyUsername contextKey = "username"
)

// WithUser stores user identity in the context.
func WithUser(ctx context.Context, userID, username string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	return context.WithValue(ctx, keyUsername, username)
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

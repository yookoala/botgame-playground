package comms

import "context"

type contextKey int

const (
	sessionIDKey contextKey = iota
)

// WithSessionID returns a new context with the session ID.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// GetSessionID returns the session ID from the context.
func GetSessionID(ctx context.Context) string {
	v := ctx.Value(sessionIDKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

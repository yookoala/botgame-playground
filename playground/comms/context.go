package comms

import "context"

type contextKey int

const (
	sessionIDKey contextKey = iota
	sessionCollectionKey
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

// WithSessionCollection returns a new context with the session collection.
func WithSessionCollection(ctx context.Context, sc SessionCollection) context.Context {
	return context.WithValue(ctx, sessionCollectionKey, sc)
}

// GetSessionCollection returns the session collection from the context.
func GetSessionCollection(ctx context.Context) SessionCollection {
	v := ctx.Value(sessionCollectionKey)
	if v == nil {
		return nil
	}
	return v.(SessionCollection)
}

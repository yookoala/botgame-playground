package comms

import (
	"context"
	"log/slog"
)

type contextKey int

const (
	sessionIDKey contextKey = iota
	sessionCollectionKey
	loggerKey
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

// WithLogger returns a new context with the logger.
func WithLogger(ctx context.Context, logger slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger returns the logger from the context.
func GetLogger(ctx context.Context) slog.Logger {
	v := ctx.Value(loggerKey)
	if v == nil {
		panic("logger not found in context")
	}
	return v.(slog.Logger)
}

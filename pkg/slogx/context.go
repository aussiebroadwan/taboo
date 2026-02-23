package slogx

import (
	"context"
	"log/slog"
)

type contextKey struct{}

// NewContext returns a new context with the logger attached.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns the default slog logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// With returns a new context with additional attributes added to the logger.
func With(ctx context.Context, args ...any) context.Context {
	logger := FromContext(ctx).With(args...)
	return NewContext(ctx, logger)
}

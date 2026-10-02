// Package logger provides structured logging using Go's standard slog package.
// All log entries include a request_id when available from context,
// and follow consistent field naming for log aggregation.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const (
	requestIDKey  contextKey = "request_id"
	organizationIDKey contextKey = "org_id"
	userIDKey     contextKey = "user_id"
)

// Logger wraps slog.Logger for convenience.
type Logger struct {
	*slog.Logger
}

// Config configures the logger.
type Config struct {
	Level     slog.Level
	Format    string // "text" | "json"
	AddSource bool
	Output    io.Writer
}

// New creates a new Logger.
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{slog.New(handler)}
}

// NewDefault creates a logger with default development settings.
func NewDefault() *Logger {
	return New(Config{
		Level:     slog.LevelDebug,
		Format:    "text",
		AddSource: false,
	})
}

// WithRequestID returns a new Logger that includes the request ID in every log entry.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{l.With("request_id", requestID)}
}

// WithOrgID returns a new Logger that includes the organization ID.
func (l *Logger) WithOrgID(orgID string) *Logger {
	return &Logger{l.With("org_id", orgID)}
}

// WithUserID returns a new Logger that includes the user ID.
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{l.With("user_id", userID)}
}

// WithContext extracts log fields from context and returns an enriched logger.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	log := l
	if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
		log = log.WithRequestID(reqID)
	}
	if orgID, ok := ctx.Value(organizationIDKey).(string); ok && orgID != "" {
		log = log.WithOrgID(orgID)
	}
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		log = log.WithUserID(userID)
	}
	return log
}

// --- Context helpers ---------------------------------------------------------

// ContextWithRequestID stores the request ID in context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// ContextWithOrgID stores the organization ID in context.
func ContextWithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, organizationIDKey, orgID)
}

// OrgIDFromContext extracts the organization ID from context.
func OrgIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(organizationIDKey).(string); ok {
		return v
	}
	return ""
}

// ContextWithUserID stores the user ID in context.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

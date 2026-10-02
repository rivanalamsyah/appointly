// Package middleware provides HTTP middleware for the Appointly API server.
// All middleware follow the Chi pattern: func(http.Handler) http.Handler
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/rs/cors"

	"github.com/appointly/appointly/backend/internal/config"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	"github.com/appointly/appointly/backend/internal/pkg/crypto"
	"github.com/appointly/appointly/backend/internal/pkg/logger"
)

// contextKey is an unexported type for middleware context keys.
type contextKey string

const (
	requestIDCtxKey contextKey = "request_id"
)

// RequestID adds a unique request ID to every request.
// The ID is read from X-Request-ID header if present, otherwise generated.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			var err error
			requestID, err = crypto.GenerateSecureTokenHex(8)
			if err != nil {
				requestID = "unknown"
			}
		}

		// Store in context and set response header
		ctx := context.WithValue(r.Context(), requestIDCtxKey, requestID)
		ctx = logger.ContextWithRequestID(ctx, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDCtxKey).(string); ok {
		return v
	}
	return ""
}

// RequestLogger logs all HTTP requests with structured fields.
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := newResponseWriter(w)

			defer func() {
				reqLog := log.WithContext(r.Context())
				reqLog.Info("http request",
					"method", r.Method,
					"path", r.URL.Path,
					"query", r.URL.RawQuery,
					"status", ww.status,
					"bytes", ww.bytes,
					"duration_ms", time.Since(start).Milliseconds(),
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// CORS configures Cross-Origin Resource Sharing.
func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-Request-ID", "X-Total-Count"},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           300,
	})
	return c.Handler
}

// RateLimit applies token bucket rate limiting for authenticated API endpoints.
func RateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	if cfg.App.Env == "development" || cfg.App.Env == "test" {
		// Skip rate limiting in development/test
		return func(next http.Handler) http.Handler { return next }
	}
	return httprate.LimitByIP(120, time.Minute)
}

// PublicRateLimit applies stricter rate limiting for public booking endpoints.
func PublicRateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	return httprate.LimitByIP(60, time.Minute)
}

// AuthRateLimit applies the strictest rate limiting for authentication endpoints.
func AuthRateLimit(cfg *config.Config) func(http.Handler) http.Handler {
	return httprate.LimitByIP(20, time.Minute)
}

// --- Response Writer Wrapper -------------------------------------------------

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// --- Response Helpers --------------------------------------------------------

// respondJSON writes a JSON response with the given status code.
// Used by handlers to send structured responses.
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	// Implementation lives in handler/v1/response.go
	// This stub is here for middleware to use if needed
	w.WriteHeader(status)
}

// RespondError writes a structured error response.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperror.AsAppError(err)
	if appErr == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(appErr.HTTPCode)
}

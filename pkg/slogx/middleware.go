package slogx

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Middleware returns an HTTP middleware that adds request logging.
// It attaches a request-scoped logger to the context with request ID,
// method, and path. It also logs the request completion with status and duration.
//
// Paths in quietPaths are logged at DEBUG instead of INFO, useful for
// high-frequency endpoints like health probes that would otherwise be noise.
func Middleware(logger *slog.Logger, quietPaths ...string) func(http.Handler) http.Handler {
	quietSet := make(map[string]struct{}, len(quietPaths))
	for _, p := range quietPaths {
		quietSet[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate or extract request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Create request-scoped logger
			reqLogger := logger.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("client_ip", clientIP(r)),
			)

			// Add logger to context
			ctx := NewContext(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			// Set request ID header on response
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request completion (DEBUG for quiet paths, INFO otherwise)
			duration := time.Since(start)
			completionAttrs := []slog.Attr{
				slog.Int("status", wrapped.status),
				slog.Duration("duration", duration),
				slog.Int("bytes", wrapped.bytes),
			}

			level := slog.LevelInfo
			if _, quiet := quietSet[r.URL.Path]; quiet {
				level = slog.LevelDebug
			}
			reqLogger.LogAttrs(ctx, level, "Request completed", completionAttrs...)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Flush implements http.Flusher, delegating to the underlying ResponseWriter if supported.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the underlying ResponseWriter for compatibility checks.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// clientIP extracts the client IP from the request, checking proxy headers first.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i != -1 {
			xff = xff[:i]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

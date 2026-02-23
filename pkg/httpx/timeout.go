package httpx

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Timeout returns middleware that applies a timeout to requests.
func Timeout(timeout time.Duration) Middleware {
	return TimeoutWithSkip(timeout)
}

// TimeoutWithSkip returns middleware that applies a timeout to requests,
// skipping requests to paths that match any of the skip patterns.
// This is useful for SSE endpoints that need long-lived connections.
func TimeoutWithSkip(timeout time.Duration, skipPaths ...string) Middleware {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipSet[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip timeout for specified paths
			if _, skip := skipSet[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})
			tw := &timeoutWriter{
				ResponseWriter: w,
				done:           done,
			}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Timeout occurred
				tw.mu.Lock()
				defer tw.mu.Unlock()
				if !tw.wroteHeader {
					http.Error(w, "request timeout", http.StatusGatewayTimeout)
				}
			}
		})
	}
}

// timeoutWriter wraps ResponseWriter to track if headers were written.
type timeoutWriter struct {
	http.ResponseWriter
	done        chan struct{}
	wroteHeader bool
	mu          sync.Mutex
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if !tw.wroteHeader {
		tw.wroteHeader = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	if !tw.wroteHeader {
		tw.wroteHeader = true
	}
	tw.mu.Unlock()
	return tw.ResponseWriter.Write(b)
}

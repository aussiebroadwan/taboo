package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middleware into a single middleware.
// Middleware is applied in the order provided.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Recoverer is middleware that recovers from panics and logs the error.
//
//nolint:contextcheck // Using r.Context() inside defer is correct for panic recovery
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := slogx.FromContext(r.Context())
				logger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

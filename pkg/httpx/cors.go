package httpx

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS middleware configuration.
type CORSConfig struct {
	// AllowedOrigins is the list of allowed origins.
	// If empty in production, no CORS headers are set.
	// In development mode, all origins are allowed.
	AllowedOrigins []string

	// Development enables permissive CORS (allow all origins).
	Development bool
}

// CORS returns middleware that handles Cross-Origin Resource Sharing.
func CORS(cfg CORSConfig) Middleware {
	allowedSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedSet[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine if origin is allowed
			var allowOrigin string
			if cfg.Development {
				// Development mode: allow all origins
				if origin != "" {
					allowOrigin = origin
				} else {
					allowOrigin = "*"
				}
			} else if origin != "" {
				// Production mode: check against allowed list
				if _, ok := allowedSet[origin]; ok {
					allowOrigin = origin
				}
			}

			// Set CORS headers if origin is allowed
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")

				// Don't set Vary for wildcard
				if allowOrigin != "*" {
					w.Header().Add("Vary", "Origin")
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSFromConfig creates a CORSConfig from application configuration values.
func CORSFromConfig(environment string, origins []string) CORSConfig {
	return CORSConfig{
		AllowedOrigins: origins,
		Development:    strings.EqualFold(environment, "development"),
	}
}

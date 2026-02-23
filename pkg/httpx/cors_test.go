package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_DevelopmentMode(t *testing.T) {
	cfg := CORSConfig{Development: true}
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		origin         string
		wantAllowAll   bool
		wantAllowed    bool
		wantVaryOrigin bool
	}{
		{
			name:           "with origin header",
			origin:         "http://example.com",
			wantAllowAll:   false,
			wantAllowed:    true,
			wantVaryOrigin: false,
		},
		{
			name:           "without origin header",
			origin:         "",
			wantAllowAll:   true,
			wantAllowed:    true,
			wantVaryOrigin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.wantAllowAll {
				if allowOrigin != "*" {
					t.Errorf("expected Access-Control-Allow-Origin = *, got %q", allowOrigin)
				}
			} else if tt.wantAllowed {
				if allowOrigin != tt.origin {
					t.Errorf("expected Access-Control-Allow-Origin = %q, got %q", tt.origin, allowOrigin)
				}
			}
		})
	}
}

func TestCORS_ProductionMode(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"http://allowed.com", "http://also-allowed.com"},
		Development:    false,
	}
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name        string
		origin      string
		wantAllowed bool
	}{
		{
			name:        "allowed origin",
			origin:      "http://allowed.com",
			wantAllowed: true,
		},
		{
			name:        "another allowed origin",
			origin:      "http://also-allowed.com",
			wantAllowed: true,
		},
		{
			name:        "disallowed origin",
			origin:      "http://evil.com",
			wantAllowed: false,
		},
		{
			name:        "no origin",
			origin:      "",
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.wantAllowed {
				if allowOrigin != tt.origin {
					t.Errorf("expected Access-Control-Allow-Origin = %q, got %q", tt.origin, allowOrigin)
				}
				if vary := rec.Header().Get("Vary"); vary != "Origin" {
					t.Errorf("expected Vary = Origin, got %q", vary)
				}
			} else {
				if allowOrigin != "" {
					t.Errorf("expected no Access-Control-Allow-Origin, got %q", allowOrigin)
				}
			}
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	cfg := CORSConfig{Development: true}
	called := false
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if called {
		t.Error("handler should not be called for preflight requests")
	}
}

func TestCORSFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		origins     []string
		wantDev     bool
	}{
		{
			name:        "development",
			environment: "development",
			origins:     nil,
			wantDev:     true,
		},
		{
			name:        "Development uppercase",
			environment: "Development",
			origins:     nil,
			wantDev:     true,
		},
		{
			name:        "production",
			environment: "production",
			origins:     []string{"http://example.com"},
			wantDev:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := CORSFromConfig(tt.environment, tt.origins)
			if cfg.Development != tt.wantDev {
				t.Errorf("Development = %v, want %v", cfg.Development, tt.wantDev)
			}
		})
	}
}

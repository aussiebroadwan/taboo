package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_AllowsWithinLimit(t *testing.T) {
	cfg := RateLimitConfig{
		Rate:  10,
		Burst: 5,
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 5 requests should succeed (burst)
	for i := range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusOK, rec.Code)
		}
	}
}

func TestRateLimit_BlocksExcessRequests(t *testing.T) {
	cfg := RateLimitConfig{
		Rate:  1,
		Burst: 2,
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the burst
	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if i < 2 && rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusOK, rec.Code)
		}
		if i >= 2 && rec.Code != http.StatusTooManyRequests {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusTooManyRequests, rec.Code)
		}
	}
}

func TestRateLimit_SeparateLimitsPerIP(t *testing.T) {
	cfg := RateLimitConfig{
		Rate:  1,
		Burst: 1,
	}
	handler := RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP uses its quota
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "192.168.1.10:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("first request from IP1: expected %d, got %d", http.StatusOK, rec1.Code)
	}

	// Second request from same IP should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "192.168.1.10:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request from IP1: expected %d, got %d", http.StatusTooManyRequests, rec2.Code)
	}

	// Different IP should still be allowed
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "192.168.1.11:12345"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Errorf("first request from IP2: expected %d, got %d", http.StatusOK, rec3.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		wantIP     string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.1:12345",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single",
			remoteAddr: "192.168.1.1:12345",
			xff:        "10.0.0.1",
			wantIP:     "10.0.0.1",
		},
		{
			name:       "X-Forwarded-For multiple",
			remoteAddr: "192.168.1.1:12345",
			xff:        "10.0.0.1, 10.0.0.2, 10.0.0.3",
			wantIP:     "10.0.0.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xri:        "10.0.0.5",
			wantIP:     "10.0.0.5",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xff:        "10.0.0.1",
			xri:        "10.0.0.5",
			wantIP:     "10.0.0.1",
		},
		{
			name:       "remote addr without port",
			remoteAddr: "192.168.1.1",
			wantIP:     "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := GetClientIP(req)
			if got != tt.wantIP {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.wantIP)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := newRateLimiter(RateLimitConfig{
		Rate:            10,
		Burst:           5,
		CleanupInterval: 10 * time.Millisecond,
		MaxAge:          50 * time.Millisecond,
	})

	// Add a limiter
	rl.getLimiter("192.168.1.100")

	rl.mu.RLock()
	if len(rl.limiters) != 1 {
		t.Errorf("expected 1 limiter, got %d", len(rl.limiters))
	}
	rl.mu.RUnlock()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	rl.mu.RLock()
	if len(rl.limiters) != 0 {
		t.Errorf("expected 0 limiters after cleanup, got %d", len(rl.limiters))
	}
	rl.mu.RUnlock()
}

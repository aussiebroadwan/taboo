package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// Rate is the number of requests allowed per second.
	Rate int

	// Burst is the maximum burst size.
	Burst int

	// CleanupInterval is how often to clean up stale limiters.
	// Defaults to 1 minute if not set.
	CleanupInterval time.Duration

	// MaxAge is how long a limiter can be idle before cleanup.
	// Defaults to 5 minutes if not set.
	MaxAge time.Duration
}

// rateLimiter manages per-IP rate limiters.
type rateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	maxAge   time.Duration
}

// limiterEntry holds a rate limiter and its last access time.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newRateLimiter creates a new rate limiter manager.
func newRateLimiter(cfg RateLimitConfig) *rateLimiter {
	maxAge := cfg.MaxAge
	if maxAge == 0 {
		maxAge = 5 * time.Minute
	}

	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = time.Minute
	}

	rl := &rateLimiter{
		limiters: make(map[string]*limiterEntry),
		rate:     rate.Limit(cfg.Rate),
		burst:    cfg.Burst,
		maxAge:   maxAge,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop(cleanupInterval)

	return rl
}

// getLimiter returns the rate limiter for the given IP.
func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	entry, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		rl.mu.Lock()
		entry.lastSeen = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists = rl.limiters[ip]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = &limiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanupLoop periodically removes stale limiters.
func (rl *rateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes limiters that haven't been accessed recently.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastSeen) > rl.maxAge {
			delete(rl.limiters, ip)
		}
	}
}

// RateLimit returns middleware that rate limits requests per IP.
func RateLimit(cfg RateLimitConfig) Middleware {
	rl := newRateLimiter(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP from the request.
// It checks X-Forwarded-For and X-Real-IP headers first,
// then falls back to the remote address.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := indexByte(xff, ','); idx != -1 {
			xff = xff[:idx]
		}
		xff = trimSpaces(xff)
		if xff != "" {
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return trimSpaces(xri)
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// indexByte returns the index of the first instance of c in s, or -1.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// trimSpaces removes leading and trailing spaces from s.
func trimSpaces(s string) string {
	start := 0
	for start < len(s) && s[start] == ' ' {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

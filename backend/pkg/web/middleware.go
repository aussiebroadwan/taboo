package web

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Use(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			next = xs[i](next)
		}
		return next
	}
}

type wrappedWritter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWritter) WriteHeader(statusCode int) {
	if statusCode == http.StatusOK {
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *wrappedWritter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := w.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func WithRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("web: recovered web call %v", rvr)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow credentials (e.g., cookies)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow the origin that matches your frontend
		// TODO: have a list of possible origins or a dynamic origin
		w.Header().Set("Access-Control-Allow-Origin", origin)

		// Allow common HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Allow headers necessary for your requests
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapped writer to capture the status code
		wrapped := &wrappedWritter{w, http.StatusOK}

		// Add the logger to the context
		next.ServeHTTP(wrapped, r)

		// Log the request
		log.Printf("web: %s %s with response code %d took %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start).String(),
		)
	})
}

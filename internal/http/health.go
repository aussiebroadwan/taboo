package http

import (
	"net/http"

	"github.com/aussiebroadwan/taboo/pkg/httpx"
)

// handleLivez is a liveness probe endpoint.
// It returns 200 OK if the server is running.
func (s *Server) handleLivez(w http.ResponseWriter, r *http.Request) {
	_ = httpx.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleReadyz is a readiness probe endpoint.
// It checks all dependencies and returns their status.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)

	// Check database
	if err := s.store.Ping(r.Context()); err != nil {
		checks["database"] = "error: " + err.Error()
	} else {
		checks["database"] = "ok"
	}

	// Check game engine
	if s.engine != nil && s.engine.IsRunning() {
		checks["engine"] = "ok"
	} else {
		checks["engine"] = "not running"
	}

	// Determine overall status
	status := "ok"
	statusCode := http.StatusOK
	for _, v := range checks {
		if v != "ok" {
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	_ = httpx.JSON(w, statusCode, map[string]any{
		"status": status,
		"checks": checks,
	})
}

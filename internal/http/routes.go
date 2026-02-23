package http

import "net/http"

// registerRoutes sets up all HTTP routes.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Health endpoints
	mux.HandleFunc("GET /livez", s.handleLivez)
	mux.HandleFunc("GET /readyz", s.handleReadyz)

	// API v1 endpoints
	mux.HandleFunc("GET /api/v1/games", s.handleListGames)
	mux.HandleFunc("GET /api/v1/games/{id}", s.handleGetGame)
	mux.HandleFunc("GET /api/v1/events", s.handleEvents)

	// Static files (catch-all, must be last)
	mux.Handle("GET /", s.staticHandler())
}

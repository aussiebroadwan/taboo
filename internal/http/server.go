package http

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/service"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/pkg/httpx"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// Server represents the HTTP server.
type Server struct {
	server      *http.Server
	logger      *slog.Logger
	store       store.Store
	cfg         *config.Config
	gameService *service.GameService
	engine      *service.Engine
}

// NewServer creates a new HTTP server.
func NewServer(cfg *config.Config, logger *slog.Logger, store store.Store, gameService *service.GameService, engine *service.Engine) *Server {
	s := &Server{
		logger:      logger,
		store:       store,
		cfg:         cfg,
		gameService: gameService,
		engine:      engine,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	// Configure CORS
	corsConfig := httpx.CORSFromConfig(cfg.Environment, cfg.Server.CORSOrigins)

	// Configure rate limiting
	rateLimitConfig := httpx.RateLimitConfig{
		Rate:  cfg.Server.RateLimit,
		Burst: cfg.Server.RateBurst,
	}

	// SSE endpoint should skip timeout and gzip
	sseEndpoint := "/api/v1/events"

	// Apply middleware chain
	handler := httpx.Chain(
		httpx.CORS(corsConfig),
		httpx.RateLimit(rateLimitConfig),
		httpx.Gzip(sseEndpoint),
		httpx.TimeoutWithSkip(cfg.Server.RequestTimeout.Duration(), sseEndpoint),
		slogx.Middleware(logger, "/livez", "/readyz"),
		httpx.Recoverer,
	)(mux)

	s.server = &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout.Duration(),
		WriteTimeout: cfg.Server.WriteTimeout.Duration(),
	}

	return s
}

// Handler returns the fully-built HTTP handler with all middleware applied.
func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

// Run starts the HTTP server and blocks until the context is cancelled.
// It performs graceful shutdown when the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	// Set base context so all request contexts are derived from the parent.
	// This ensures SSE and other long-lived handlers are cancelled on shutdown.
	s.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("HTTP server started", slog.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server...",
			slog.Duration("shutdown_timeout", s.cfg.Server.ShutdownTimeout.Duration()),
		)
	}

	// Graceful shutdown
	//nolint:contextcheck // Intentionally using Background - parent ctx is cancelled during shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout.Duration())
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck // Intentionally using Background for shutdown
		return err
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

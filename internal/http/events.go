package http

import (
	"net/http"
	"time"

	"github.com/aussiebroadwan/taboo/pkg/httpx"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// handleEvents handles GET /api/v1/events (SSE endpoint)
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// Disable write timeout for SSE (long-lived connection)
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternal("failed to disable write deadline"))
		return
	}

	// Create SSE stream
	stream := httpx.NewSSEStream(w)
	if stream == nil {
		_ = httpx.WriteError(w, httpx.ErrInternal("streaming not supported"))
		return
	}

	ctx := r.Context()

	// Subscribe to game events
	events := s.gameService.Subscribe(ctx)

	slogx.FromContext(ctx).Debug("SSE client connected")

	// Single-goroutine event loop: heartbeats and game events share one select
	// so there is no concurrent access to the SSE stream.
	heartbeat := time.NewTicker(s.cfg.Server.SSEHeartbeat.Duration())
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if err := stream.SendHeartbeat(); err != nil {
				return
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := stream.Send(event.Type, event.Data); err != nil {
				return
			}
		}
	}
}

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/pkg/httpx"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
	"github.com/aussiebroadwan/taboo/sdk"
)

// handleListGames handles GET /api/v1/games
func (s *Server) handleListGames(w http.ResponseWriter, r *http.Request) {
	// Parse cursor (default 0)
	cursor := int64(0)
	if c := r.URL.Query().Get("cursor"); c != "" {
		parsed, err := strconv.ParseInt(c, 10, 64)
		if err != nil || parsed < 0 {
			_ = httpx.WriteError(w, httpx.ErrBadRequest("invalid cursor parameter"))
			return
		}
		cursor = parsed
	}

	// Parse limit (default 20, max 100)
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 100 {
			_ = httpx.WriteError(w, httpx.ErrBadRequest("limit must be between 1 and 100"))
			return
		}
		limit = parsed
	}

	// Fetch games
	games, err := s.gameService.ListGames(r.Context(), cursor, limit+1)
	if err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternal("failed to fetch games"))
		return
	}

	// Build response
	resp := sdk.GameListResponse{
		Games: make([]sdk.Game, 0, len(games)),
	}

	// Check if there's a next page
	hasMore := len(games) > limit
	if hasMore {
		games = games[:limit]
	}

	for _, g := range games {
		resp.Games = append(resp.Games, sdk.Game{
			ID:        g.ID,
			Picks:     g.Picks,
			CreatedAt: g.CreatedAt,
		})
	}

	// Set next cursor if there are more results
	// Cursor points to the next page's starting ID (exclusive of current page)
	if hasMore && len(games) > 0 {
		nextCursor := games[len(games)-1].ID + 1
		resp.NextCursor = &nextCursor
	}

	if err := httpx.JSON(w, http.StatusOK, resp); err != nil {
		slogx.FromContext(r.Context()).Warn("Failed to write JSON response", slogx.Error(err))
	}
}

// handleGetGame handles GET /api/v1/games/{id}
func (s *Server) handleGetGame(w http.ResponseWriter, r *http.Request) {
	// Parse game ID from path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		_ = httpx.WriteError(w, httpx.ErrBadRequest("invalid game ID"))
		return
	}

	// Fetch game
	game, err := s.gameService.GetGame(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			_ = httpx.WriteError(w, httpx.ErrNotFound(fmt.Sprintf("game %d not found", id)))
			return
		}
		_ = httpx.WriteError(w, httpx.ErrInternal("failed to fetch game"))
		return
	}

	if err := httpx.JSON(w, http.StatusOK, sdk.Game{
		ID:        game.ID,
		Picks:     game.Picks,
		CreatedAt: game.CreatedAt,
	}); err != nil {
		slogx.FromContext(r.Context()).Warn("Failed to write JSON response",
			slogx.Error(err),
			slog.Int64("game_id", id),
		)
	}
}

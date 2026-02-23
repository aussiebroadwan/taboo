package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/domain"
	"github.com/aussiebroadwan/taboo/internal/service"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/sdk"
)

var errMockDB = errors.New("mock database error")

// mockStore implements store.Store for testing.
type mockStore struct {
	games      map[int64]*domain.Game
	latestGame *domain.Game

	pingErr   error
	createErr error
	getErr    error
	listErr   error
	latestErr error
}

func newMockStore() *mockStore {
	return &mockStore{
		games: make(map[int64]*domain.Game),
	}
}

func (m *mockStore) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockStore) Close() error {
	return nil
}

func (m *mockStore) CreateGame(ctx context.Context, game *domain.Game) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.games[game.ID] = game
	m.latestGame = game
	return nil
}

func (m *mockStore) GetGame(ctx context.Context, id int64) (*domain.Game, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	game, ok := m.games[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return game, nil
}

func (m *mockStore) GetLatestGame(ctx context.Context) (*domain.Game, error) {
	if m.latestErr != nil {
		return nil, m.latestErr
	}
	if m.latestGame == nil {
		return nil, store.ErrNotFound
	}
	return m.latestGame, nil
}

func (m *mockStore) ListGames(ctx context.Context, startID int64, limit int) ([]*domain.Game, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*domain.Game
	for id := startID + 1; id <= startID+int64(limit)+1; id++ {
		if game, ok := m.games[id]; ok {
			result = append(result, game)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

type testServer struct {
	*Server
	mockStore   *mockStore
	gameService *service.GameService
	engine      *service.Engine
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	store := newMockStore()
	cfg := config.Default()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gameService := service.NewGameService(store, &cfg.Game)
	engine := service.NewEngine(gameService, &cfg.Game, logger)
	server := NewServer(cfg, logger, store, gameService, engine)
	return &testServer{
		Server:      server,
		mockStore:   store,
		gameService: gameService,
		engine:      engine,
	}
}

func TestHandleListGames_DefaultParams(t *testing.T) {
	ts := newTestServer(t)

	// Add some games
	for i := int64(1); i <= 5; i++ {
		ts.mockStore.games[i] = &domain.Game{
			ID:        i,
			Picks:     []uint8{uint8(i % 256)}, //nolint:gosec // test values are within uint8 range
			CreatedAt: time.Now(),
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
	w := httptest.NewRecorder()

	ts.handleListGames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp sdk.GameListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Games) == 0 {
		t.Error("expected some games in response")
	}
}

func TestHandleListGames_WithCursor(t *testing.T) {
	ts := newTestServer(t)

	for i := int64(1); i <= 10; i++ {
		ts.mockStore.games[i] = &domain.Game{
			ID:        i,
			Picks:     []uint8{uint8(i % 256)}, //nolint:gosec // test values are within uint8 range
			CreatedAt: time.Now(),
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games?cursor=5", nil)
	w := httptest.NewRecorder()

	ts.handleListGames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp sdk.GameListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should only include games with ID > 5
	for _, g := range resp.Games {
		if g.ID <= 5 {
			t.Errorf("expected game ID > 5, got %d", g.ID)
		}
	}
}

func TestHandleListGames_WithLimit(t *testing.T) {
	ts := newTestServer(t)

	for i := int64(1); i <= 10; i++ {
		ts.mockStore.games[i] = &domain.Game{
			ID:        i,
			Picks:     []uint8{uint8(i % 256)}, //nolint:gosec // test values are within uint8 range
			CreatedAt: time.Now(),
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games?limit=3", nil)
	w := httptest.NewRecorder()

	ts.handleListGames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp sdk.GameListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Games) > 3 {
		t.Errorf("expected at most 3 games, got %d", len(resp.Games))
	}
}

func TestHandleListGames_InvalidCursor(t *testing.T) {
	ts := newTestServer(t)

	tests := []struct {
		name   string
		cursor string
	}{
		{"negative", "-1"},
		{"non-numeric", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/games?cursor="+tc.cursor, nil)
			w := httptest.NewRecorder()

			ts.handleListGames(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleListGames_InvalidLimit(t *testing.T) {
	ts := newTestServer(t)

	tests := []struct {
		name  string
		limit string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too large", "101"},
		{"non-numeric", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/games?limit="+tc.limit, nil)
			w := httptest.NewRecorder()

			ts.handleListGames(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleListGames_Pagination(t *testing.T) {
	ts := newTestServer(t)

	// Create more games than the requested limit
	for i := int64(1); i <= 25; i++ {
		ts.mockStore.games[i] = &domain.Game{
			ID:        i,
			Picks:     []uint8{uint8(i % 256)}, //nolint:gosec // test values are within uint8 range
			CreatedAt: time.Now(),
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games?limit=10", nil)
	w := httptest.NewRecorder()

	ts.handleListGames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp sdk.GameListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

}

func TestHandleListGames_StoreError(t *testing.T) {
	ts := newTestServer(t)
	ts.mockStore.listErr = errors.New("database error")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
	w := httptest.NewRecorder()

	ts.handleListGames(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandleGetGame_Success(t *testing.T) {
	ts := newTestServer(t)

	game := &domain.Game{
		ID:        42,
		Picks:     []uint8{1, 2, 3, 4, 5},
		CreatedAt: time.Now(),
	}
	ts.mockStore.games[42] = game

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games/42", nil)
	req.SetPathValue("id", "42")
	w := httptest.NewRecorder()

	ts.handleGetGame(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp sdk.Game
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("expected ID 42, got %d", resp.ID)
	}
}

func TestHandleGetGame_NotFound(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	ts.handleGetGame(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleGetGame_InvalidID(t *testing.T) {
	ts := newTestServer(t)

	tests := []struct {
		name string
		id   string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"non-numeric", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/games/"+tc.id, nil)
			req.SetPathValue("id", tc.id)
			w := httptest.NewRecorder()

			ts.handleGetGame(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleGetGame_StoreError(t *testing.T) {
	ts := newTestServer(t)
	ts.mockStore.getErr = errors.New("database error")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/games/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	ts.handleGetGame(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

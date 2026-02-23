package sdk_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/internal/config"
	taboohttp "github.com/aussiebroadwan/taboo/internal/http"
	"github.com/aussiebroadwan/taboo/internal/service"
	"github.com/aussiebroadwan/taboo/internal/store/drivers/sqlite"
	"github.com/aussiebroadwan/taboo/sdk"
)

// testServer wraps an httptest.Server with the game engine and services.
type testServer struct {
	Server      *httptest.Server
	URL         string
	GameService *service.GameService
	Engine      *service.Engine
	cancel      context.CancelFunc
}

// setupTestServer creates a test server with a temporary store and fast game timings.
func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	// Create a temp file for SQLite (in-memory doesn't work well with concurrent access)
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Fast test configuration
	cfg := &config.Config{
		Environment: "development",
		Server: config.ServerConfig{
			Host:            "127.0.0.1",
			Port:            0,
			ReadTimeout:     config.Duration(30 * time.Second),
			WriteTimeout:    config.Duration(30 * time.Second),
			ShutdownTimeout: config.Duration(5 * time.Second),
			SSEHeartbeat:    config.Duration(100 * time.Millisecond),
			RequestTimeout:  config.Duration(30 * time.Second),
			CORSOrigins:     []string{"*"},
			RateLimit:       1000,
			RateBurst:       100,
		},
		Game: config.GameConfig{
			DrawDuration: config.Duration(150 * time.Millisecond), // 50ms per pick with 3 picks
			WaitDuration: config.Duration(50 * time.Millisecond),
			PickCount:    3,
			MaxNumber:    10,
		},
	}

	logger := slog.New(slog.NewTextHandler(testWriter{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create services
	gameService := service.NewGameService(store, &cfg.Game)
	engine := service.NewEngine(gameService, &cfg.Game, logger)

	// Use the real HTTP server handler (routes + middleware)
	srv := taboohttp.NewServer(cfg, logger, store, gameService, engine)
	ts := httptest.NewServer(srv.Handler())

	// Start engine in background
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		engine.Run(ctx)
	}()

	// Register cleanup
	t.Cleanup(func() {
		cancel()
		ts.Close()
		store.Close()
	})

	return &testServer{
		Server:      ts,
		URL:         ts.URL,
		GameService: gameService,
		Engine:      engine,
		cancel:      cancel,
	}
}

// testWriter adapts testing.T to io.Writer for slog.
type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

// waitForGames waits for at least n games to be created.
func waitForGames(t *testing.T, ctx context.Context, client *sdk.Client, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.ListGames(ctx, nil)
		if err == nil && len(resp.Games) >= n {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %d games", n)
}

// --- REST API Integration Tests ---

func TestIntegration_ListGames(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	client := sdk.NewClient(ts.URL)
	ctx := context.Background()

	// Wait for at least 2 games so game 1 is no longer active (picks are hidden for active game)
	waitForGames(t, ctx, client, 2)

	// List games
	resp, err := client.ListGames(ctx, nil)
	if err != nil {
		t.Fatalf("ListGames failed: %v", err)
	}

	if len(resp.Games) < 2 {
		t.Fatalf("expected at least 2 games, got %d", len(resp.Games))
	}

	// Verify completed game structure (first game is no longer active)
	game := resp.Games[0]
	if game.ID < 1 {
		t.Errorf("expected game ID >= 1, got %d", game.ID)
	}
	if len(game.Picks) != 3 {
		t.Errorf("expected 3 picks, got %d", len(game.Picks))
	}
	if game.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// Verify active game has picks hidden
	activeGame := resp.Games[len(resp.Games)-1]
	if len(activeGame.Picks) != 0 {
		t.Errorf("expected 0 picks for active game, got %d", len(activeGame.Picks))
	}
}

func TestIntegration_GetGame(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	client := sdk.NewClient(ts.URL)
	ctx := context.Background()

	// Wait for at least 2 games so game 1 is no longer active (picks are hidden for active game)
	waitForGames(t, ctx, client, 2)

	// Get the first game (completed, not active)
	game, err := client.GetGame(ctx, 1)
	if err != nil {
		t.Fatalf("GetGame failed: %v", err)
	}

	if game.ID != 1 {
		t.Errorf("expected game ID 1, got %d", game.ID)
	}
	if len(game.Picks) != 3 {
		t.Errorf("expected 3 picks, got %d", len(game.Picks))
	}

	// Verify picks are valid (1-10)
	for _, pick := range game.Picks {
		if pick < 1 || pick > 10 {
			t.Errorf("pick %d out of range [1, 10]", pick)
		}
	}
}

func TestIntegration_GetGame_NotFound(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	client := sdk.NewClient(ts.URL)
	ctx := context.Background()

	// Try to get a non-existent game
	_, err := client.GetGame(ctx, 999999)
	if err == nil {
		t.Fatal("expected error for non-existent game")
	}

	var apiErr *sdk.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestIntegration_Pagination(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	client := sdk.NewClient(ts.URL)
	ctx := context.Background()

	// Wait for at least 3 games
	waitForGames(t, ctx, client, 3)

	// List with limit 1
	resp, err := client.ListGames(ctx, &sdk.ListGamesOptions{
		Limit: new(1),
	})
	if err != nil {
		t.Fatalf("ListGames failed: %v", err)
	}

	if len(resp.Games) != 1 {
		t.Errorf("expected 1 game, got %d", len(resp.Games))
	}
	if resp.NextCursor == nil {
		t.Error("expected next cursor for pagination")
	}

	firstGameID := resp.Games[0].ID

	// Get next page using cursor
	resp2, err := client.ListGames(ctx, &sdk.ListGamesOptions{
		Cursor: resp.NextCursor,
		Limit:  new(1),
	})
	if err != nil {
		t.Fatalf("ListGames (page 2) failed: %v", err)
	}

	if len(resp2.Games) != 1 {
		t.Errorf("expected 1 game on page 2, got %d", len(resp2.Games))
	}
	if resp2.Games[0].ID == firstGameID {
		t.Error("page 2 returned same game as page 1")
	}
}

// --- SSE Integration Tests ---

// testEventHandler collects SSE events for testing.
type testEventHandler struct {
	sdk.BaseEventHandler
	mu         sync.Mutex
	states     []sdk.GameStateEvent
	picks      []sdk.GamePickEvent
	completes  []sdk.GameCompleteEvent
	heartbeats int
	connected  chan struct{}
}

func newTestEventHandler() *testEventHandler {
	return &testEventHandler{
		connected: make(chan struct{}, 1),
	}
}

func (h *testEventHandler) OnGameState(e sdk.GameStateEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.states = append(h.states, e)
}

func (h *testEventHandler) OnGamePick(e sdk.GamePickEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.picks = append(h.picks, e)
}

func (h *testEventHandler) OnGameComplete(e sdk.GameCompleteEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.completes = append(h.completes, e)
}

func (h *testEventHandler) OnHeartbeat() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.heartbeats++
}

func (h *testEventHandler) OnConnect() {
	select {
	case h.connected <- struct{}{}:
	default:
	}
}

func (h *testEventHandler) getStats() (states, picks, completes, heartbeats int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.states), len(h.picks), len(h.completes), h.heartbeats
}

func TestIntegration_SSE_ReceivesEvents(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	handler := newTestEventHandler()
	sseClient := sdk.NewSSEClient(ts.URL, handler,
		sdk.WithMaxRetries(1),
		sdk.WithReconnectDelay(50*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Connect in background
	done := make(chan error, 1)
	go func() {
		done <- sseClient.Connect(ctx)
	}()

	// Wait for connection
	select {
	case <-handler.connected:
		// Connected
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SSE connection")
	}

	// Wait for some events
	time.Sleep(500 * time.Millisecond)

	states, picks, _, heartbeats := handler.getStats()

	// Should receive at least one state event and some heartbeats
	if states < 1 {
		t.Errorf("expected at least 1 state event, got %d", states)
	}
	if heartbeats < 1 {
		t.Errorf("expected at least 1 heartbeat, got %d", heartbeats)
	}
	if picks < 1 {
		t.Errorf("expected at least 1 pick event, got %d", picks)
	}

	cancel()
	<-done
}

func TestIntegration_SSE_GameLifecycle(t *testing.T) {
	t.Parallel()

	ts := setupTestServer(t)
	handler := newTestEventHandler()
	sseClient := sdk.NewSSEClient(ts.URL, handler,
		sdk.WithMaxRetries(1),
		sdk.WithReconnectDelay(50*time.Millisecond),
	)

	// Connect for a full game cycle (draw + wait = ~200ms, with some buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- sseClient.Connect(ctx)
	}()

	// Wait for connection
	select {
	case <-handler.connected:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for connection")
	}

	// Wait for at least one complete game cycle
	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		_, _, completes, _ := handler.getStats()
		if completes >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	states, picks, completes, _ := handler.getStats()

	// Should have seen a full game lifecycle
	if states < 1 {
		t.Errorf("expected at least 1 state event, got %d", states)
	}
	if picks < 3 {
		t.Errorf("expected at least 3 pick events (one full game), got %d", picks)
	}
	if completes < 1 {
		t.Errorf("expected at least 1 complete event, got %d", completes)
	}

	cancel()
	<-done
}

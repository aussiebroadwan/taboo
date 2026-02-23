package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/domain"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/sdk"
)

// mockStore implements store.Store for testing.
type mockStore struct {
	games      map[int64]*domain.Game
	latestGame *domain.Game

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
	return nil
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
	for _, g := range m.games {
		if g.ID > startID {
			result = append(result, g)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func defaultGameConfig() *config.GameConfig {
	return &config.GameConfig{
		DrawDuration: config.Duration(90 * time.Second),
		WaitDuration: config.Duration(90 * time.Second),
		PickCount:    20,
		MaxNumber:    80,
	}
}

func TestGameService_GetGame_Success(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	game := &domain.Game{
		ID:        1,
		Picks:     []uint8{1, 2, 3, 4, 5},
		CreatedAt: time.Now(),
	}
	store.games[1] = game

	result, err := svc.GetGame(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != game.ID {
		t.Errorf("expected ID %d, got %d", game.ID, result.ID)
	}
}

func TestGameService_GetGame_NotFound(t *testing.T) {
	ms := newMockStore()
	svc := NewGameService(ms, defaultGameConfig())

	_, err := svc.GetGame(context.Background(), 999)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected store.ErrNotFound, got %v", err)
	}
}

func TestGameService_ListGames_Success(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	for i := int64(1); i <= 5; i++ {
		store.games[i] = &domain.Game{ID: i, Picks: []uint8{uint8(i % 256)}} //nolint:gosec // test values are within uint8 range
	}

	games, err := svc.ListGames(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(games) != 5 {
		t.Errorf("expected 5 games, got %d", len(games))
	}
}

func TestGameService_ListGames_Empty(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	games, err := svc.ListGames(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(games) != 0 {
		t.Errorf("expected 0 games, got %d", len(games))
	}
}

func TestGameService_CreateGame_Success(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	game := &domain.Game{
		ID:        1,
		Picks:     []uint8{10, 20, 30},
		CreatedAt: time.Now(),
	}

	err := svc.CreateGame(context.Background(), game)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.games[1] == nil {
		t.Error("game was not persisted")
	}
}

func TestGameService_GetLatestGame(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	game := &domain.Game{
		ID:        42,
		Picks:     []uint8{1, 2, 3},
		CreatedAt: time.Now(),
	}
	store.latestGame = game

	result, err := svc.GetLatestGame(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != game.ID {
		t.Errorf("expected ID %d, got %d", game.ID, result.ID)
	}
}

func TestGameService_Subscribe(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := svc.Subscribe(ctx)
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
}

func TestGameService_BroadcastState(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := svc.Subscribe(ctx)

	state := sdk.GameStateEvent{
		GameID:   1,
		Picks:    sdk.Picks{1, 2, 3},
		NextGame: time.Now().Add(90 * time.Second),
	}
	svc.BroadcastState(state)

	select {
	case event := <-ch:
		if event.Type != sdk.EventGameState {
			t.Errorf("expected type %s, got %s", sdk.EventGameState, event.Type)
		}
		data, ok := event.Data.(sdk.GameStateEvent)
		if !ok {
			t.Fatal("unexpected data type")
		}
		if data.GameID != state.GameID {
			t.Errorf("expected GameID %d, got %d", state.GameID, data.GameID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestGameService_BroadcastPick(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := svc.Subscribe(ctx)

	svc.BroadcastPick(42)

	select {
	case event := <-ch:
		if event.Type != sdk.EventGamePick {
			t.Errorf("expected type %s, got %s", sdk.EventGamePick, event.Type)
		}
		data, ok := event.Data.(sdk.GamePickEvent)
		if !ok {
			t.Fatal("unexpected data type")
		}
		if data.Pick != 42 {
			t.Errorf("expected Pick 42, got %d", data.Pick)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestGameService_BroadcastComplete(t *testing.T) {
	store := newMockStore()
	svc := NewGameService(store, defaultGameConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := svc.Subscribe(ctx)

	svc.BroadcastComplete(123)

	select {
	case event := <-ch:
		if event.Type != sdk.EventGameComplete {
			t.Errorf("expected type %s, got %s", sdk.EventGameComplete, event.Type)
		}
		data, ok := event.Data.(sdk.GameCompleteEvent)
		if !ok {
			t.Fatal("unexpected data type")
		}
		if data.GameID != 123 {
			t.Errorf("expected GameID 123, got %d", data.GameID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestGameService_CreateGame_StoreError(t *testing.T) {
	store := newMockStore()
	store.createErr = errors.New("database error")
	svc := NewGameService(store, defaultGameConfig())

	err := svc.CreateGame(context.Background(), &domain.Game{ID: 1})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGameService_GetGame_StoreError(t *testing.T) {
	store := newMockStore()
	store.getErr = errors.New("database error")
	svc := NewGameService(store, defaultGameConfig())

	_, err := svc.GetGame(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGameService_ListGames_StoreError(t *testing.T) {
	store := newMockStore()
	store.listErr = errors.New("database error")
	svc := NewGameService(store, defaultGameConfig())

	_, err := svc.ListGames(context.Background(), 0, 10)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGameService_GetLatestGame_StoreError(t *testing.T) {
	store := newMockStore()
	store.latestErr = errors.New("database error")
	svc := NewGameService(store, defaultGameConfig())

	_, err := svc.GetLatestGame(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

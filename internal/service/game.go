package service

import (
	"context"
	"sync/atomic"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/domain"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/pkg/pubsub"
	"github.com/aussiebroadwan/taboo/sdk"
)

// Event represents a game event to be broadcast to subscribers.
type Event struct {
	Type string
	Data any
}

// GameService handles game business logic and event broadcasting.
type GameService struct {
	store  store.Store
	config *config.GameConfig
	broker *pubsub.Broker[Event]

	activeGameID atomic.Int64
}

// NewGameService creates a new GameService.
func NewGameService(store store.Store, cfg *config.GameConfig) *GameService {
	return &GameService{
		store:  store,
		config: cfg,
		broker: pubsub.New[Event](),
	}
}

// Subscribe returns a channel that receives game events.
// The caller should cancel the context when done to unsubscribe.
func (s *GameService) Subscribe(ctx context.Context) <-chan Event {
	return s.broker.Subscribe(ctx)
}

// Broadcast sends an event to all subscribers.
func (s *GameService) Broadcast(event Event) {
	s.broker.Publish(event)
}

// BroadcastState broadcasts a game state event.
func (s *GameService) BroadcastState(state sdk.GameStateEvent) {
	s.Broadcast(Event{
		Type: sdk.EventGameState,
		Data: state,
	})
}

// BroadcastPick broadcasts a pick event.
func (s *GameService) BroadcastPick(pick uint8) {
	s.Broadcast(Event{
		Type: sdk.EventGamePick,
		Data: sdk.GamePickEvent{Pick: pick},
	})
}

// BroadcastComplete broadcasts a game complete event.
func (s *GameService) BroadcastComplete(gameID int64) {
	s.Broadcast(Event{
		Type: sdk.EventGameComplete,
		Data: sdk.GameCompleteEvent{GameID: gameID},
	})
}

// SetActiveGameID sets the currently active game ID. Picks for the active
// game are hidden from query results until the next game starts.
func (s *GameService) SetActiveGameID(id int64) {
	s.activeGameID.Store(id)
}

// GetGame retrieves a game by ID. Picks are hidden if the game is still active.
func (s *GameService) GetGame(ctx context.Context, id int64) (*domain.Game, error) {
	game, err := s.store.GetGame(ctx, id)
	if err != nil {
		return nil, err
	}
	if game.ID == s.activeGameID.Load() {
		game.Picks = nil
	}
	return game, nil
}

// ListGames retrieves games with cursor pagination. Picks are hidden for the
// active game.
func (s *GameService) ListGames(ctx context.Context, cursor int64, limit int) ([]*domain.Game, error) {
	games, err := s.store.ListGames(ctx, cursor, limit)
	if err != nil {
		return nil, err
	}
	activeID := s.activeGameID.Load()
	for _, g := range games {
		if g.ID == activeID {
			g.Picks = nil
		}
	}
	return games, nil
}

// CreateGame persists a new game.
func (s *GameService) CreateGame(ctx context.Context, game *domain.Game) error {
	return s.store.CreateGame(ctx, game)
}

// GetLatestGame retrieves the most recent game.
func (s *GameService) GetLatestGame(ctx context.Context) (*domain.Game, error) {
	return s.store.GetLatestGame(ctx)
}

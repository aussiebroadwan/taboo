package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/aussiebroadwan/taboo/internal/store"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/domain"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
	"github.com/aussiebroadwan/taboo/sdk"
)

// Engine runs the game loop, generating picks and broadcasting events.
type Engine struct {
	gameService *GameService
	config      *config.GameConfig
	logger      *slog.Logger

	running atomic.Bool
}

// NewEngine creates a new game engine.
func NewEngine(gameService *GameService, cfg *config.GameConfig, logger *slog.Logger) *Engine {
	return &Engine{
		gameService: gameService,
		config:      cfg,
		logger:      logger.With(slog.String("component", "engine")),
	}
}

// IsRunning returns whether the engine is currently running.
func (e *Engine) IsRunning() bool {
	return e.running.Load()
}

// SetRunning sets the running state. This is primarily for testing.
func (e *Engine) SetRunning(running bool) {
	e.running.Store(running)
}

// Run starts the game loop. It blocks until the context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	e.running.Store(true)
	defer e.running.Store(false)

	e.logger.Info("Game engine started",
		slog.Duration("draw_duration", e.config.DrawDuration.Duration()),
		slog.Duration("wait_duration", e.config.WaitDuration.Duration()),
		slog.Int("pick_count", e.config.PickCount),
		slog.Int("max_number", e.config.MaxNumber),
	)

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Game engine stopped")
			return ctx.Err()
		default:
			if err := e.runGame(ctx); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				e.logger.Warn("Game cycle failed", slogx.Error(err))
			}
		}
	}
}

// runGame executes a single game cycle: draw phase -> complete -> wait phase.
func (e *Engine) runGame(ctx context.Context) error {
	// Generate all picks at the start
	picks := e.generatePicks()

	// Calculate timing
	drawDuration := e.config.DrawDuration.Duration()
	waitDuration := e.config.WaitDuration.Duration()
	pickInterval := drawDuration / time.Duration(e.config.PickCount)
	nextGame := time.Now().Add(drawDuration + waitDuration)

	// Get next game ID
	nextID := int64(1)
	latestGame, err := e.gameService.GetLatestGame(ctx)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}
	if latestGame != nil {
		nextID = latestGame.ID + 1
	}

	// Create and persist the game
	game := domain.NewGame(nextID, picks)
	if err := e.gameService.CreateGame(ctx, game); err != nil {
		return err
	}

	e.logger.Info("Game started",
		slog.Int64("game_id", game.ID),
		slog.Int("picks", len(picks)),
	)

	// Broadcast initial state (no picks revealed yet)
	e.gameService.BroadcastState(sdk.GameStateEvent{
		GameID:   game.ID,
		Picks:    []uint8{},
		NextGame: nextGame,
	})

	// Draw phase: reveal picks one by one
	for i, pick := range picks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pickInterval):
			e.gameService.BroadcastPick(pick)

			// Also broadcast updated state with all revealed picks so far
			e.gameService.BroadcastState(sdk.GameStateEvent{
				GameID:   game.ID,
				Picks:    picks[:i+1],
				NextGame: nextGame,
			})
		}
	}

	// Game complete
	e.logger.Info("Game complete", slog.Int64("game_id", game.ID))
	e.gameService.BroadcastComplete(game.ID)

	// Wait phase
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		return nil
	}
}

// generatePicks generates random unique picks for a game.
func (e *Engine) generatePicks() []uint8 {
	// Create a pool of all possible numbers
	pool := make([]uint8, e.config.MaxNumber)
	for i := range pool {
		pool[i] = uint8(i + 1) //nolint:gosec // MaxNumber is validated <= 80, fits in uint8
	}

	// Fisher-Yates shuffle using crypto/rand for secure randomness
	for i := len(pool) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		pool[i], pool[j] = pool[j], pool[i]
	}

	// Take the first PickCount numbers
	return pool[:e.config.PickCount]
}

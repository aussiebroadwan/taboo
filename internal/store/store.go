package store

import (
	"context"
	"errors"

	"github.com/aussiebroadwan/taboo/internal/domain"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// Store defines the interface for data persistence.
type Store interface {
	// Ping checks the database connection.
	Ping(ctx context.Context) error

	// Close closes the database connection.
	Close() error

	// CreateGame persists a new game.
	CreateGame(ctx context.Context, game *domain.Game) error

	// GetGame retrieves a game by its ID.
	GetGame(ctx context.Context, id int64) (*domain.Game, error)

	// GetLatestGame retrieves the most recent game.
	GetLatestGame(ctx context.Context) (*domain.Game, error)

	// ListGames retrieves games starting from a given ID with a limit.
	ListGames(ctx context.Context, startID int64, limit int) ([]*domain.Game, error)
}

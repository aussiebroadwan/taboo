package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aussiebroadwan/taboo/internal/domain"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/internal/store/drivers/sqlite/gen"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store implements store.Store using SQLite.
type Store struct {
	db      *sql.DB
	queries *gen.Queries
}

// OpenDB opens a database connection without running migrations.
// This is useful for CLI commands that need direct database access.
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	return db, nil
}

// NewMigrate creates a new migrate instance for CLI migration commands.
func NewMigrate(db *sql.DB) (*migrate.Migrate, error) {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("creating migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}

	return m, nil
}

// New creates a new SQLite store and runs migrations.
func New(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &Store{
		db:      db,
		queries: gen.New(db),
	}, nil
}

func runMigrations(db *sql.DB) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// Ensure Store implements store.Store.
var _ store.Store = (*Store)(nil)

// Ping checks the database connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// CreateGame persists a new game.
func (s *Store) CreateGame(ctx context.Context, game *domain.Game) error {
	picks, err := json.Marshal(game.Picks)
	if err != nil {
		return fmt.Errorf("marshaling picks: %w", err)
	}

	err = s.queries.CreateGame(ctx, gen.CreateGameParams{
		GameID: game.ID,
		Picks:  string(picks),
	})
	if err != nil {
		return fmt.Errorf("inserting game: %w", err)
	}

	return nil
}

// GetGame retrieves a game by its ID.
func (s *Store) GetGame(ctx context.Context, id int64) (*domain.Game, error) {
	row, err := s.queries.GetGameByGameID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("getting game: %w", err)
	}

	return rowToGame(row)
}

// GetLatestGame retrieves the most recent game.
func (s *Store) GetLatestGame(ctx context.Context) (*domain.Game, error) {
	row, err := s.queries.GetLatestGame(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("getting latest game: %w", err)
	}

	return rowToGame(gen.GetGameByGameIDRow(row))
}

// ListGames retrieves games starting from a given ID with a limit.
func (s *Store) ListGames(ctx context.Context, startID int64, limit int) ([]*domain.Game, error) {
	rows, err := s.queries.GetGamesByRange(ctx, gen.GetGamesByRangeParams{
		Start: startID,
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("querying games: %w", err)
	}

	games := make([]*domain.Game, 0, len(rows))
	for _, row := range rows {
		game, err := rowToGame(gen.GetGameByGameIDRow(row))
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

// rowToGame converts a generated query row to a domain.Game.
func rowToGame(row gen.GetGameByGameIDRow) (*domain.Game, error) {
	var picks []uint8
	if err := json.Unmarshal([]byte(row.Picks), &picks); err != nil {
		return nil, fmt.Errorf("unmarshaling picks: %w", err)
	}

	return &domain.Game{
		ID:        row.GameID,
		Picks:     picks,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

package app

import (
	"fmt"
	"log/slog"

	"github.com/aussiebroadwan/taboo/internal/config"
	"github.com/aussiebroadwan/taboo/internal/store"
	"github.com/aussiebroadwan/taboo/internal/store/drivers/sqlite"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// Version information, set at build time.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// App holds application dependencies.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	Store  store.Store
}

// New creates a new App with all dependencies initialized.
func New(configPath, logLevel string, verbose bool) (*App, error) {
	// Determine effective log level
	effectiveLevel := logLevel
	if verbose && logLevel == "" {
		effectiveLevel = "debug"
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Override log level if specified via CLI
	if effectiveLevel != "" {
		cfg.Logging.Level = effectiveLevel
	}

	// Create logger
	logger := slogx.New(
		slogx.WithLevel(slogx.ParseLevel(cfg.Logging.Level)),
		slogx.WithFormat(slogx.ParseFormat(cfg.Logging.Format)),
		slogx.WithService("taboo"),
		slogx.WithVersion(Version),
	)

	// Create store
	var st store.Store
	switch cfg.Database.Driver {
	case "sqlite":
		st, err = sqlite.New(cfg.Database.DSN)
		if err != nil {
			return nil, fmt.Errorf("creating sqlite store: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	logger.Info("Application initialized",
		slog.String("version", Version),
		slog.String("log_level", cfg.Logging.Level),
	)

	return &App{
		Config: cfg,
		Logger: logger,
		Store:  st,
	}, nil
}

// Close releases all application resources.
func (a *App) Close() error {
	if a.Store != nil {
		return a.Store.Close()
	}
	return nil
}

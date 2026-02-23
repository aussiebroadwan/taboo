package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aussiebroadwan/taboo/internal/http"
	"github.com/aussiebroadwan/taboo/internal/service"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// RunServe runs the serve command.
func RunServe(configPath, logLevel string, verbose bool) error {
	// Create application
	app, err := New(configPath, logLevel, verbose)
	if err != nil {
		return err
	}
	defer func() {
		if err := app.Close(); err != nil {
			app.Logger.Error("Failed to close application", slogx.Error(err))
		}
	}()

	// Create game service and engine
	gameService := service.NewGameService(app.Store, &app.Config.Game)
	engine := service.NewEngine(gameService, &app.Config.Game, app.Logger)

	// Create HTTP server
	server := http.NewServer(app.Config, app.Logger, app.Store, gameService, engine)

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start game engine in background
	go func() {
		if err := engine.Run(ctx); err != nil && ctx.Err() == nil {
			app.Logger.Error("Game engine failed",
				slogx.Error(err),
				slog.String("component", "engine"),
			)
		}
	}()

	// Run server
	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

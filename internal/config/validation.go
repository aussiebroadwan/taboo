package config

import (
	"strings"

	"github.com/aussiebroadwan/taboo/pkg/lint"
)

// Lint checks the configuration and returns all issues (errors, warnings, info).
func Lint(cfg *Config) lint.Issues {
	c := lint.NewCollector()

	lintEnvironment(c, cfg)
	lintServer(c, cfg)
	lintGame(c, cfg)
	lintDatabase(c, cfg)
	lintLogging(c, cfg)
	lintDiscord(c, cfg)

	return c.Issues()
}

// Validate checks the configuration and returns an error if any errors are found.
func Validate(cfg *Config) error {
	return Lint(cfg).Errors().Err()
}

func lintEnvironment(c *lint.Collector, cfg *Config) {
	env := strings.ToLower(cfg.Environment)
	switch env {
	case "development":
		c.Warn("dev-mode-cors", "environment", "running in development mode (CORS allows all origins)")
	case "production":
		// Valid, no issues
	default:
		c.Errorf("env-invalid", "environment", "must be 'development' or 'production', got %q", cfg.Environment)
	}
}

func lintServer(c *lint.Collector, cfg *Config) {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		c.Errorf("port-invalid", "server.port", "must be between 1 and 65535, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout.Duration() <= 0 {
		c.Error("timeout-invalid", "server.read_timeout", "must be positive")
	}
	if cfg.Server.WriteTimeout.Duration() <= 0 {
		c.Error("timeout-invalid", "server.write_timeout", "must be positive")
	}
	if cfg.Server.ShutdownTimeout.Duration() <= 0 {
		c.Error("timeout-invalid", "server.shutdown_timeout", "must be positive")
	}
	if cfg.Server.RequestTimeout.Duration() <= 0 {
		c.Error("timeout-invalid", "server.request_timeout", "must be positive")
	}
	if cfg.Server.RateLimit < 1 {
		c.Errorf("rate-limit-invalid", "server.rate_limit", "must be at least 1, got %d", cfg.Server.RateLimit)
	}
	if cfg.Server.RateBurst < 1 {
		c.Errorf("rate-limit-invalid", "server.rate_burst", "must be at least 1, got %d", cfg.Server.RateBurst)
	}
}

func lintGame(c *lint.Collector, cfg *Config) {
	if cfg.Game.PickCount < 1 {
		c.Errorf("game-invalid", "game.pick_count", "must be at least 1, got %d", cfg.Game.PickCount)
	}
	if cfg.Game.MaxNumber < cfg.Game.PickCount {
		c.Errorf("game-invalid", "game.max_number", "must be >= pick_count (%d), got %d", cfg.Game.PickCount, cfg.Game.MaxNumber)
	}
	if cfg.Game.DrawDuration.Duration() <= 0 {
		c.Error("timeout-invalid", "game.draw_duration", "must be positive")
	}
	if cfg.Game.WaitDuration.Duration() <= 0 {
		c.Error("timeout-invalid", "game.wait_duration", "must be positive")
	}
}

func lintDatabase(c *lint.Collector, cfg *Config) {
	if cfg.Database.Driver == "" {
		c.Error("db-invalid", "database.driver", "is required")
	} else if cfg.Database.Driver != "sqlite" {
		c.Errorf("db-invalid", "database.driver", "must be 'sqlite', got %q", cfg.Database.Driver)
	}

	//nolint:staticcheck // if/else is clearer than switch for empty + specific value check
	if cfg.Database.DSN == "" {
		c.Error("db-invalid", "database.dsn", "is required")
	} else if cfg.Database.DSN == ":memory:" {
		c.Warn("db-memory", "database.dsn", "using in-memory database (data will be lost on restart)")
	}
}

func lintLogging(c *lint.Collector, cfg *Config) {
	level := strings.ToLower(cfg.Logging.Level)
	switch level {
	case "debug":
		c.Warn("debug-logging", "logging.level", "debug logging enabled (may impact performance)")
	case "info", "warn", "error":
		// Valid
	default:
		c.Errorf("logging-invalid", "logging.level", "must be one of: debug, info, warn, error; got %q", cfg.Logging.Level)
	}

	format := strings.ToLower(cfg.Logging.Format)
	if format != "text" && format != "json" {
		c.Errorf("logging-invalid", "logging.format", "must be one of: text, json; got %q", cfg.Logging.Format)
	}
}

func lintDiscord(c *lint.Collector, cfg *Config) {
	if cfg.Discord.ClientID == "" || cfg.Discord.ClientSecret == "" {
		c.Warn("discord-missing", "discord", "Discord credentials not configured (Discord Activity will not work)")
	}
}

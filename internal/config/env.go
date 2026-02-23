package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// applyEnv applies environment variable overrides to the config.
// Environment variables take precedence over config file values.
func applyEnv(cfg *Config) {
	// Environment
	if v := os.Getenv("TABOO_ENVIRONMENT"); v != "" {
		cfg.Environment = v
	}

	// Server
	if v := os.Getenv("TABOO_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("TABOO_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("TABOO_SERVER_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_SERVER_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_SERVER_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ShutdownTimeout = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_SERVER_REQUEST_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.RequestTimeout = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_SERVER_CORS_ORIGINS"); v != "" {
		cfg.Server.CORSOrigins = splitAndTrim(v, ",")
	}
	if v := os.Getenv("TABOO_SERVER_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.RateLimit = n
		}
	}
	if v := os.Getenv("TABOO_SERVER_RATE_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.RateBurst = n
		}
	}

	// Game
	if v := os.Getenv("TABOO_GAME_DRAW_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Game.DrawDuration = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_GAME_WAIT_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Game.WaitDuration = Duration(d)
		}
	}
	if v := os.Getenv("TABOO_GAME_PICK_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Game.PickCount = n
		}
	}
	if v := os.Getenv("TABOO_GAME_MAX_NUMBER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Game.MaxNumber = n
		}
	}

	// Database
	if v := os.Getenv("TABOO_DATABASE_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}
	if v := os.Getenv("TABOO_DATABASE_DSN"); v != "" {
		cfg.Database.DSN = v
	}

	// Logging
	if v := os.Getenv("TABOO_LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("TABOO_LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}

	// Discord
	if v := os.Getenv("DISCORD_CLIENT_ID"); v != "" {
		cfg.Discord.ClientID = v
	}
	if v := os.Getenv("DISCORD_CLIENT_SECRET"); v != "" {
		cfg.Discord.ClientSecret = v
	}
}

// splitAndTrim splits a string by separator and trims whitespace from each part.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

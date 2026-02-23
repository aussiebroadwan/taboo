package config

import (
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Valid configs
		{"valid full config", testdataPath("valid_full.yaml"), false},
		{"valid minimal config", testdataPath("valid_minimal.yaml"), false},
		{"valid memory db", testdataPath("valid_memory_db.yaml"), false},

		// Invalid configs
		{"invalid environment", testdataPath("invalid_environment.yaml"), true},
		{"invalid port zero", testdataPath("invalid_port_zero.yaml"), true},
		{"invalid port high", testdataPath("invalid_port_high.yaml"), true},
		{"invalid driver", testdataPath("invalid_driver.yaml"), true},
		{"invalid driver empty", testdataPath("invalid_driver_empty.yaml"), true},
		{"invalid dsn empty", testdataPath("invalid_dsn_empty.yaml"), true},
		{"invalid game pick zero", testdataPath("invalid_game_pick_zero.yaml"), true},
		{"invalid game max lt pick", testdataPath("invalid_game_max_lt_pick.yaml"), true},
		{"invalid log level", testdataPath("invalid_log_level.yaml"), true},
		{"invalid log format", testdataPath("invalid_log_format.yaml"), true},
		{"invalid rate limit", testdataPath("invalid_rate_limit.yaml"), true},
		{"invalid rate burst", testdataPath("invalid_rate_burst.yaml"), true},
		{"invalid timeout zero", testdataPath("invalid_timeout_zero.yaml"), true},
		{"invalid draw duration zero", testdataPath("invalid_draw_duration.yaml"), true},

		// Parse error
		{"malformed yaml", testdataPath("malformed.yaml"), true},

		// Edge cases
		{"empty path uses defaults", "", false},
		{"nonexistent file uses defaults", testdataPath("nonexistent.yaml"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load(%q) expected error, got nil", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q) unexpected error: %v", tt.path, err)
			}
			if cfg == nil {
				t.Fatalf("Load(%q) returned nil config", tt.path)
			}
		})
	}
}

func TestLoad_ValidFullValues(t *testing.T) {
	cfg, err := Load(testdataPath("valid_full.yaml"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	// Environment
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}

	// Server
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if got := cfg.Server.ReadTimeout.Duration(); got != 15*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want %v", got, 15*time.Second)
	}
	if got := cfg.Server.WriteTimeout.Duration(); got != 15*time.Second {
		t.Errorf("Server.WriteTimeout = %v, want %v", got, 15*time.Second)
	}
	if got := cfg.Server.ShutdownTimeout.Duration(); got != 5*time.Second {
		t.Errorf("Server.ShutdownTimeout = %v, want %v", got, 5*time.Second)
	}
	if got := cfg.Server.SSEHeartbeat.Duration(); got != 10*time.Second {
		t.Errorf("Server.SSEHeartbeat = %v, want %v", got, 10*time.Second)
	}
	if got := cfg.Server.RequestTimeout.Duration(); got != 20*time.Second {
		t.Errorf("Server.RequestTimeout = %v, want %v", got, 20*time.Second)
	}
	if cfg.Server.RateLimit != 50 {
		t.Errorf("Server.RateLimit = %d, want %d", cfg.Server.RateLimit, 50)
	}
	if cfg.Server.RateBurst != 10 {
		t.Errorf("Server.RateBurst = %d, want %d", cfg.Server.RateBurst, 10)
	}

	// CORS origins
	wantOrigins := []string{"https://example.com", "https://app.example.com"}
	if len(cfg.Server.CORSOrigins) != len(wantOrigins) {
		t.Fatalf("Server.CORSOrigins length = %d, want %d", len(cfg.Server.CORSOrigins), len(wantOrigins))
	}
	for i, got := range cfg.Server.CORSOrigins {
		if got != wantOrigins[i] {
			t.Errorf("Server.CORSOrigins[%d] = %q, want %q", i, got, wantOrigins[i])
		}
	}

	// Game
	if got := cfg.Game.DrawDuration.Duration(); got != 60*time.Second {
		t.Errorf("Game.DrawDuration = %v, want %v", got, 60*time.Second)
	}
	if got := cfg.Game.WaitDuration.Duration(); got != 30*time.Second {
		t.Errorf("Game.WaitDuration = %v, want %v", got, 30*time.Second)
	}
	if cfg.Game.PickCount != 10 {
		t.Errorf("Game.PickCount = %d, want %d", cfg.Game.PickCount, 10)
	}
	if cfg.Game.MaxNumber != 40 {
		t.Errorf("Game.MaxNumber = %d, want %d", cfg.Game.MaxNumber, 40)
	}

	// Database
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Database.DSN != "production.db" {
		t.Errorf("Database.DSN = %q, want %q", cfg.Database.DSN, "production.db")
	}

	// Logging
	if cfg.Logging.Level != "warn" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "warn")
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "json")
	}

	// Discord
	if cfg.Discord.ClientID != "test-client-id" {
		t.Errorf("Discord.ClientID = %q, want %q", cfg.Discord.ClientID, "test-client-id")
	}
	if cfg.Discord.ClientSecret != "test-client-secret" {
		t.Errorf("Discord.ClientSecret = %q, want %q", cfg.Discord.ClientSecret, "test-client-secret")
	}
}

func TestLoad_MinimalUsesDefaults(t *testing.T) {
	cfg, err := Load(testdataPath("valid_minimal.yaml"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	defaults := Default()

	// Environment should be default
	if cfg.Environment != defaults.Environment {
		t.Errorf("Environment = %q, want default %q", cfg.Environment, defaults.Environment)
	}

	// Server should all be defaults
	if cfg.Server.Host != defaults.Server.Host {
		t.Errorf("Server.Host = %q, want default %q", cfg.Server.Host, defaults.Server.Host)
	}
	if cfg.Server.Port != defaults.Server.Port {
		t.Errorf("Server.Port = %d, want default %d", cfg.Server.Port, defaults.Server.Port)
	}
	if cfg.Server.ReadTimeout != defaults.Server.ReadTimeout {
		t.Errorf("Server.ReadTimeout = %v, want default %v", cfg.Server.ReadTimeout.Duration(), defaults.Server.ReadTimeout.Duration())
	}
	if cfg.Server.WriteTimeout != defaults.Server.WriteTimeout {
		t.Errorf("Server.WriteTimeout = %v, want default %v", cfg.Server.WriteTimeout.Duration(), defaults.Server.WriteTimeout.Duration())
	}
	if cfg.Server.RateLimit != defaults.Server.RateLimit {
		t.Errorf("Server.RateLimit = %d, want default %d", cfg.Server.RateLimit, defaults.Server.RateLimit)
	}
	if cfg.Server.RateBurst != defaults.Server.RateBurst {
		t.Errorf("Server.RateBurst = %d, want default %d", cfg.Server.RateBurst, defaults.Server.RateBurst)
	}

	// Game should all be defaults
	if cfg.Game.PickCount != defaults.Game.PickCount {
		t.Errorf("Game.PickCount = %d, want default %d", cfg.Game.PickCount, defaults.Game.PickCount)
	}
	if cfg.Game.MaxNumber != defaults.Game.MaxNumber {
		t.Errorf("Game.MaxNumber = %d, want default %d", cfg.Game.MaxNumber, defaults.Game.MaxNumber)
	}
	if cfg.Game.DrawDuration != defaults.Game.DrawDuration {
		t.Errorf("Game.DrawDuration = %v, want default %v", cfg.Game.DrawDuration.Duration(), defaults.Game.DrawDuration.Duration())
	}

	// Logging should be defaults
	if cfg.Logging.Level != defaults.Logging.Level {
		t.Errorf("Logging.Level = %q, want default %q", cfg.Logging.Level, defaults.Logging.Level)
	}
	if cfg.Logging.Format != defaults.Logging.Format {
		t.Errorf("Logging.Format = %q, want default %q", cfg.Logging.Format, defaults.Logging.Format)
	}

	// Database should be overridden by the minimal file
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Database.DSN != "minimal.db" {
		t.Errorf("Database.DSN = %q, want %q", cfg.Database.DSN, "minimal.db")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	// Should not be nil
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	// Should pass validation
	if err := Validate(cfg); err != nil {
		t.Fatalf("Default() config failed validation: %v", err)
	}

	// Spot-check key defaults
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Game.PickCount != 20 {
		t.Errorf("Game.PickCount = %d, want %d", cfg.Game.PickCount, 20)
	}
	if cfg.Game.MaxNumber != 80 {
		t.Errorf("Game.MaxNumber = %d, want %d", cfg.Game.MaxNumber, 80)
	}
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Database.DSN != "taboo.db" {
		t.Errorf("Database.DSN = %q, want %q", cfg.Database.DSN, "taboo.db")
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "text")
	}
	if cfg.Server.RateLimit != 100 {
		t.Errorf("Server.RateLimit = %d, want %d", cfg.Server.RateLimit, 100)
	}
	if cfg.Server.RateBurst != 20 {
		t.Errorf("Server.RateBurst = %d, want %d", cfg.Server.RateBurst, 20)
	}
}

func TestApplyEnv(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		value  string
		check  func(t *testing.T, cfg *Config)
	}{
		{
			name:   "TABOO_ENVIRONMENT",
			envVar: "TABOO_ENVIRONMENT",
			value:  "production",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Environment != "production" {
					t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
				}
			},
		},
		{
			name:   "TABOO_SERVER_HOST",
			envVar: "TABOO_SERVER_HOST",
			value:  "localhost",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.Host != "localhost" {
					t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "localhost")
				}
			},
		},
		{
			name:   "TABOO_SERVER_PORT",
			envVar: "TABOO_SERVER_PORT",
			value:  "3000",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != 3000 {
					t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 3000)
				}
			},
		},
		{
			name:   "TABOO_SERVER_PORT invalid is ignored",
			envVar: "TABOO_SERVER_PORT",
			value:  "abc",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != Default().Server.Port {
					t.Errorf("Server.Port = %d, want default %d", cfg.Server.Port, Default().Server.Port)
				}
			},
		},
		{
			name:   "TABOO_SERVER_READ_TIMEOUT",
			envVar: "TABOO_SERVER_READ_TIMEOUT",
			value:  "5s",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.ReadTimeout.Duration() != 5*time.Second {
					t.Errorf("Server.ReadTimeout = %v, want %v", cfg.Server.ReadTimeout.Duration(), 5*time.Second)
				}
			},
		},
		{
			name:   "TABOO_SERVER_READ_TIMEOUT invalid is ignored",
			envVar: "TABOO_SERVER_READ_TIMEOUT",
			value:  "not-a-duration",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.ReadTimeout != Default().Server.ReadTimeout {
					t.Errorf("Server.ReadTimeout = %v, want default %v", cfg.Server.ReadTimeout.Duration(), Default().Server.ReadTimeout.Duration())
				}
			},
		},
		{
			name:   "TABOO_SERVER_CORS_ORIGINS",
			envVar: "TABOO_SERVER_CORS_ORIGINS",
			value:  "https://a.com, https://b.com",
			check: func(t *testing.T, cfg *Config) {
				want := []string{"https://a.com", "https://b.com"}
				if len(cfg.Server.CORSOrigins) != len(want) {
					t.Fatalf("CORSOrigins length = %d, want %d", len(cfg.Server.CORSOrigins), len(want))
				}
				for i, got := range cfg.Server.CORSOrigins {
					if got != want[i] {
						t.Errorf("CORSOrigins[%d] = %q, want %q", i, got, want[i])
					}
				}
			},
		},
		{
			name:   "TABOO_SERVER_RATE_LIMIT",
			envVar: "TABOO_SERVER_RATE_LIMIT",
			value:  "200",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.RateLimit != 200 {
					t.Errorf("Server.RateLimit = %d, want %d", cfg.Server.RateLimit, 200)
				}
			},
		},
		{
			name:   "TABOO_SERVER_RATE_BURST",
			envVar: "TABOO_SERVER_RATE_BURST",
			value:  "50",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Server.RateBurst != 50 {
					t.Errorf("Server.RateBurst = %d, want %d", cfg.Server.RateBurst, 50)
				}
			},
		},
		{
			name:   "TABOO_GAME_PICK_COUNT",
			envVar: "TABOO_GAME_PICK_COUNT",
			value:  "15",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Game.PickCount != 15 {
					t.Errorf("Game.PickCount = %d, want %d", cfg.Game.PickCount, 15)
				}
			},
		},
		{
			name:   "TABOO_GAME_MAX_NUMBER",
			envVar: "TABOO_GAME_MAX_NUMBER",
			value:  "60",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Game.MaxNumber != 60 {
					t.Errorf("Game.MaxNumber = %d, want %d", cfg.Game.MaxNumber, 60)
				}
			},
		},
		{
			name:   "TABOO_GAME_DRAW_DURATION",
			envVar: "TABOO_GAME_DRAW_DURATION",
			value:  "2m",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Game.DrawDuration.Duration() != 2*time.Minute {
					t.Errorf("Game.DrawDuration = %v, want %v", cfg.Game.DrawDuration.Duration(), 2*time.Minute)
				}
			},
		},
		{
			name:   "TABOO_DATABASE_DRIVER",
			envVar: "TABOO_DATABASE_DRIVER",
			value:  "sqlite",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Database.Driver != "sqlite" {
					t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
				}
			},
		},
		{
			name:   "TABOO_DATABASE_DSN",
			envVar: "TABOO_DATABASE_DSN",
			value:  "test.db",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Database.DSN != "test.db" {
					t.Errorf("Database.DSN = %q, want %q", cfg.Database.DSN, "test.db")
				}
			},
		},
		{
			name:   "TABOO_LOGGING_LEVEL",
			envVar: "TABOO_LOGGING_LEVEL",
			value:  "debug",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Logging.Level != "debug" {
					t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
				}
			},
		},
		{
			name:   "TABOO_LOGGING_FORMAT",
			envVar: "TABOO_LOGGING_FORMAT",
			value:  "json",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Logging.Format != "json" {
					t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "json")
				}
			},
		},
		{
			name:   "DISCORD_CLIENT_ID",
			envVar: "DISCORD_CLIENT_ID",
			value:  "my-client-id",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Discord.ClientID != "my-client-id" {
					t.Errorf("Discord.ClientID = %q, want %q", cfg.Discord.ClientID, "my-client-id")
				}
			},
		},
		{
			name:   "DISCORD_CLIENT_SECRET",
			envVar: "DISCORD_CLIENT_SECRET",
			value:  "my-secret",
			check: func(t *testing.T, cfg *Config) {
				if cfg.Discord.ClientSecret != "my-secret" {
					t.Errorf("Discord.ClientSecret = %q, want %q", cfg.Discord.ClientSecret, "my-secret")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			t.Setenv(tt.envVar, tt.value)
			applyEnv(cfg)
			tt.check(t, cfg)
		})
	}
}

func TestDuration(t *testing.T) {
	t.Run("UnmarshalYAML", func(t *testing.T) {
		type wrapper struct {
			D Duration `yaml:"d"`
		}

		tests := []struct {
			name    string
			input   string
			want    time.Duration
			wantErr bool
		}{
			{"seconds", "d: 30s", 30 * time.Second, false},
			{"minutes", "d: 5m", 5 * time.Minute, false},
			{"hours", "d: 1h", time.Hour, false},
			{"composite", "d: 1m30s", 90 * time.Second, false},
			{"invalid", "d: notaduration", 0, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var w wrapper
				err := yaml.Unmarshal([]byte(tt.input), &w)
				if tt.wantErr {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got := w.D.Duration(); got != tt.want {
					t.Errorf("Duration() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("MarshalYAML", func(t *testing.T) {
		d := Duration(90 * time.Second)
		got, err := d.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML() error: %v", err)
		}
		if got != "1m30s" {
			t.Errorf("MarshalYAML() = %v, want %q", got, "1m30s")
		}
	})

	t.Run("Duration method", func(t *testing.T) {
		d := Duration(5 * time.Second)
		if got := d.Duration(); got != 5*time.Second {
			t.Errorf("Duration() = %v, want %v", got, 5*time.Second)
		}
	})
}

func TestServerConfig_Addr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default", "0.0.0.0", 8080, "0.0.0.0:8080"},
		{"localhost", "localhost", 3000, "localhost:3000"},
		{"empty host", "", 8080, ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ServerConfig{Host: tt.host, Port: tt.port}
			if got := s.Addr(); got != tt.want {
				t.Errorf("Addr() = %q, want %q", got, tt.want)
			}
		})
	}
}

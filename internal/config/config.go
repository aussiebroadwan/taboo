package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Environment string         `yaml:"environment"` // "development" or "production"
	Server      ServerConfig   `yaml:"server"`
	Game        GameConfig     `yaml:"game"`
	Database    DatabaseConfig `yaml:"database"`
	Logging     LoggingConfig  `yaml:"logging"`
	Discord     DiscordConfig  `yaml:"discord"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host            string   `yaml:"host"`
	Port            int      `yaml:"port"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
	SSEHeartbeat    Duration `yaml:"sse_heartbeat"`
	RequestTimeout  Duration `yaml:"request_timeout"`
	CORSOrigins     []string `yaml:"cors_origins"`
	RateLimit       int      `yaml:"rate_limit"`
	RateBurst       int      `yaml:"rate_burst"`
}

// Addr returns the server address in host:port format.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GameConfig holds game engine configuration.
type GameConfig struct {
	DrawDuration Duration `yaml:"draw_duration"`
	WaitDuration Duration `yaml:"wait_duration"`
	PickCount    int      `yaml:"pick_count"`
	MaxNumber    int      `yaml:"max_number"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// DiscordConfig holds Discord integration configuration.
type DiscordConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// Duration is a wrapper around time.Duration that supports YAML unmarshaling.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(duration)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (any, error) {
	return time.Duration(d).String(), nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Load reads configuration from a YAML file and applies environment overrides.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config file: %w", err)
			}
			// File doesn't exist, use defaults
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file: %w", err)
			}
		}
	}

	// Apply environment variable overrides
	applyEnv(cfg)

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

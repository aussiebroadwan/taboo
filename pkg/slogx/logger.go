package slogx

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New creates a new slog.Logger with the specified options.
func New(opts ...Option) *slog.Logger {
	cfg := &config{
		level:   slog.LevelInfo,
		format:  FormatText,
		output:  os.Stdout,
		service: "",
		version: "",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level: cfg.level,
	}

	switch cfg.format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.output, handlerOpts)
	default:
		handler = slog.NewTextHandler(cfg.output, handlerOpts)
	}

	logger := slog.New(handler)

	if cfg.service != "" {
		logger = logger.With(slog.String("service", cfg.service))
	}
	if cfg.version != "" {
		logger = logger.With(slog.String("version", cfg.version))
	}

	return logger
}

// Format represents the log output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// ParseFormat parses a string into a Format.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	default:
		return FormatText
	}
}

// ParseLevel parses a string into a slog.Level.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type config struct {
	level   slog.Level
	format  Format
	output  io.Writer
	service string
	version string
}

// Option configures a logger.
type Option func(*config)

// WithLevel sets the log level.
func WithLevel(level slog.Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithFormat sets the log format.
func WithFormat(format Format) Option {
	return func(c *config) {
		c.format = format
	}
}

// WithOutput sets the log output writer.
func WithOutput(w io.Writer) Option {
	return func(c *config) {
		c.output = w
	}
}

// WithService adds a service name to all log entries.
func WithService(name string) Option {
	return func(c *config) {
		c.service = name
	}
}

// WithVersion adds a version to all log entries.
func WithVersion(version string) Option {
	return func(c *config) {
		c.version = version
	}
}

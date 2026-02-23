package slogx

import "log/slog"

// Error returns a structured slog.Attr for logging errors consistently.
func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

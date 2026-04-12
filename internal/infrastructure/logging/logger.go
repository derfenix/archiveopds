package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New создаёт slog.Logger с текстовым выводом в stdout.
func New(level string) *slog.Logger {
	lvl := ParseLevel(level)
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	return slog.New(h)
}

// ParseLevel понимает debug, info, warn, warning, error; иначе info.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
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

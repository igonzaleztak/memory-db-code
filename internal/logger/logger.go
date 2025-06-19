package logger

import (
	"log/slog"
	"memorydb/internal/enums"
	"os"
)

// NewLogger creates a new slog.Logger instance with the specified verbose level.
func NewLogger(varboseLevel enums.VerboseLevel) *slog.Logger {
	var level slog.Level
	switch varboseLevel {
	case enums.VerboseLevelDebug:
		level = slog.LevelDebug
	case enums.VerboseLevelInfo:
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo // Default to Info if no valid level is provided
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

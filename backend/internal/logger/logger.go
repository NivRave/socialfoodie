package logger

import (
	"context"
	"log/slog"
	"os"
)

// Setup initializes the global slog JSON handler
func Setup() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

// WithTrace returns a logger with the trace_id attached, if it exists in the context
func WithTrace(ctx context.Context) *slog.Logger {
	traceID, ok := ctx.Value("trace_id").(string)
	if ok && traceID != "" {
		return slog.With("trace_id", traceID)
	}
	return slog.Default()
}

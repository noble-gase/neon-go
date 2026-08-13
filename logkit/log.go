package logkit

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
)

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, slog.GroupAttrs("context", attrs...))
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	// Skip level 1 to get the caller function
	pc, file, line, _ := runtime.Caller(1)
	// Get the function details
	var name string
	if fn := runtime.FuncForPC(pc); fn != nil {
		parts := strings.Split(fn.Name(), "/")
		name = parts[len(parts)-1]
	}

	caller := slog.GroupAttrs(
		"caller",
		slog.String("func", name),
		slog.String("file", file),
		slog.Int("line", line),
	)

	slog.LogAttrs(ctx, slog.LevelWarn, msg, slog.GroupAttrs("context", attrs...), caller)
}

func Error(ctx context.Context, err error, attrs ...slog.Attr) {
	errInfo := "<nil>"
	if err != nil {
		errInfo = err.Error()
	}

	// Skip level 1 to get the caller function
	pc, file, line, _ := runtime.Caller(1)
	// Get the function details
	var name string
	if fn := runtime.FuncForPC(pc); fn != nil {
		parts := strings.Split(fn.Name(), "/")
		name = parts[len(parts)-1]
	}

	caller := slog.GroupAttrs(
		"caller",
		slog.String("func", name),
		slog.String("file", file),
		slog.Int("line", line),
	)

	slog.LogAttrs(ctx, slog.LevelError, errInfo, slog.GroupAttrs("context", attrs...), caller)
}

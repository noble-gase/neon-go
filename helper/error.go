package helper

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"strings"

	"github.com/noble-gase/neon/codekit"
)

type NilError string

func (e NilError) Error() string { return string(e) }

// Error logs the error with caller, then returns the codes.Err
func Error(ctx context.Context, err error, attrs ...slog.Attr) error {
	if err == nil {
		return nil
	}

	var code codekit.Code
	if errors.As(err, &code) {
		return code
	}

	// Skip level 1 to get the caller function
	pc, file, line, _ := runtime.Caller(1)
	// Get the function details
	var name string
	if fn := runtime.FuncForPC(pc); fn != nil {
		parts := strings.Split(fn.Name(), "/")
		name = parts[len(parts)-1]
	}

	caller := slog.Attr{
		Key: "caller",
		Value: slog.GroupValue(
			slog.String("func", name),
			slog.String("file", file),
			slog.Int("line", line),
		),
	}

	slog.LogAttrs(ctx, slog.LevelError, err.Error(), slog.Attr{
		Key:   "context",
		Value: slog.GroupValue(attrs...),
	}, caller)

	return codekit.Err
}

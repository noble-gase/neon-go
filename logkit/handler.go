package logkit

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/noble-gase/neon/closekit"
	"github.com/noble-gase/neon/helper"
)

var hostname, _ = os.Hostname()

type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String("hostname", hostname))
	// traceId
	if traceId := helper.MDTraceIDFromCtx(ctx); traceId != "" {
		r.AddAttrs(slog.String("trace_id", traceId))
	}
	return h.Handler.Handle(ctx, r)
}

func NewContextHandler(w io.WriteCloser, opts *slog.HandlerOptions) slog.Handler {
	closekit.Add("log", closekit.P100, func() error {
		return w.Close()
	})
	return &contextHandler{
		Handler: slog.NewJSONHandler(w, opts),
	}
}

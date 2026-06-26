package middleware

import (
	"context"
	"log/slog"
)

type contextSlogHandler struct {
	handler slog.Handler
}

func NewContextSlogHandler(h slog.Handler) slog.Handler {
	return &contextSlogHandler{handler: h}
}

func (h *contextSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *contextSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestIDFromCtx(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.handler.Handle(ctx, r)
}

func (h *contextSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextSlogHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *contextSlogHandler) WithGroup(name string) slog.Handler {
	return &contextSlogHandler{handler: h.handler.WithGroup(name)}
}

package slogan

import (
	"context"
	"log/slog"
)

type (
	commonHandler struct {
		h           slog.Handler
		ContextKeys []ContextKeyGetter
	}
)

func (ch *commonHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if ctxLvl := ContextLogLvl(ctx); ctxLvl != nil {
		return level >= ctxLvl.Level()
	}

	return ch.h.Enabled(ctx, level)
}

func (ch *commonHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, len(ch.ContextKeys))

	for _, ck := range ch.ContextKeys {
		value := ctx.Value(ck)
		if value == nil {
			continue
		}

		attrs = append(attrs, Any(ck.GetKey(), value))
	}

	record.AddAttrs(attrs...)

	return ch.h.Handle(ctx, record)
}

func (ch *commonHandler) WithAttrs(attrs []slog.Attr) *commonHandler {
	return &commonHandler{
		h:           ch.h.WithAttrs(attrs),
		ContextKeys: ch.ContextKeys,
	}
}

func (ch *commonHandler) WithGroup(name string) *commonHandler {
	return &commonHandler{
		h:           ch.h.WithGroup(name),
		ContextKeys: ch.ContextKeys,
	}
}

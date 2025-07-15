package slogan

import (
	"context"
	"io"
	"log/slog"
)

type (
	JSONHandler struct {
		*commonHandler
	}
)

func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions, contextKeys ...ContextKeyGetter) slog.Handler {
	return &JSONHandler{
		commonHandler: &commonHandler{
			h:           slog.NewJSONHandler(w, opts),
			ContextKeys: contextKeys,
		},
	}
}

func (jh *JSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return jh.commonHandler.Enabled(ctx, level)
}

func (jh *JSONHandler) Handle(ctx context.Context, record slog.Record) error {
	return jh.commonHandler.Handle(ctx, record)
}

func (jh *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &JSONHandler{
		commonHandler: jh.commonHandler.WithAttrs(attrs),
	}
}

func (jh *JSONHandler) WithGroup(name string) slog.Handler {
	return &JSONHandler{
		commonHandler: jh.commonHandler.WithGroup(name),
	}
}

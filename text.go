package slogan

import (
	"context"
	"io"
	"log/slog"
)

type (
	TextHandler struct {
		*commonHandler
	}
)

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions, contextKeys ...ContextKeyGetter) slog.Handler {
	return &TextHandler{
		commonHandler: &commonHandler{
			h:           slog.NewTextHandler(w, opts),
			ContextKeys: contextKeys,
		},
	}
}

func (th *TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return th.commonHandler.Enabled(ctx, level)
}

func (th *TextHandler) Handle(ctx context.Context, record slog.Record) error {
	return th.commonHandler.Handle(ctx, record)
}

func (th *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TextHandler{
		commonHandler: th.commonHandler.WithAttrs(attrs),
	}
}

func (th *TextHandler) WithGroup(name string) slog.Handler {
	return &TextHandler{
		commonHandler: th.commonHandler.WithGroup(name),
	}
}

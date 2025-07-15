package slogan

import (
	"context"
	"log/slog"
)

type (
	ContextKeyGetter interface {
		GetKey() string
	}

	// ctxLogLvlKey - context key for storing log level value for a specific context
	ctxLogLvlKey struct{}
)

func DebugContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxLogLvlKey{}, slog.LevelDebug)
}

func InfoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxLogLvlKey{}, slog.LevelInfo)
}

func WarnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxLogLvlKey{}, slog.LevelWarn)
}

func ErrorContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxLogLvlKey{}, slog.LevelError)
}

func LvlContext(ctx context.Context, lvl slog.Leveler) context.Context {
	return context.WithValue(ctx, ctxLogLvlKey{}, lvl.Level())
}

func ContextLogLvl(ctx context.Context) slog.Leveler {
	lvl := ctx.Value(ctxLogLvlKey{})
	if lvl == nil {
		return nil
	}

	return lvl.(slog.Leveler)
}

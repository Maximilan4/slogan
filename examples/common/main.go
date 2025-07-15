package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Maximilan4/slogan"
)

// ReqIdCKey - example type for storing value in ctx
type ReqIdCKey struct{}

// handler will call GetKey and use string value as an atribute key
func (r ReqIdCKey) GetKey() string {
	return "reqid"
}

func main() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		// DefaultReplaceFunc - can disable time and rewrite source key value
		ReplaceAttr: slogan.DefaultReplaceFunc(true, true),
	}

	// third argument - context keys, must implement ContextKeyGetter
	handler := slogan.NewTextHandler(os.Stdout, &opts, ReqIdCKey{})
	logger := slog.New(handler)

	logger.Info("first log")
	// create a new context with value and log it
	ctx := context.WithValue(context.Background(), ReqIdCKey{}, "example")
	logger.InfoContext(ctx, "log with context value")

	// debug skipping via global handler settings
	logger.DebugContext(ctx, "you will not see this")

	// decrease log level for a specific call, wrap ctx above with debug value
	debugCtx := slogan.DebugContext(ctx)
	logger.DebugContext(debugCtx, "now you see me")

	// increase log level for a specific call
	warnCtx := slogan.WarnContext(ctx)
	logger.WarnContext(warnCtx, "filter info and debug calls")
	logger.InfoContext(warnCtx, "skip info")
	logger.DebugContext(warnCtx, "skip debug")
}

package main

import (
	"log/slog"
	"os"

	"github.com/Maximilan4/slogan"
)

func main() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		// DefaultReplaceFunc - can disable time and rewrite source key value
		ReplaceAttr: slogan.DefaultReplaceFunc(true, true),
	}

	// third argument - context keys, must implement ContextKeyGetter
	handler := slogan.NewTextHandler(os.Stdout, &opts)
	logger := slog.New(handler)

	exampleValue := "example string value"
	logger.Info("log a pointer",
		slog.Any("slog", &exampleValue),         // produce a pointer value like slog=0x1400010af30
		slogan.Pointer("slogan", &exampleValue)) // will dereference p value

	mValue := map[string]any{
		"strkey": "this is a string",
		"intkey": 123,
		"nested": map[int]int{
			1: 2,
			3: 4,
		},
	}

	logger.Info("log a map value",
		slog.Any("slog", mValue),     // default short format output map[intkey:123 nested:map[1:2 3:4] strkey:this is a string]
		slogan.Map("slogan", mValue), // will produce group with name sloganmap, nested maps are supported
	)

	parts := []string{"my", "split", "string"}
	logger.Info("log a slice value",
		slog.Any("slog", parts),                         // default output slog="[my split string]"
		slogan.Slice("slogan", parts, false),            // will produce group with indexes keys
		slogan.Slice("unwrap", []string{"value"}, true), // will unwrap slice single value to a key like unwrap=value
	)
}

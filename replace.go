package slogan

import (
	"fmt"
	"log/slog"
)

// DefaultReplaceFunc - func which operating with default slog features
// disableTime - omits time, if set to true
// enableSource - true adds to log func name and line, false - omits it
func DefaultReplaceFunc(disableTime, enableSource bool) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			if disableTime {
				return slog.Attr{}
			}
		case slog.SourceKey:
			if !enableSource {
				return slog.Attr{}
			}

			source, ok := a.Value.Any().(*slog.Source)
			if !ok {
				return slog.Attr{}
			}

			// rewrite source file by function name -> this is more informative
			return slog.Attr{
				Key:   slog.SourceKey,
				Value: slog.StringValue(fmt.Sprintf("%s:%d", source.Function, source.Line)),
			}
		}

		return a
	}
}

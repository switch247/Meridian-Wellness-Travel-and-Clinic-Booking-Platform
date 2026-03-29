package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	return NewWithWriter(os.Stdout)
}

func NewWithWriter(w io.Writer) *Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case "authorization":
				return slog.String(a.Key, "REDACTED")
			case "phone", "address":
				return slog.String(a.Key, RedactSensitive(a.Value.String()))
			default:
				return a
			}
		},
	})
	return &Logger{Logger: slog.New(handler)}
}

func RedactSensitive(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	// mask everything except last 4 characters
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}

func StubWriter(w io.Writer) *Logger {
	return NewWithWriter(w)
}

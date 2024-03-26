package pkg

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// skip timestamp
				return slog.Attr{}
				// // return timestamp in milliseconds
				// return slog.Int64("timestamp", a.Value.Time().UnixMilli())
			}
			return a
		},
	}))
}

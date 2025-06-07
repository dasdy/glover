package keylog

import (
	"errors"
	"log/slog"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/keylog/parser"
)

func Loop(ch <-chan string, storage db.Storage, trackers []db.Tracker, enableLogs bool) {
	for line := range ch {
		parsed, err := parser.ParseLine(line)
		if err != nil && !errors.Is(err, parser.ErrEmptyLine) {
			slog.Error("Failed to parse line", "error", err, "line", line)
		}

		if parsed != nil {
			if enableLogs {
				slog.Info("Got keypress", "keypress", parsed)
			}

			if storage.Store(parsed) != nil {
				slog.Error("Failed to log item", "error", err)
			}

			for _, tracker := range trackers {
				tracker.HandleKeyNow(parsed.Position, parsed.Pressed, enableLogs)
			}
		}
	}

	slog.Info("Channel closed; bailing out")
}

package main

import (
	"log/slog"
	"os"

	"github.com/dasdy/glover/cmd/glover"
)

func main() {
	// This does not work: See the SetDefault() comments, but it causes a deadlock, since we'll try locking some internal mutex in slog lib twice.
	// handler := logging.ContextHandler{Handler: slog.Default().Handler()}

	// handler := logging.ContextHandler{Handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 	AddSource: true,
	// })}
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
		// slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}),
		// logging.ContextHandler{Handler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})},
	)

	slog.SetDefault(logger)
	glover.Execute()
}

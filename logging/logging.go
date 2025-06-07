package logging

import (
	"context"
	"fmt"
	"log/slog"
)

type ctxKey string

const (
	slogFields  ctxKey = "slog_fields"
	PackageName string = "package"
)

type ContextHandler struct {
	slog.Handler
}

// Handle adds contextual attributes to the Record before calling the underlying handler.
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	err := h.Handler.Handle(ctx, r)
	if err != nil {
		return fmt.Errorf("error handling record for a log: %+v: %w", r, err)
	}

	return nil
}

// AppendCtx adds an slog attribute to the provided context so that it will be included in any Record created with such context.
func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attr)

		return context.WithValue(parent, slogFields, v)
	}

	v := []slog.Attr{attr}

	return context.WithValue(parent, slogFields, v)
}

func PackageCtx(packageName string) context.Context {
	return AppendCtx(context.Background(), slog.String(PackageName, packageName))
}

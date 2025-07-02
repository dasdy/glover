package routes

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/model"
)

// ServerHandler holds all dependencies needed for the web server handlers.
type ServerHandler struct {
	Storage         db.Storage
	KeyNames        []string
	ComboTracker    db.Tracker
	NeighborTracker db.Tracker
	LocationsOnGrid *model.KeyboardLayout
}

// SafeRenderTemplate safely renders a templ component to an http.ResponseWriter.
func SafeRenderTemplate(component templ.Component, w http.ResponseWriter) error {
	// Do not write to w because it implies 200 status
	var buf bytes.Buffer

	err := component.Render(context.Background(), &buf)
	if err != nil {
		return fmt.Errorf("could not render template: %w", err)
	}

	// Template executed successfully to the buffer.
	// Now, copy it over to the ResponseWriter
	// This implies a 200 OK status code
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("Failed to write response", "error", err)

		return fmt.Errorf("could not write to response writer: %w", err)
	}

	return nil
}

// initEmptyMap initializes a map with empty key events for all keys in the layout.
func InitEmptyMap(names []string, locationsOnGrid map[model.KeyPosition]model.Location) map[model.RowCol]*model.MinimalKeyEventWithLabel {
	// put empty items in the map so that we show them properly later
	groupedItems := make(map[model.RowCol]*model.MinimalKeyEventWithLabel)

	for pos, key := range locationsOnGrid {
		name := "<OOB>"
		if int(pos) < len(names) {
			name = names[pos]
		}

		groupedItems[model.RowCol{Row: key.Row, Col: key.Col}] = &model.MinimalKeyEventWithLabel{Count: 0, Position: pos, KeyLabel: name, Location: key}
	}

	return groupedItems
}

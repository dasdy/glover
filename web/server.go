package web

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/a-h/templ"
	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/layout"
	"github.com/dasdy/glover/model"
	cs "github.com/dasdy/glover/web/components"
)

type ServerHandler struct {
	Storage         db.Storage
	KeyNames        []string
	ComboTracker    db.Tracker
	NeighborTracker db.Tracker
	LocationsOnGrid *model.KeyboardLayout
}

func SafeRenderTemplate(component templ.Component, w http.ResponseWriter) error {
	// Do not write to w because it implies 200 status
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

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

func (s *ServerHandler) BuildStatsRenderContext(dbStats []model.MinimalKeyEvent) cs.RenderContext {
	// TODO: this init leads to reading layout file on every request. Need to get rid of this somehow
	groupedItems := initEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)

	maxVal := 0
	// set non-zero items in the map
	for _, key := range dbStats {
		loc, ok := s.LocationsOnGrid.Locations[key.Position]
		if !ok {
			slog.Error("Position not found in layout", "position", key.Position)
		}

		if maxVal < key.Count {
			maxVal = key.Count
		}

		groupedItems[model.RowCol{Row: loc.Row, Col: loc.Col}].Count += key.Count
	}

	// Iterate over total grid and add items that exit in the layout.
	items := make([]cs.Item, 0, len(groupedItems))

	for _, item := range groupedItems {
		locationOnGrid := item.Location
		items = append(items, cs.Item{
			Position:       item.Position,
			KeypressAmount: strconv.Itoa(item.Count),
			KeyName:        item.KeyLabel,
			Location:       locationOnGrid,
		})
	}

	return cs.RenderContext{TotalCols: s.LocationsOnGrid.Cols, TotalRows: s.LocationsOnGrid.Rows, Items: items, MaxVal: maxVal, Page: cs.PageTypeStats}
}

func (s *ServerHandler) StatsHandle(w http.ResponseWriter, _ *http.Request) {
	slog.Info("Handling stats page request")

	curStats, err := s.Storage.GatherAll()
	if err != nil {
		slog.Error("Failed to get stats", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	slog.Debug("Gathered current stats")

	renderContext := s.BuildStatsRenderContext(curStats)

	slog.Debug("Built render context")

	_ = SafeRenderTemplate(cs.HeatMap(&renderContext), w)
}

func initEmptyMap(names []string, locationsOnGrid map[model.KeyPosition]model.Location) map[model.RowCol]*model.MinimalKeyEventWithLabel {
	// put empty items in the map so that we show them later properly
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

func (s *ServerHandler) BuildCombosRenderContext(combos []model.Combo, position model.KeyPosition) cs.RenderContext {
	slog.Debug("Building combos context", "comboCount", len(combos))

	// Sort combos by press count to get top 5
	slices.SortFunc(combos, func(a, b model.Combo) int {
		return -cmp.Compare(a.Pressed, b.Pressed) // Negative to sort in descending order
	})

	// Keep only top 5 combos
	// if len(combosToDisplay) > 5 {
	// 	combosToDisplay = combosToDisplay[:5]
	// }

	groupedItems := initEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)
	maxVal := 0

	// set non-zero items in the map
	for _, combo := range combos {
		for _, key := range combo.Keys {
			loc, ok := s.LocationsOnGrid.Locations[key]
			if !ok {
				slog.Error("Position not found in layout", "position", key)
			}

			groupedItems[model.RowCol{Row: loc.Row, Col: loc.Col}].Count += combo.Pressed
		}

		if maxVal < combo.Pressed {
			maxVal = combo.Pressed
		}
	}

	// Iterate over total grid and add real and hidden items.
	items := make([]cs.Item, 0)
	l := model.RowCol{Row: 0, Col: 0}

	for i := 0; i <= s.LocationsOnGrid.Rows; i++ {
		for j := 0; j <= s.LocationsOnGrid.Cols; j++ {
			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				locationOnGrid := item.Location
				highlight := item.Position == position
				items = append(items, cs.Item{
					Position:       item.Position,
					KeypressAmount: strconv.Itoa(item.Count),
					KeyName:        item.KeyLabel,
					Highlight:      highlight,
					Location:       locationOnGrid,
				})
			}
		}
	}

	// Create combo connections for the top combos
	connections := make([]cs.ComboConnection, 0, 5)

	for _, combo := range combos {
		var otherPos model.KeyPosition

		for _, key := range combo.Keys {
			if key != position {
				otherPos = key

				break
			}
		}

		connections = append(connections, cs.ComboConnection{
			FromPosition: position,
			ToPosition:   otherPos,
			PressCount:   combo.Pressed,
		})
		if len(connections) >= 5 {
			break
		}
	}

	slog.Debug("Found combo connections", "count", len(connections))

	return cs.RenderContext{
		TotalCols:         s.LocationsOnGrid.Cols,
		TotalRows:         s.LocationsOnGrid.Rows,
		Items:             items,
		MaxVal:            maxVal,
		HighlightPosition: position,
		ComboConnections:  connections,
		Page:              cs.PageTypeCombo,
	}
}

func (s *ServerHandler) CombosHandle(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling combos page request")

	positionString := r.URL.Query().Get("position")

	position, err := strconv.ParseInt(positionString, 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	positionCasted := model.KeyPosition(position)
	combos := s.ComboTracker.GatherCombos(positionCasted)

	renderContext := s.BuildCombosRenderContext(combos, positionCasted)
	_ = SafeRenderTemplate(cs.HeatMap(&renderContext), w)
}

func (s *ServerHandler) BuildNeighborsRenderContext(neighbors []model.Combo, position model.KeyPosition) cs.RenderContext {
	groupedItems := initEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)
	maxVal := 0

	// set non-zero items in the map
	for _, combo := range neighbors {
		neighborPosition := position

		for _, key := range combo.Keys {
			if key != position {
				neighborPosition = key

				break
			}
		}

		loc, ok := s.LocationsOnGrid.Locations[neighborPosition]
		if !ok {
			slog.Error("Position not found in layout", "position", neighborPosition)

			continue
		}

		groupedItems[model.RowCol{Row: loc.Row, Col: loc.Col}].Count += combo.Pressed

		if maxVal < combo.Pressed {
			maxVal = combo.Pressed
		}
	}

	items := make([]cs.Item, 0)
	l := model.RowCol{Row: 0, Col: 0}

	for i := 0; i <= s.LocationsOnGrid.Rows; i++ {
		for j := 0; j <= s.LocationsOnGrid.Cols; j++ {
			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				locationOnGrid := item.Location
				highlight := item.Position == position
				items = append(items, cs.Item{
					Position:       item.Position,
					KeypressAmount: strconv.Itoa(item.Count),
					KeyName:        item.KeyLabel,
					Highlight:      highlight,
					Location:       locationOnGrid,
				})
			}
		}
	}

	// Create combo connections for the top combos
	connections := make([]cs.ComboConnection, 0, 5)

	// Sort combos by press count to get top 5
	slices.SortFunc(neighbors, func(a, b model.Combo) int {
		return -cmp.Compare(a.Pressed, b.Pressed) // Negative to sort in descending order
	})

	for _, combo := range neighbors {
		var otherPos model.KeyPosition

		for _, key := range combo.Keys {
			if key != position {
				otherPos = key

				break
			}
		}

		connections = append(connections, cs.ComboConnection{
			FromPosition: position,
			ToPosition:   otherPos,
			PressCount:   combo.Pressed,
		})
		if len(connections) >= 5 {
			break
		}
	}

	slog.Debug("Found neighbor connections", "count", len(connections))

	return cs.RenderContext{
		TotalCols:         s.LocationsOnGrid.Cols,
		TotalRows:         s.LocationsOnGrid.Rows,
		Items:             items,
		MaxVal:            maxVal,
		HighlightPosition: position,
		ComboConnections:  connections,
		Page:              cs.PageTypeNeighbors,
	}
}

func (s *ServerHandler) NeighborsHandle(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling neighbors page request")

	positionString := r.URL.Query().Get("position")

	position, err := strconv.ParseInt(positionString, 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	positionCasted := model.KeyPosition(position)
	neighbors := s.NeighborTracker.GatherCombos(positionCasted)

	renderContext := s.BuildNeighborsRenderContext(neighbors, positionCasted)
	_ = SafeRenderTemplate(cs.HeatMap(&renderContext), w)
}

func disableCacheInDevMode(dev bool, next http.Handler) http.Handler {
	if !dev {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func loadLocationsOnGrid(infoJSONFile string) (*model.KeyboardLayout, error) {
	reader, err := layout.OpenPath(infoJSONFile)
	if err != nil {
		return nil, fmt.Errorf("could not open layout file %s. %w", infoJSONFile, err)
	}

	defer reader.Close()

	locationsParsed, err := layout.LoadZmkLocationsJSON(reader)
	if err != nil {
		return nil, fmt.Errorf("could not parse info.json: %w", err)
	}

	return locationsParsed, nil
}

func BuildServer(storage db.Storage, comboTracker db.Tracker, neighborTracker db.Tracker, keymapFile string, infoFilePath string, dev bool) *http.ServeMux {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/",
		disableCacheInDevMode(dev,
			http.StripPrefix("/assets",
				http.FileServer(http.Dir("assets")))))

	slog.Info("Parsing keyboard layout", "file", infoFilePath)

	locationsParsed, err := loadLocationsOnGrid(infoFilePath)
	if err != nil {
		slog.Error("Failed to parse keyboard layout", "error", err, "file", infoFilePath)
		log.Fatal(err)
	}

	keyNames, err := layout.GetKeyLabels(keymapFile)
	if err != nil {
		slog.Error("Failed to parse keymap file", "error", err, "file", keymapFile)
	}

	slog.Info("Successfully parsed keyboard layout",
		"locations", len(locationsParsed.Locations),
		"rows", locationsParsed.Rows,
		"cols", locationsParsed.Cols)

	handler := ServerHandler{
		Storage:         storage,
		KeyNames:        keyNames,
		ComboTracker:    comboTracker,
		NeighborTracker: neighborTracker,
		LocationsOnGrid: locationsParsed,
	}
	mux.Handle("/combo", http.HandlerFunc(handler.CombosHandle))
	mux.Handle("/neighbors", http.HandlerFunc(handler.NeighborsHandle))
	mux.Handle("/", http.HandlerFunc(handler.StatsHandle))

	return mux
}

func StartServer(port int, storage db.Storage, comboTracker db.Tracker, neighborTracker db.Tracker, keymapFile string, infoFilePath string, dev bool) {
	slog.Info("Starting server", "port", port)

	err := http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		BuildServer(storage, comboTracker, neighborTracker, keymapFile, infoFilePath, dev))
	if err != nil {
		slog.Error("Server failed to start", "error", err)
		log.Fatal(err)
	}
}

package routes

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dasdy/glover/model"
	cs "github.com/dasdy/glover/web/components"
)

// BuildStatsRenderContext builds the render context for the stats page.
func (s *ServerHandler) BuildStatsRenderContext(dbStats []model.MinimalKeyEvent) cs.RenderContext {
	// TODO: this init leads to reading layout file on every request. Need to get rid of this somehow
	groupedItems := InitEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)

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

// StatsHandle handles requests to the stats page.
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

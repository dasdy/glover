package routes

import (
	"cmp"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/dasdy/glover/model"
	cs "github.com/dasdy/glover/web/components"
)

// BuildNeighborsRenderContext builds the render context for the neighbors page.
func (s *ServerHandler) BuildNeighborsRenderContext(neighbors []model.Combo, position model.KeyPosition) cs.RenderContext {
	groupedItems := InitEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)
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

// NeighborsHandle handles requests to the neighbors page.
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

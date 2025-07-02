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

// BuildCombosRenderContext builds the render context for the combos page.
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

	groupedItems := InitEmptyMap(s.KeyNames, s.LocationsOnGrid.Locations)
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

// CombosHandle handles requests to the combos page.
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

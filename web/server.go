package web

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"

	"github.com/a-h/templ"
	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/layout"
	"github.com/dasdy/glover/model"
	cs "github.com/dasdy/glover/web/components"
)

type ServerHandler struct {
	Storage    db.Storage
	KeymapFile string
}

func GetKeyLabels(filename string) ([]string, error) {
	// TODO: parameterize;
	//nolint:dogsled
	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	fp := filepath.Join(filepath.Dir(b), "..")

	file, err := os.Open(filepath.Join(fp, "data", filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keymap, err := layout.Parse(file)
	if err != nil {
		return nil, err
	}

	if len(keymap.Layers) < 1 {
		return nil, errors.New("expected at least 1 layer in layout")
	}

	results := make([]string, 0, len(keymap.Layers[0].Bindings))

	for _, b := range keymap.Layers[0].Bindings {
		switch b.Action {
		case "&kp":
			for i := range b.Modifiers {
				if v, ok := labels[b.Modifiers[i]]; ok {
					b.Modifiers[i] = v
				}
			}

			if len(b.Modifiers) > 1 {
				results = append(results, fmt.Sprintf("%+v", b.Modifiers))
			} else {
				results = append(results, b.Modifiers[0])
			}
		case "&magic":
			results = append(results, "🪄")
		default:
			results = append(results, fmt.Sprintf("%s %+v", b.Action, b.Modifiers))
		}
	}

	return results, nil
}

func SafeRenderTemplate(component templ.Component, w http.ResponseWriter) error {
	// Do not write to w because it implies 200 status
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		log.Printf("Could not render: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return err
	}

	// Template executed successfully to the buffer.
	// Now, copy it over to the ResponseWriter
	// This implies a 200 OK status code
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Could not render: %s", err.Error())

		return err
	}

	return nil
}

func (s *ServerHandler) BuildStatsRenderContext(dbStats []model.MinimalKeyEvent) cs.RenderContext {
	groupedItems, maxVal, totalCols, totalRows := initEmptyMap(s.KeymapFile)

	// set non-zero items in the map
	for _, key := range dbStats {
		loc, ok := locationsOnGrid[key.Position]
		if !ok {
			log.Printf("Could not find position %d, wtf", key.Position)
		}

		if maxVal < key.Count {
			maxVal = key.Count
		}

		groupedItems[cs.Location{Row: loc.Row, Col: loc.Col}].Count += key.Count
	}

	// Iterate over total grid and add real and hidden items.
	// TODO: can this be done without a bunch of hidden items?
	items := make([]cs.Item, 0)
	l := cs.Location{Row: 0, Col: 0}

	for i := 0; i <= totalRows; i++ {
		for j := 0; j <= totalCols; j++ {
			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				items = append(items, cs.Item{
					Position:       item.Position,
					Row:            i,
					Col:            j,
					KeypressAmount: strconv.Itoa(item.Count),
					KeyName:        item.KeyLabel,
				})
			}
		}
	}

	return cs.RenderContext{TotalCols: 18, Items: items, MaxVal: maxVal, Page: cs.PageTypeStats}
}

func (s *ServerHandler) StatsHandle(w http.ResponseWriter, _ *http.Request) {
	log.Print("Got request to stats page")

	curStats, err := s.Storage.GatherAll()
	if err != nil {
		log.Printf("Could not get stats: %s", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	renderContext := s.BuildStatsRenderContext(curStats)
	_ = SafeRenderTemplate(cs.HeatMap(&renderContext), w)
}

func initEmptyMap(name string) (map[cs.Location]*model.MinimalKeyEventWithLabel, int, int, int) {
	totalRows := 0
	totalCols := 0
	maxVal := 0

	names, _ := GetKeyLabels(name)
	// put empty items in the map so that we show them later properly
	groupedItems := make(map[cs.Location]*model.MinimalKeyEventWithLabel)

	for pos, key := range locationsOnGrid {
		name := "<OOB>"
		if pos < len(names) {
			name = names[pos]
		}

		groupedItems[key] = &model.MinimalKeyEventWithLabel{Row: key.Row, Col: key.Col, Count: 0, Position: pos, KeyLabel: name}

		if key.Row > totalRows {
			totalRows = key.Row
		}

		if key.Col > totalCols {
			totalCols = key.Col
		}
	}

	return groupedItems, maxVal, totalCols, totalRows
}

func (s *ServerHandler) BuildCombosRenderContext(combos []model.Combo, position int64) cs.RenderContext {
	combosToDisplay := make([]model.Combo, 0)

	for _, c := range combos {
		if len(c.Keys) > 2 {
			continue
		}

		for _, k := range c.Keys {
			if int64(k.Position) == position {
				combosToDisplay = append(combosToDisplay, c)
			}
		}
	}

	log.Printf("Found combos: %d", len(combosToDisplay))

	// Sort combos by press count to get top 5
	slices.SortFunc(combosToDisplay, func(a, b model.Combo) int {
		return -cmp.Compare(a.Pressed, b.Pressed) // Negative to sort in descending order
	})

	// Keep only top 5 combos
	// if len(combosToDisplay) > 5 {
	// 	combosToDisplay = combosToDisplay[:5]
	// }

	groupedItems, maxVal, totalCols, totalRows := initEmptyMap(s.KeymapFile)

	// set non-zero items in the map
	for _, combo := range combosToDisplay {
		for _, key := range combo.Keys {
			loc, ok := locationsOnGrid[key.Position]
			if !ok {
				log.Printf("Could not find position %d, wtf", key.Position)
			}

			groupedItems[cs.Location{Row: loc.Row, Col: loc.Col}].Count += combo.Pressed
		}

		if maxVal < combo.Pressed {
			maxVal = combo.Pressed
		}
	}

	// Iterate over total grid and add real and hidden items.
	items := make([]cs.Item, 0)
	l := cs.Location{Row: 0, Col: 0}

	for i := 0; i <= totalRows; i++ {
		for j := 0; j <= totalCols; j++ {
			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				highlight := int64(item.Position) == position
				items = append(items, cs.Item{
					Position:       item.Position,
					Row:            item.Row,
					Col:            item.Col,
					KeypressAmount: strconv.Itoa(item.Count),
					KeyName:        item.KeyLabel,
					Highlight:      highlight,
				})
			}
		}
	}

	// Create combo connections for the top combos
	connections := make([]cs.ComboConnection, 0, 5)

	for _, combo := range combosToDisplay {
		var otherPos int

		for _, key := range combo.Keys {
			if int64(key.Position) != position {
				otherPos = key.Position

				break
			}
		}
		log.Printf("pressCount: %d", combo.Pressed)

		connections = append(connections, cs.ComboConnection{
			FromPosition: int(position),
			ToPosition:   otherPos,
			PressCount:   combo.Pressed,
		})
		if len(connections) >= 5 {
			break
		}
	}

	log.Printf("Found connections: %d", len(connections))

	return cs.RenderContext{
		TotalCols:         18,
		Items:             items,
		MaxVal:            maxVal,
		HighlightPosition: int(position),
		ComboConnections:  connections,
		Page:              cs.PageTypeCombo,
	}
}

func (s *ServerHandler) CombosHandle(w http.ResponseWriter, r *http.Request) {
	log.Print("Got request to combos page")

	combos := s.Storage.GatherCombos()
	positionString := r.URL.Query().Get("position")

	position, err := strconv.ParseInt(positionString, 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	renderContext := s.BuildCombosRenderContext(combos, position)
	_ = SafeRenderTemplate(cs.HeatMap(&renderContext), w)
}

func (s *ServerHandler) BuildNeighborsRenderContext(neighbors []model.Combo, position int64) cs.RenderContext {
	groupedItems, maxVal, totalCols, totalRows := initEmptyMap(s.KeymapFile)

	// set non-zero items in the map
	for _, combo := range neighbors {
		neighborPosition := int(position)

		for _, key := range combo.Keys {
			if int64(key.Position) != position {
				neighborPosition = key.Position

				break
			}
		}

		loc, ok := locationsOnGrid[neighborPosition]
		if !ok {
			log.Printf("Could not find position %d, wtf", neighborPosition)

			continue
		}

		groupedItems[cs.Location{Row: loc.Row, Col: loc.Col}].Count += combo.Pressed

		if maxVal < combo.Pressed {
			maxVal = combo.Pressed
		}
	}

	items := make([]cs.Item, 0)
	l := cs.Location{Row: 0, Col: 0}

	for i := 0; i <= totalRows; i++ {
		for j := 0; j <= totalCols; j++ {
			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				highlight := int64(item.Position) == position
				items = append(items, cs.Item{
					Position:       item.Position,
					Row:            item.Row,
					Col:            item.Col,
					KeypressAmount: strconv.Itoa(item.Count),
					KeyName:        item.KeyLabel,
					Highlight:      highlight,
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
		var otherPos int

		for _, key := range combo.Keys {
			if int64(key.Position) != position {
				otherPos = key.Position

				break
			}
		}

		log.Printf("pressCount: %d", combo.Pressed)

		connections = append(connections, cs.ComboConnection{
			FromPosition: int(position),
			ToPosition:   otherPos,
			PressCount:   combo.Pressed,
		})
		if len(connections) >= 5 {
			break
		}
	}

	log.Printf("Found connections: %d", len(connections))

	return cs.RenderContext{
		TotalCols:         18,
		Items:             items,
		MaxVal:            maxVal,
		HighlightPosition: int(position),
		ComboConnections:  connections,
		Page:              cs.PageTypeNeighbors,
	}
}

func (s *ServerHandler) NeighborsHandle(w http.ResponseWriter, r *http.Request) {
	log.Print("Got request to neighbors page")

	positionString := r.URL.Query().Get("position")

	position, err := strconv.ParseInt(positionString, 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	neighbors, err := s.Storage.GatherNeighbors(int(position))
	if err != nil {
		log.Printf("Could not get neighbors: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	renderContext := s.BuildNeighborsRenderContext(neighbors, position)
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

func BuildServer(storage db.Storage, keymapFile string, dev bool) *http.ServeMux {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/",
		disableCacheInDevMode(dev,
			http.StripPrefix("/assets",
				http.FileServer(http.Dir("assets")))))

	handler := ServerHandler{
		Storage:    storage,
		KeymapFile: keymapFile,
	}
	mux.Handle("/combo", http.HandlerFunc(handler.CombosHandle))
	mux.Handle("/neighbors", http.HandlerFunc(handler.NeighborsHandle))
	mux.Handle("/", http.HandlerFunc(handler.StatsHandle))

	return mux
}

func StartServer(port int, storage db.Storage, keymapFile string, dev bool) {
	log.Printf("Running interface on port %d\n", port)

	err := http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		BuildServer(storage, keymapFile, dev))
	if err != nil {
		log.Fatalf("Could not run server: %s", err)
	}
}

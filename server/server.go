package server

import (
	"fmt"
	"glover/db"
	"html/template"
	"log"
	"net/http"
)

type ServerHandler struct {
	Storage db.Storage
}

type Item struct {
	Label   string
	Visible bool
}

type RenderContext struct {
	TotalCols int
	Items     []Item
	MaxVal    int
}

type Location struct {
	Row int
	Col int
}

var tpl *template.Template = template.Must(template.ParseFiles("templates/heatmap.gohtml"))

func (s *ServerHandler) StatsHandle(w http.ResponseWriter, r *http.Request) {
	curStats, err := s.Storage.GatherAll()
	if err != nil {
		log.Printf("Could not get stats: %s", err.Error())
	}

	totalRows := 0
	totalCols := 0
	maxVal := 0

	// put empty items in the map so that we show them later properly
	groupedItems := make(map[Location]db.MinimalKeyEvent)
	for _, key := range locationsOnGrid {
		groupedItems[key] = db.MinimalKeyEvent{Row: key.Row, Col: key.Col, Count: 0}
	}

	// set non-zero items in the map
	for _, key := range curStats {
		loc, ok := locationsOnGrid[key.Position]
		if !ok {
			log.Printf("Could not find position %d, wtf", key.Position)
		}
		if loc.Row > totalRows {
			totalRows = loc.Row
		}
		if loc.Col > totalCols {
			totalCols = loc.Col
		}
		if maxVal < key.Count {
			maxVal = key.Count
		}

		groupedItems[Location{Row: loc.Row, Col: loc.Col}] = key
	}

	// Iterate over total grid and add real and hidden items.
	// TODO: can this be done without a bunch of hidden items?
	items := make([]Item, 0)
	l := Location{0, 0}
	for i := 0; i <= totalRows; i++ {
		for j := 0; j <= totalCols; j++ {

			l.Row = i
			l.Col = j

			item, ok := groupedItems[l]
			if ok {
				// items = append(items, Item{fmt.Sprintf("(%d %d): %d", item.Row, item.Col, item.Count), true})
				items = append(items, Item{fmt.Sprintf("%d", item.Count), true})
			} else {
				items = append(items, Item{"-", false})
			}
		}
	}

	log.Printf("Rendering...totalCols:%d", totalCols)
	err = tpl.Execute(w, RenderContext{18, items, maxVal})
	if err != nil {
		log.Printf("Could not render: %s", err.Error())
	}
}

func BuildServer(storage db.Storage) *http.ServeMux {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	handler := ServerHandler{storage}
	mux.Handle("/", http.HandlerFunc(handler.StatsHandle))
	return mux
}

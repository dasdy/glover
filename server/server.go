package server

import (
	"bytes"
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
	Position string
	Label    string
	Visible  bool
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
	log.Print("Got request to stats page")

	curStats, err := s.Storage.GatherAll()
	if err != nil {
		log.Printf("Could not get stats: %s", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
			elementId := fmt.Sprintf("%d", item.Position)
			if ok {
				// items = append(items, Item{fmt.Sprintf("(%d %d): %d", item.Row, item.Col, item.Count), true})
				items = append(items, Item{elementId, fmt.Sprintf("%d", item.Count), true})
			} else {
				items = append(items, Item{elementId, "-", false})
			}
		}
	}

	// Do not write to w because it implies 200 status
	var buf bytes.Buffer
	err = tpl.Execute(&buf, RenderContext{18, items, maxVal})
	if err != nil {
		log.Printf("Could not render: %s", err.Error())
		// fmt.Fprintf(w, "Could not render: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Template executed successfully to the buffer.
	// Now, copy it over to the ResponseWriter
	// This implies a 200 OK status code
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	buf.WriteTo(w)
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

func BuildServer(storage db.Storage, dev bool) *http.ServeMux {
	mux := http.NewServeMux()
	// Serve the JS bundle.
	mux.Handle("/assets/",
		disableCacheInDevMode(dev,
			http.StripPrefix("/assets",
				http.FileServer(http.Dir("assets")))))
	handler := ServerHandler{storage}
	mux.Handle("/", http.HandlerFunc(handler.StatsHandle))
	return mux
}

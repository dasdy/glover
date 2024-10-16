package server

import (
	"fmt"
	"glover/db"
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

func (s *ServerHandler) StatsHandle(w http.ResponseWriter, r *http.Request) {
	curStats, err := s.Storage.GatherAll()
	if err != nil {
		log.Printf("Could not get stats: %s", err.Error())
	}

	totalRows := 0
	totalCols := 0
	maxVal := 0

	groupedItems := make(map[Location]db.MinimalKeyEvent)

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

package web

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/layout"
	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/routes"
)

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

	handler := routes.ServerHandler{
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

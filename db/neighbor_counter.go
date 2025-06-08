package db

import (
	"iter"
	"log/slog"
	"sync"

	"github.com/dasdy/glover/model"
)

// NeighborCounterImpl implements the NeighborCounter interface.
type NeighborCounterImpl struct {
	lastKey   model.KeyPosition
	counts    map[model.KeyPosition]map[model.KeyPosition]int
	stateLock sync.RWMutex
}

// NewNeighborCounter creates a new NeighborCounter.
func newNeighborCounter() *NeighborCounterImpl {
	return &NeighborCounterImpl{
		lastKey:   -1,
		counts:    make(map[model.KeyPosition]map[model.KeyPosition]int),
		stateLock: sync.RWMutex{},
	}
}

func NewNeighborCounterFromDb(storage Storage) (*NeighborCounterImpl, error) {
	tracker := newNeighborCounter()

	go func() {
		iterator, err := storage.AllIterator()
		if err != nil {
			panic(err)
		}

		tracker.initCounter(iterator)
	}()

	return tracker, nil
}

func (nc *NeighborCounterImpl) HandleKeyNow(position model.KeyPosition, pressed bool, verbose bool) {
	nc.handleKey(position, pressed, verbose)
}

// GetAllNeighborCounts returns all recorded neighbor counts.
func (nc *NeighborCounterImpl) GatherCombos(position model.KeyPosition) []model.Combo {
	counts := nc.counts[position]

	result := make([]model.Combo, 0, len(counts))

	for k, v := range counts {
		result = append(result, model.Combo{
			Keys:    []model.KeyPosition{k, position},
			Pressed: v,
		})
	}

	return result
}

func (nc *NeighborCounterImpl) initCounter(items iter.Seq[model.KeyEventWithTimestamp]) {
	nc.stateLock.Lock()
	defer nc.stateLock.Unlock()

	for item := range items {
		if !item.Pressed {
			continue
		}

		nc.handleKey(item.Position, item.Pressed, false)
	}
}

// RecordKeyPress records a key press and updates neighbor counts.
func (nc *NeighborCounterImpl) handleKey(position model.KeyPosition, pressed, verbose bool) {
	// only process keypresses, not key releases
	if !pressed {
		return
	}

	if nc.lastKey >= 0 {
		// Initialize the map for the last key if it doesn't exist
		if _, exists := nc.counts[nc.lastKey]; !exists {
			nc.counts[nc.lastKey] = make(map[model.KeyPosition]int)
		}

		if verbose {
			slog.Info("key press sequence",
				"current", position,
				"previous", nc.lastKey)
		}
		// Increment the count for this neighbor pair
		nc.counts[nc.lastKey][position]++
	}

	// Update the last key pressed
	nc.lastKey = position
}

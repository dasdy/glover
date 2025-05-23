package db

import (
	"iter"
	"log"
	"sync"

	"github.com/dasdy/glover/model"
)

// NeighborCounterImpl implements the NeighborCounter interface.
type NeighborCounterImpl struct {
	lastKey   int
	counts    map[int]map[int]int
	stateLock sync.RWMutex
}

// NewNeighborCounter creates a new NeighborCounter.
func newNeighborCounter() *NeighborCounterImpl {
	return &NeighborCounterImpl{
		lastKey:   -1,
		counts:    make(map[int]map[int]int),
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

func (nc *NeighborCounterImpl) HandleKeyNow(position int, pressed bool, verbose bool) {
	nc.handleKey(position, pressed, verbose)
}

// GetAllNeighborCounts returns all recorded neighbor counts.
func (nc *NeighborCounterImpl) GatherCombos(position int) []model.Combo {
	counts := nc.counts[position]

	result := make([]model.Combo, 0, len(counts))

	for k, v := range counts {
		result = append(result, model.Combo{
			Keys: []model.ComboKey{
				{Position: k},
				{Position: position},
			},
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
func (nc *NeighborCounterImpl) handleKey(position int, pressed, verbose bool) {
	// only process keypresses, not key releases
	if !pressed {
		return
	}

	if nc.lastKey >= 0 {
		// Initialize the map for the last key if it doesn't exist
		if _, exists := nc.counts[nc.lastKey]; !exists {
			nc.counts[nc.lastKey] = make(map[int]int)
		}

		if verbose {
			log.Printf("Key %d pressed after %d", position, nc.lastKey)
		}
		// Increment the count for this neighbor pair
		nc.counts[nc.lastKey][position]++
	}

	// Update the last key pressed
	nc.lastKey = position
}

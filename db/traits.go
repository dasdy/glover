package db

import (
	"iter"

	"github.com/dasdy/glover/model"
)

// NeighborCounter tracks and counts keys pressed directly before or after each other.
type Tracker interface {
	HandleKeyNow(position model.KeyPosition, pressed bool, verbose bool)
	GatherCombos(position model.KeyPosition) []model.Combo
}

type Storage interface {
	Store(event *model.KeyEvent) error
	GatherAll() ([]model.MinimalKeyEvent, error)
	AllIterator() (iter.Seq[model.KeyEventWithTimestamp], error)
	Close()
}

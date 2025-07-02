package routes_test

import (
	"iter"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/routes"
)

// Constants for KeyPosition to match key names
const (
	KeyA model.KeyPosition = iota
	KeyB
	KeyC
	KeyD
)

// SimpleStorageMock is a simple manual mock implementation of the Storage interface
type SimpleStorageMock struct {
	ReturnStats []model.MinimalKeyEvent
	ReturnError error
	CallCount   int
}

func (m *SimpleStorageMock) GatherAll() ([]model.MinimalKeyEvent, error) {
	m.CallCount++
	return m.ReturnStats, m.ReturnError
}

// Implement AllIterator method required by db.Storage interface with correct signature
func (m *SimpleStorageMock) AllIterator() (iter.Seq[model.KeyEventWithTimestamp], error) {
	// Return a simple iterator function that yields nothing
	return func(yield func(model.KeyEventWithTimestamp) bool) {
		// Empty iterator - no events to yield
	}, nil
}

// Implement Close method required by db.Storage interface
func (m *SimpleStorageMock) Close() {
	// No-op for testing
}

// Implement Store method required by db.Storage interface
func (m *SimpleStorageMock) Store(event *model.KeyEvent) error {
	// No-op for testing
	return nil
}

// TrackerMock is a simple mock implementation of the Tracker interface
type TrackerMock struct {
	ReturnCombos []model.Combo
	CallCount    int
	LastPosition model.KeyPosition
}

func (m *TrackerMock) HandleKeyNow(position model.KeyPosition, pressed bool, verbose bool) {
	// No-op for testing
}

func (m *TrackerMock) GatherCombos(position model.KeyPosition) []model.Combo {
	m.CallCount++
	m.LastPosition = position
	return m.ReturnCombos
}

// createTestKeyboardLayout creates a standard keyboard layout for testing
func createTestKeyboardLayout() (*model.KeyboardLayout, []string) {
	// Convert keyNames map to slice for ServerHandler
	keyNamesSlice := make([]string, 4) // Allocate enough for all keys we'll use (A, B, C, D)
	keyNamesSlice[KeyA] = "A"
	keyNamesSlice[KeyB] = "B"
	keyNamesSlice[KeyC] = "C"
	keyNamesSlice[KeyD] = "D"

	locations := map[model.KeyPosition]model.Location{
		KeyA: {RowCol: model.RowCol{Row: 0, Col: 0}},
		KeyB: {RowCol: model.RowCol{Row: 0, Col: 1}},
		KeyC: {RowCol: model.RowCol{Row: 1, Col: 0}},
	}

	locationsOnGrid := model.KeyboardLayout{
		Rows:      2,
		Cols:      2,
		Locations: locations,
	}

	return &locationsOnGrid, keyNamesSlice
}

// MockServerHandler helper struct for testing
type MockServerHandler struct {
	routes.ServerHandler
	MockStorage *SimpleStorageMock
}

// MockComboServerHandler extends MockServerHandler to include a mock tracker
type MockComboServerHandler struct {
	routes.ServerHandler
	MockStorage *SimpleStorageMock
	MockTracker *TrackerMock
}

// MockNeighborServerHandler extends MockServerHandler to include a mock neighbor tracker
type MockNeighborServerHandler struct {
	routes.ServerHandler
	MockStorage         *SimpleStorageMock
	MockNeighborTracker *TrackerMock
}

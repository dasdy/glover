package routes_test

import (
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/routes"
	"github.com/stretchr/testify/assert"
)

// Constants for KeyPosition to match key names.
const (
	KeyA model.KeyPosition = iota
	KeyB
	KeyC
	KeyD
)

// SimpleStorageMock is a simple manual mock implementation of the Storage interface.
type SimpleStorageMock struct {
	ReturnStats []model.MinimalKeyEvent
	ReturnError error
	CallCount   int
}

func (m *SimpleStorageMock) GatherAll() ([]model.MinimalKeyEvent, error) {
	m.CallCount++

	return m.ReturnStats, m.ReturnError
}

// Implement AllIterator method required by db.Storage interface with correct signature.
func (m *SimpleStorageMock) AllIterator() (iter.Seq[model.KeyEventWithTimestamp], error) {
	// Return a simple iterator function that yields nothing
	return func(_ func(model.KeyEventWithTimestamp) bool) {
		// Empty iterator - no events to yield
	}, nil
}

// Implement Close method required by db.Storage interface.
func (m *SimpleStorageMock) Close() {
	// No-op for testing
}

// Implement Store method required by db.Storage interface.
func (m *SimpleStorageMock) Store(_ *model.KeyEvent) error {
	// No-op for testing
	return nil
}

// TrackerMock is a simple mock implementation of the Tracker interface.
type TrackerMock struct {
	ReturnCombos []model.Combo
	CallCount    int
	LastPosition model.KeyPosition
}

func (m *TrackerMock) HandleKeyNow(_ model.KeyPosition, _ bool, _ bool) {
	// No-op for testing
}

func (m *TrackerMock) GatherCombos(position model.KeyPosition) []model.Combo {
	m.CallCount++
	m.LastPosition = position

	return m.ReturnCombos
}

// MockServerHandler helper struct for testing.
type MockServerHandler struct {
	routes.ServerHandler

	MockStorage *SimpleStorageMock
}

// MockNeighborServerHandler extends MockServerHandler to include a mock neighbor tracker.
type MockNeighborServerHandler struct {
	routes.ServerHandler

	MockStorage         *SimpleStorageMock
	MockNeighborTracker *TrackerMock
	MockComboTracker    *TrackerMock
}

type handleTestCase struct {
	name              string
	queryParam        string
	trackerReturns    []model.Combo
	expectedStatus    int
	shouldCallTracker bool
}

type renderContextTestCase struct {
	name              string
	combos            []model.Combo
	position          model.KeyPosition
	expectedMaxVal    int
	expectedItems     int
	expectedConnCount int
}

func renderContextTestCases() []renderContextTestCase {
	return []renderContextTestCase{
		{
			name:              "Empty combos",
			combos:            []model.Combo{},
			position:          KeyA,
			expectedMaxVal:    0,
			expectedItems:     3, // All keys from the layout
			expectedConnCount: 0,
		},
		{
			name: "Some combos",
			combos: []model.Combo{
				{Keys: []model.KeyPosition{KeyA, KeyB}, Pressed: 5},
				{Keys: []model.KeyPosition{KeyA, KeyC}, Pressed: 10},
			},
			position:          KeyA,
			expectedMaxVal:    10,
			expectedItems:     3, // All keys from the layout
			expectedConnCount: 2,
		},
	}
}

func handleTestCases() []handleTestCase {
	return []handleTestCase{
		{
			name:              "Success case",
			queryParam:        "position=1",
			trackerReturns:    []model.Combo{{Keys: []model.KeyPosition{1, 2}, Pressed: 5}},
			expectedStatus:    http.StatusOK,
			shouldCallTracker: true,
		},
		{
			name:              "Missing position parameter",
			queryParam:        "",
			trackerReturns:    []model.Combo{},
			expectedStatus:    http.StatusBadRequest,
			shouldCallTracker: false,
		},
		{
			name:              "Invalid position parameter",
			queryParam:        "position=abc",
			trackerReturns:    []model.Combo{},
			expectedStatus:    http.StatusBadRequest,
			shouldCallTracker: false,
		},
	}
}

// Helper function to test handlers with similar behavior.
func testHandlerWithMock(
	t *testing.T,
	tests []handleTestCase,
	baseURL string,
	handlerFunc func(MockNeighborServerHandler, http.ResponseWriter, *http.Request),
	trackerAccessor func(*MockNeighborServerHandler) (*[]model.Combo, *int, *model.KeyPosition),
) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := setupMockNeighborServerHandler()

			// Get references to the mock's fields
			returnCombos, callCount, lastPosition := trackerAccessor(&handler)

			// Set up the mock's return values
			*returnCombos = tc.trackerReturns
			*callCount = 0

			// Create a request to pass to our handler
			url := baseURL
			if tc.queryParam != "" {
				url += "?" + tc.queryParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			// Call the handler
			handlerFunc(handler, w, req)

			// Check the status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Verify the tracker was called when expected
			if tc.shouldCallTracker {
				assert.Equal(t, 1, *callCount, "GatherCombos should be called exactly once")
				// If position parameter was provided and valid, check that it was passed to the tracker
				if tc.queryParam == "position=1" {
					assert.Equal(t, model.KeyPosition(1), *lastPosition)
				}
			} else {
				assert.Equal(t, 0, *callCount, "GatherCombos should not be called")
			}
		})
	}
}

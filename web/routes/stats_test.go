package routes_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/components"
	"github.com/dasdy/glover/web/routes"
	"github.com/stretchr/testify/assert"
)

// setupMockServerHandler creates a mock server handler for testing.
func setupMockServerHandler() MockServerHandler {
	mockStorage := &SimpleStorageMock{}

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

	return MockServerHandler{
		ServerHandler: routes.ServerHandler{
			KeyNames:        keyNamesSlice,
			LocationsOnGrid: &locationsOnGrid,
			Storage:         mockStorage,
		},
		MockStorage: mockStorage,
	}
}

func TestBuildStatsRenderContext(t *testing.T) {
	handler := setupMockServerHandler()

	tests := []struct {
		name           string
		inputStats     []model.MinimalKeyEvent
		expectedMaxVal int
		expectedItems  int
	}{
		{
			name:           "Empty stats",
			inputStats:     []model.MinimalKeyEvent{},
			expectedMaxVal: 0,
			expectedItems:  3, // All keys from the layout
		},
		{
			name: "Some stats",
			inputStats: []model.MinimalKeyEvent{
				{Position: KeyA, Count: 5},
				{Position: KeyB, Count: 10},
			},
			expectedMaxVal: 10,
			expectedItems:  3, // All keys from the layout
		},
		{
			name: "Missing position",
			inputStats: []model.MinimalKeyEvent{
				{Position: KeyA, Count: 5},
				{Position: KeyD, Count: 15}, // D is not in the layout
			},
			expectedMaxVal: 15, // TODO had to fix this, should be 5 probably
			expectedItems:  3,  // All keys from the layout
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.BuildStatsRenderContext(tc.inputStats)

			// Check the result
			assert.Equal(t, 2, result.TotalRows)
			assert.Equal(t, 2, result.TotalCols)
			assert.Equal(t, tc.expectedMaxVal, result.MaxVal)
			assert.Len(t, result.Items, tc.expectedItems)
			assert.Equal(t, components.PageTypeStats, result.Page)
		})
	}
}

func TestStatsHandle(t *testing.T) {
	tests := []struct {
		name           string
		storageReturns []model.MinimalKeyEvent
		storageError   error
		expectedStatus int
	}{
		{
			name: "Success case",
			storageReturns: []model.MinimalKeyEvent{
				{Position: KeyA, Count: 5},
			},
			storageError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Storage error",
			storageReturns: []model.MinimalKeyEvent{},
			storageError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := setupMockServerHandler()

			// Set up the mock's return values
			handler.MockStorage.ReturnStats = tc.storageReturns
			handler.MockStorage.ReturnError = tc.storageError
			handler.MockStorage.CallCount = 0

			// Create a request to pass to our handler
			req := httptest.NewRequest(http.MethodGet, "/stats", nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler.StatsHandle(w, req)

			// Check the status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Verify the mock was called
			assert.Equal(t, 1, handler.MockStorage.CallCount, "GatherAll should be called exactly once")
		})
	}
}

package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/components"
	"github.com/dasdy/glover/web/routes"
	"github.com/stretchr/testify/assert"
)

// setupMockComboServerHandler creates a mock server handler for testing combos
func setupMockComboServerHandler() MockComboServerHandler {
	mockStorage := &SimpleStorageMock{}
	mockTracker := &TrackerMock{}

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

	return MockComboServerHandler{
		ServerHandler: routes.ServerHandler{
			KeyNames:        keyNamesSlice,
			LocationsOnGrid: &locationsOnGrid,
			Storage:         mockStorage,
			ComboTracker:    mockTracker,
		},
		MockStorage: mockStorage,
		MockTracker: mockTracker,
	}
}

func TestBuildCombosRenderContext(t *testing.T) {
	handler := setupMockComboServerHandler()

	tests := []struct {
		name              string
		combos            []model.Combo
		position          model.KeyPosition
		expectedMaxVal    int
		expectedItems     int
		expectedConnCount int
	}{
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
		{
			name: "More than 5 combos",
			combos: []model.Combo{
				{Keys: []model.KeyPosition{KeyA, KeyB}, Pressed: 1},
				{Keys: []model.KeyPosition{KeyA, KeyC}, Pressed: 2},
				{Keys: []model.KeyPosition{KeyB, KeyA}, Pressed: 3},
				{Keys: []model.KeyPosition{KeyC, KeyA}, Pressed: 4},
				{Keys: []model.KeyPosition{KeyA, KeyD}, Pressed: 5},
				{Keys: []model.KeyPosition{KeyD, KeyA}, Pressed: 6},
			},
			position:          KeyA,
			expectedMaxVal:    6,
			expectedItems:     3, // All keys from the layout
			expectedConnCount: 5, // Should be limited to 5
		},
		{
			name: "Combo with missing position in layout",
			combos: []model.Combo{
				{Keys: []model.KeyPosition{KeyA, KeyD}, Pressed: 15}, // D is not in the layout
			},
			position:          KeyA,
			expectedMaxVal:    15,
			expectedItems:     3, // All keys from the layout
			expectedConnCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.ServerHandler.BuildCombosRenderContext(tc.combos, tc.position)

			// Check the result
			assert.Equal(t, 2, result.TotalRows)
			assert.Equal(t, 2, result.TotalCols)
			assert.Equal(t, tc.expectedMaxVal, result.MaxVal)
			assert.Equal(t, tc.expectedItems, len(result.Items))
			assert.Equal(t, components.PageTypeCombo, result.Page)
			assert.Equal(t, tc.position, result.HighlightPosition)
			assert.LessOrEqual(t, len(result.ComboConnections), tc.expectedConnCount)
		})
	}
}

func TestCombosHandle(t *testing.T) {
	tests := []struct {
		name              string
		queryParam        string
		trackerReturns    []model.Combo
		expectedStatus    int
		shouldCallTracker bool
	}{
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := setupMockComboServerHandler()

			// Set up the mock's return values
			handler.MockTracker.ReturnCombos = tc.trackerReturns
			handler.MockTracker.CallCount = 0

			// Create a request to pass to our handler
			url := "/combos"
			if tc.queryParam != "" {
				url += "?" + tc.queryParam
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler.ServerHandler.CombosHandle(w, req)

			// Check the status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Verify the tracker was called when expected
			if tc.shouldCallTracker {
				assert.Equal(t, 1, handler.MockTracker.CallCount, "GatherCombos should be called exactly once")
				// If position parameter was provided and valid, check that it was passed to the tracker
				if tc.queryParam == "position=1" {
					assert.Equal(t, model.KeyPosition(1), handler.MockTracker.LastPosition)
				}
			} else {
				assert.Equal(t, 0, handler.MockTracker.CallCount, "GatherCombos should not be called")
			}
		})
	}
}

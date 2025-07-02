package routes_test

import (
	"net/http"
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/components"
	"github.com/stretchr/testify/assert"
)

func TestBuildCombosRenderContext(t *testing.T) {
	handler := setupMockNeighborServerHandler()

	tests := append(renderContextTestCases(), []renderContextTestCase{
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
	}...)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.BuildCombosRenderContext(tc.combos, tc.position)

			// Check the result
			assert.Equal(t, 2, result.TotalRows)
			assert.Equal(t, 2, result.TotalCols)
			assert.Equal(t, tc.expectedMaxVal, result.MaxVal)
			assert.Len(t, result.Items, tc.expectedItems)
			assert.Equal(t, components.PageTypeCombo, result.Page)
			assert.Equal(t, tc.position, result.HighlightPosition)
			assert.LessOrEqual(t, len(result.ComboConnections), tc.expectedConnCount)
		})
	}
}

func TestCombosHandle(t *testing.T) {
	tests := handleTestCases()

	testHandlerFunc := func(handler MockNeighborServerHandler, w http.ResponseWriter, r *http.Request) {
		handler.CombosHandle(w, r)
	}

	testHandlerWithMock(t, tests, "/combos", testHandlerFunc, func(handler *MockNeighborServerHandler) (*[]model.Combo, *int, *model.KeyPosition) {
		return &handler.MockComboTracker.ReturnCombos, &handler.MockComboTracker.CallCount, &handler.MockComboTracker.LastPosition
	})
}

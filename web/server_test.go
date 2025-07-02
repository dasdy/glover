package web_test

import (
	"testing"

	"github.com/dasdy/glover/layout"
	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/components"
	"github.com/dasdy/glover/web/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openKeymapFile(t *testing.T, filepath string) *model.KeyboardLayout {
	t.Helper()

	file, err := layout.OpenPath(filepath)

	require.NoError(t, err)

	layout, err := layout.LoadZmkLocationsJSON(file)

	assert.NoError(t, err)

	return layout
}

func loadKeyNames(t *testing.T, filepath string) []string {
	t.Helper()

	keyNames, err := layout.GetKeyLabels(filepath)

	assert.NoError(t, err)

	return keyNames
}

func TestBuildStatsContext(t *testing.T) {
	t.Run("builds empty context", func(t *testing.T) {
		stats := make([]model.MinimalKeyEvent, 0)

		handler := routes.ServerHandler{
			Storage:         nil,
			KeyNames:        loadKeyNames(t, "data/glove80.keymap"),
			LocationsOnGrid: openKeymapFile(t, "data/info.json"),
		}
		items := handler.BuildStatsRenderContext(stats)

		assert.Len(t, items.Items, 80)
		assert.Equal(t, components.PageTypeStats, items.Page)
	})
}

func TestBuildCombosContext(t *testing.T) {
	t.Run("builds empty context", func(t *testing.T) {
		stats := make([]model.Combo, 0)
		handler := routes.ServerHandler{
			Storage:         nil,
			KeyNames:        loadKeyNames(t, "data/glove80.keymap"),
			LocationsOnGrid: openKeymapFile(t, "data/info.json"),
		}

		items := handler.BuildCombosRenderContext(stats, 10)

		assert.Len(t, items.Items, 80)
		assert.Equal(t, components.PageTypeCombo, items.Page)
	})
}

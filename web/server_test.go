package web_test

import (
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web"
	"github.com/stretchr/testify/assert"
)

func TestBuildStatsContext(t *testing.T) {
	t.Run("builds empty context", func(t *testing.T) {
		stats := make([]model.MinimalKeyEvent, 0)

		handler := web.ServerHandler{
			Storage:    nil,
			KeymapFile: "glove80.keymap",
		}
		items := handler.BuildStatsRenderContext(stats)

		assert.Len(t, items.Items, 80)
	})
}

func TestBuildCombosContext(t *testing.T) {
	t.Run("builds empty context", func(t *testing.T) {
		stats := make([]model.Combo, 0)
		handler := web.ServerHandler{
			Storage:    nil,
			KeymapFile: "glove80.keymap",
		}

		items := handler.BuildCombosRenderContext(stats, 10)

		assert.Len(t, items.Items, 80)
	})
}

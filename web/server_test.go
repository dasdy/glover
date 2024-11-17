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

		items := web.BuildStatsRenderContext(stats)

		assert.Len(t, items.Items, 96)

		visibleItems := 0

		for i, v := range items.Items {
			if v.Visible {
				assert.Equal(t, "0", v.Label, "Bad label on %d: %s", i, v.Label)
				visibleItems++
			} else {
				assert.Equal(t, "-", v.Label, "Bad label on %d: %s", i, v.Label)
			}
		}

		assert.Equal(t, 80, visibleItems)
	})
}

func TestBuildCombosContext(t *testing.T) {
	t.Run("builds empty context", func(t *testing.T) {
		stats := make([]model.Combo, 0)

		items := web.BuildCombosRenderContext(stats, 10)

		assert.Len(t, items.Items, 96)

		visibleItems := 0

		for i, v := range items.Items {
			if v.Visible {
				assert.Equal(t, "0", v.Label, "Bad label on %d: %s", i, v.Label)
				visibleItems++
			} else {
				assert.Equal(t, "-", v.Label, "Bad label on %d: %s", i, v.Label)
			}
		}

		assert.Equal(t, 80, visibleItems)
	})
}

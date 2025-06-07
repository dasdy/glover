package layout_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dasdy/glover/layout"
	"github.com/stretchr/testify/assert"
)

func TestGlove80ParseLayout(t *testing.T) {
	t.Run("Parses glove80 file", func(t *testing.T) {
		_, b, _, _ := runtime.Caller(0)

		// Root folder of this project
		fp := filepath.Join(filepath.Dir(b), "..")

		file, err := os.Open(filepath.Join(fp, "data", "glove80.keymap"))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		keymap, err := layout.Parse(file)
		if err != nil {
			t.Fatal(err)
		}

		if keymap == nil {
			t.Fatal("Expected keymap to be not nil")
		}

		slog.Info("parsed layout", "keymap", keymap)

		for i, l := range keymap.Layers {
			slog.Info("layer info",
				"index", i,
				"name", l.Name,
				"binding_count", len(l.Bindings))
		}

		assert.Len(t, keymap.Layers, 4, "Unexpected amt of layers")
		assert.Len(t, keymap.Layers[0].Bindings, 80, "Unexpected amt of bindings on layer 0")
	})
}

package layout_test

import (
	"log"
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

		log.Printf("Got layout: %v", keymap)

		for i, l := range keymap.Layers {
			log.Printf("Layer %d: '%s' (%d)", i, l.Name, len(l.Bindings))
		}

		assert.Equal(t, 4, len(keymap.Layers), "Unexpected amt of layers")
		assert.Equal(t, 80, len(keymap.Layers[0].Bindings), "Unexpected amt of bindings on layer 0")
	})
}

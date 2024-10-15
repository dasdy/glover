package db_test

import (
	"glover/db"
	"glover/keylog/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectToMemoryDB(t *testing.T) {
	t.Run("should insert and gather correctly", func(t *testing.T) {
		storage, error := db.ConnectDB(":memory:")

		assert.NoError(t, error)

		items, error := storage.GatherAll()

		assert.NoError(t, error)

		assert.Empty(t, items)

		item := parser.KeyEvent{Row: 0, Col: 0, Position: 0, Pressed: false}
		for i := 0; i < 10; i++ {
			item.Pressed = !item.Pressed
			storage.Store(&item)
		}

		item.Pressed = false
		for i := 0; i < 5; i++ {
			item.Col = i*2 + 1
			item.Row = i + 12
			storage.Store(&item)
		}

		items, error = storage.GatherAll()

		assert.NoError(t, error)

		assert.Equal(t, items, map[db.MinimalKeyEvent]int{
			{0, 0, 0}:  5,
			{12, 1, 0}: 1,
			{13, 3, 0}: 1,
			{14, 5, 0}: 1,
			{15, 7, 0}: 1,
			{16, 9, 0}: 1,
		})
	})
}

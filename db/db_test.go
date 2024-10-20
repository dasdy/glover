package db_test

import (
	"glover/db"
	"glover/keylog/parser"
	"log"
	"os"
	"sync"
	"testing"
	"time"

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
			err := storage.Store(&item)
			assert.NoError(t, err)
		}

		item.Pressed = false
		for i := 0; i < 5; i++ {
			item.Col = i*2 + 1
			item.Row = i + 12
			err := storage.Store(&item)
			assert.NoError(t, err)
		}

		items, error = storage.GatherAll()

		assert.NoError(t, error)

		assert.Equal(t, items, []db.MinimalKeyEvent{
			{0, 0, 0, 5},
			{12, 1, 0, 1},
			{13, 3, 0, 1},
			{14, 5, 0, 1},
			{15, 7, 0, 1},
			{16, 9, 0, 1},
		})
	})
}

func TestRaceCondition(t *testing.T) {
	t.Run("Should not fail due to race condition on db connection", func(t *testing.T) {
		file, err := os.CreateTemp("/tmp", "*.sqlite")
		assert.NoError(t, err)
		log.Printf("Created file: %s", file.Name())
		storage, err := db.ConnectDB(file.Name())

		assert.NoError(t, err)

		var wg sync.WaitGroup
		done := make(chan bool, 2)

		wg.Add(2)
		go func() {
			event := parser.KeyEvent{}
		out:
			for i := range 16_000 {
				err := storage.Store(&event)

				assert.NoError(t, err)
				// Writes can take all the cake from reads - give them some time to rest
				if i%2000 == 0 {
					// log.Println("Another 2k items written")
					time.Sleep(100 * time.Millisecond)
				}
				select {
				case <-done:
					break out
				default:
					continue
				}
			}
			wg.Done()
			done <- true
			log.Println("Done writing")
		}()

		go func() {
		out:
			for range 6_000 {
				_, err := storage.GatherAll()

				// if i%500 == 0 {
				// log.Println("Another 500 items read")
				// }
				assert.NoError(t, err)
				select {
				case <-done:
					break out
				default:
					continue
				}
			}
			wg.Done()
			done <- true
			log.Println("Done reading")
		}()

		wg.Wait()

		items, err := storage.GatherAll()
		assert.NoError(t, err)
		assert.Len(t, items, 1)
	})
}

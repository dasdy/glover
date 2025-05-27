package db_test

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockEvents(keyPositions []int) []model.KeyEvent {
	// Every position in array is an event. Repeating positions like 5,5 will
	// result in making two events: first with pressed = True, second with pressed=false
	// Also, row/col locations get jumbled up a bit because I don't really care here abt them, just make
	// them unique for the input range (0-79)
	state := make(map[int]model.KeyEvent)
	values := make([]model.KeyEvent, 0)

	for _, pos := range keyPositions {
		event, ok := state[pos]
		if !ok {
			event = model.KeyEvent{Row: pos, Col: pos, Position: pos, Pressed: true}
		} else {
			event.Pressed = !event.Pressed
		}

		state[pos] = event
		values = append(values, event)
	}

	return values
}

func TestConnectToMemoryDB(t *testing.T) {
	t.Run("should insert and gather correctly", func(t *testing.T) {
		storage, err := db.NewStorageFromPath(":memory:", false)

		require.NoError(t, err)

		items, err := storage.GatherAll()

		require.NoError(t, err)

		assert.Empty(t, items)

		item := model.KeyEvent{Row: 0, Col: 0, Position: 0, Pressed: false}
		for range 10 {
			item.Pressed = !item.Pressed
			require.NoError(t, storage.Store(&item))
		}

		item.Pressed = false
		for i := range 5 {
			item.Col = i*2 + 1
			item.Row = i + 12
			require.NoError(t, storage.Store(&item))
		}

		items, err = storage.GatherAll()

		require.NoError(t, err)

		assert.Equal(t, []model.MinimalKeyEvent{
			{Row: 0, Col: 0, Position: 0, Count: 5},
			{Row: 12, Col: 1, Position: 0, Count: 1},
			{Row: 13, Col: 3, Position: 0, Count: 1},
			{Row: 14, Col: 5, Position: 0, Count: 1},
			{Row: 15, Col: 7, Position: 0, Count: 1},
			{Row: 16, Col: 9, Position: 0, Count: 1},
		},
			items)
	})
}

func TestRaceCondition(t *testing.T) {
	t.Run("Should not fail due to race condition on db connection", func(t *testing.T) {
		file, err := os.CreateTemp("/tmp", "*.sqlite")
		require.NoError(t, err)
		log.Printf("Created file: %s", file.Name())
		storage, err := db.NewStorageFromPath(file.Name(), false)

		require.NoError(t, err)

		var wg sync.WaitGroup

		done := make(chan bool, 2)

		wg.Add(2)

		routine := func() {
			event := model.KeyEvent{}
		out:
			for i := range 16_000 {
				require.NoError(t, storage.Store(&event))
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
		}

		routine2 := func() {
		out:
			for range 6_000 {
				_, err := storage.GatherAll()
				require.NoError(t, err)
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
		}

		go routine()
		go routine2()

		wg.Wait()

		items, err := storage.GatherAll()
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})
}

func copyToMem(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	memConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	if db.InitDBStorage(memConn) != nil {
		return nil, err
	}

	rows, err := conn.Query(
		`select row, col, position, pressed, ts 
        from keypresses
        order by ts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			row, col, position int
			pressed            bool
			ts                 time.Time
		)

		if rows.Scan(&row, &col, &position, &pressed, &ts) != nil {
			return nil, err
		}

		_, err = memConn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
			row, col, position, pressed, ts)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return memConn, err
}

func TestMergeDatabases(t *testing.T) {
	t.Run("merges two storages successfully", func(t *testing.T) {
		file1, err := os.CreateTemp("/tmp", "*.sqlite")
		require.NoError(t, err)
		file2, err := os.CreateTemp("/tmp", "*.sqlite")
		require.NoError(t, err)

		storage1, err := db.NewStorageFromPath(file1.Name(), false)
		require.NoError(t, err)
		defer storage1.Close()

		storage2, err := db.NewStorageFromPath(file2.Name(), false)
		require.NoError(t, err)
		defer storage2.Close()

		event1 := model.KeyEvent{
			Row: 5, Col: 100, Position: 5, Pressed: false,
		}
		require.NoError(t, storage1.Store(&event1))

		event2 := model.KeyEvent{
			Row: 102, Col: 110, Position: 6, Pressed: true,
		}
		require.NoError(t, storage2.Store(&event2))

		file3, err := os.CreateTemp("/tmp", "*.sqlite")
		require.NoError(t, err)

		output, err := db.NewStorageFromPath(file3.Name(), false)
		require.NoError(t, err)

		require.NoError(t, db.Merge([]*db.SQLiteStorage{storage1, storage2}, output))

		conn, err := sql.Open("sqlite3", file3.Name())
		require.NoError(t, err)
		rows, err := conn.Query(
			`select row, col, position, pressed, ts 
        from keypresses
        order by ts`)
		require.NoError(t, err)

		defer rows.Close()

		assert.True(t, rows.Next())

		var (
			row, col, position int
			pressed            bool
			ts                 time.Time
		)

		require.NoError(t, rows.Scan(&row, &col, &position, &pressed, &ts))
		assert.Equal(t,
			model.KeyEvent{Row: row, Col: col, Position: position, Pressed: pressed},
			event1,
		)

		assert.True(t, rows.Next())
		require.NoError(t, rows.Scan(&row, &col, &position, &pressed, &ts))
		assert.Equal(t,
			model.KeyEvent{Row: row, Col: col, Position: position, Pressed: pressed},
			event2,
		)

		assert.False(t, rows.Next())

		assert.NoError(t, rows.Err())
	})
}

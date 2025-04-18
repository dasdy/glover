package db_test

import (
	"cmp"
	"database/sql"
	"log"
	"os"
	"slices"
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

func sortCombos(result []model.Combo) {
	// TODO: this sort might be not useful outside of tests, but maybe it's not that slow
	// (we are only looking at <200 rows here). Measure how long does it take.
	slices.SortFunc(result, func(a, b model.Combo) int {
		baseCmp := cmp.Or(
			-cmp.Compare(a.Pressed, b.Pressed),
			cmp.Compare(len(a.Keys), len(b.Keys)),
		)
		if baseCmp != 0 {
			return baseCmp
		}

		// if a.keys has different length than b.keys, we wouldn't be here.
		for i := range a.Keys {
			ak := a.Keys[i]
			bk := b.Keys[i]

			if keyCmp := cmp.Compare(ak.Position, bk.Position); keyCmp != 0 {
				return keyCmp
			}
		}

		return 0
	})
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

func TestGatherCombos(t *testing.T) {
	t.Run("returns empty combos by default", func(t *testing.T) {
		storage, err := db.NewStorageFromPath(":memory:", false)
		require.NoError(t, err)

		items := storage.GatherCombos()

		assert.Empty(t, items)
	})

	t.Run("returns one combo", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")

		require.NoError(t, err)

		require.NoError(t, db.InitDBStorage(conn))

		positions := []int{
			1, 2,
			1, 2,
		}
		events := mockEvents(positions)
		log.Printf("Events: %+v", events)

		curTime := time.Now()

		for _, event := range events {
			_, err := conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
				event.Row, event.Col, event.Position, event.Pressed, curTime)
			require.NoError(t, err)

			curTime = curTime.Add(100 * time.Millisecond)
		}

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)

		combos := storage.GatherCombos()

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 1},
					{Position: 2},
				},
				Pressed: 1,
			},
		}, combos)
	})
	t.Run("returns plain item count for complicated thing", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")

		assert.NoError(t, err)
		assert.NoError(t, db.InitDBStorage(conn))

		positions := []int{
			1, 2,
			1, 2,
			3, 1, 4, 3, 4, 1,
		}
		events := mockEvents(positions)
		log.Printf("Events: %+v", events)

		curTime := time.Now()

		for _, event := range events {
			_, err := conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
				event.Row, event.Col, event.Position, event.Pressed, curTime)
			require.NoError(t, err)

			curTime = curTime.Add(100 * time.Millisecond)
		}

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)

		combos := storage.GatherCombos()

		sortCombos(combos)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 1},
					{Position: 2},
				},
				Pressed: 1,
			},
			{
				Keys: []model.ComboKey{
					{Position: 1},
					{Position: 3},
				},
				Pressed: 1,
			},
			{
				Keys: []model.ComboKey{
					{Position: 1},
					{Position: 4},
				},
				Pressed: 1,
			},
			{
				Keys:    []model.ComboKey{{Position: 1}, {Position: 3}, {Position: 4}},
				Pressed: 1,
			},
		}, combos)
	})
	t.Run("ignores items that happened too long ago", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)

		require.NoError(t, db.InitDBStorage(conn))

		positions := []int{
			1, 2, 1, 2, // Valid combo
			3, 4, // Too old
		}
		events := mockEvents(positions)

		curTime := time.Now()

		// Insert valid combo events
		for _, event := range events[:4] {
			_, err := conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
				event.Row, event.Col, event.Position, event.Pressed, curTime)
			require.NoError(t, err)

			curTime = curTime.Add(100 * time.Millisecond)
		}

		// Insert old events
		oldTime := curTime.Add(-10 * time.Minute)
		for _, event := range events[4:] {
			_, err := conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
				event.Row, event.Col, event.Position, event.Pressed, oldTime)
			require.NoError(t, err)
		}

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)

		combos := storage.GatherCombos()

		assert.ElementsMatch(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 1},
					{Position: 2},
				},
				Pressed: 1,
			},
			{
				Keys: []model.ComboKey{
					{Position: 3},
					{Position: 4},
				},
				Pressed: 1,
			},
		}, combos)
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

func BenchmarkComboScan(b *testing.B) {
	conn, err := copyToMem("./../keypresses.sqlite")
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	for range b.N {
		if _, err = db.NewComboTrackerFromDB(conn); err != nil {
			b.Fatal(err)
		}
	}
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
func TestGatherNeighbors(t *testing.T) {
	t.Run("returns empty result for empty database", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(1)
		require.NoError(t, err)
		assert.Empty(t, neighbors)
	})

	t.Run("returns empty result for key with no neighbors", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()
		event := model.KeyEvent{Row: 5, Col: 5, Position: 5, Pressed: true}
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, event.Row, event.Col, event.Position, event.Pressed, curTime)
		require.NoError(t, err)

		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, event.Row, event.Col, event.Position, false, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)
		assert.Empty(t, neighbors)
	})

	t.Run("detects previous key neighbor", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()

		// Previous key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 3, 3, 3, true, curTime)
		require.NoError(t, err)

		// Target key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)
		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 3},
					{Position: 5},
				},
				Pressed: 1,
			},
		}, neighbors)
	})

	t.Run("detects next key neighbor", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()

		// Target key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime)
		require.NoError(t, err)

		// Next key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 7, 7, 7, true, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)
		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 7},
					{Position: 5},
				},
				Pressed: 1,
			},
		}, neighbors)
	})

	t.Run("detects both previous and next neighbors", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()

		// Previous key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 3, 3, 3, true, curTime)
		require.NoError(t, err)

		// Target key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		// Next key (pressed)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 7, 7, 7, true, curTime.Add(200*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)

		// Sort the results for consistent test comparison
		sortCombos(neighbors)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 3},
					{Position: 5},
				},
				Pressed: 1,
			},
			{
				Keys: []model.ComboKey{
					{Position: 7},
					{Position: 5},
				},
				Pressed: 1,
			},
		}, neighbors)
	})

	t.Run("counts multiple occurrences of same pattern", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()

		// First sequence
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 3, 3, 3, true, curTime)
		require.NoError(t, err)

		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		// Second sequence (same pattern, different time)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 3, 3, 3, true, curTime.Add(1*time.Second))
		require.NoError(t, err)

		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime.Add(1100*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)
		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 3},
					{Position: 5},
				},
				Pressed: 3,
			},
		}, neighbors)
	})

	t.Run("ignores non-immediate neighbors", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		require.NoError(t, db.InitDBStorage(conn))

		curTime := time.Now()

		// First key (not an immediate neighbor)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 1, 1, 1, true, curTime)
		require.NoError(t, err)

		// Second key (immediate neighbor to target)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 3, 3, 3, true, curTime.Add(100*time.Millisecond))
		require.NoError(t, err)

		// Target key
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 5, 5, 5, true, curTime.Add(200*time.Millisecond))
		require.NoError(t, err)

		// Next key (immediate neighbor to target)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 7, 7, 7, true, curTime.Add(300*time.Millisecond))
		require.NoError(t, err)

		// Last key (not an immediate neighbor)
		_, err = conn.Exec(`insert into keypresses(row, col, position, pressed, ts)
		values(?, ?, ?, ?, ?)`, 9, 9, 9, true, curTime.Add(400*time.Millisecond))
		require.NoError(t, err)

		storage, err := db.NewStorageFromConnection(conn, false)
		require.NoError(t, err)
		defer storage.Close()

		neighbors, err := storage.GatherNeighbors(5)
		require.NoError(t, err)

		// Sort the results for consistent test comparison
		sortCombos(neighbors)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.ComboKey{
					{Position: 3},
					{Position: 5},
				},
				Pressed: 1,
			},
			{
				Keys: []model.ComboKey{
					{Position: 7},
					{Position: 5},
				},
				Pressed: 1,
			},
		}, neighbors)

		// Key 1 and key 9 should not be in the results
		for _, combo := range neighbors {
			for _, key := range combo.Keys {
				assert.NotEqual(t, 1, key.Position, "Key position 1 should not be in results")
				assert.NotEqual(t, 9, key.Position, "Key position 9 should not be in results")
			}
		}
	})
}

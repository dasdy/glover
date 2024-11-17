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
		for i := 0; i < len(a.Keys); i++ {
			ak := a.Keys[i]
			bk := b.Keys[i]
			keyCmp := cmp.Compare(ak.Position, bk.Position)
			if keyCmp != 0 {
				return keyCmp
			}
		}
		return 0
	})
}

func TestConnectToMemoryDB(t *testing.T) {
	t.Run("should insert and gather correctly", func(t *testing.T) {
		storage, err := db.ConnectDB(":memory:")

		assert.NoError(t, err)

		items, err := storage.GatherAll()

		assert.NoError(t, err)

		assert.Empty(t, items)

		item := model.KeyEvent{Row: 0, Col: 0, Position: 0, Pressed: false}
		for i := 0; i < 10; i++ {
			item.Pressed = !item.Pressed
			assert.NoError(t, storage.Store(&item))
		}

		item.Pressed = false
		for i := 0; i < 5; i++ {
			item.Col = i*2 + 1
			item.Row = i + 12
			assert.NoError(t, storage.Store(&item))
		}

		items, err = storage.GatherAll()

		assert.NoError(t, err)

		assert.Equal(t, items, []model.MinimalKeyEvent{
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
			event := model.KeyEvent{}
		out:
			for i := range 16_000 {
				assert.NoError(t, storage.Store(&event))
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

func TestGatherCombos(t *testing.T) {
	t.Run("returns empty combos by default", func(t *testing.T) {
		storage, err := db.ConnectDB(":memory:")
		assert.NoError(t, err)

		items := storage.GatherCombos()

		assert.Len(t, items, 0)
	})

	t.Run("returns one combo", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")

		assert.NoError(t, err)

		assert.NoError(t, db.InitDbStorage(conn))

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
			assert.NoError(t, err)

			curTime = curTime.Add(100 * time.Millisecond)
		}

		storage, err := db.NewStorage(conn)
		assert.NoError(t, err)
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
		assert.NoError(t, db.InitDbStorage(conn))

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
			assert.NoError(t, err)

			curTime = curTime.Add(100 * time.Millisecond)
		}

		storage, err := db.NewStorage(conn)
		assert.NoError(t, err)
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
}

func copyToMem(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	memConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	err = db.InitDbStorage(memConn)
	if err != nil {
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
		var row, col, position int
		var pressed bool
		var ts time.Time

		err := rows.Scan(&row, &col, &position, &pressed, &ts)
		if err != nil {
			return nil, err
		}

		_, err = memConn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
			row, col, position, pressed, ts)
		if err != nil {
			return nil, err
		}
	}

	return memConn, err
}

func BenchmarkComboScan(b *testing.B) {
	conn, err := copyToMem("./../keypresses.sqlite")
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < b.N; i++ {
		_, err = db.NewComboTrackerFromDb(conn)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMergeDatabases(t *testing.T) {
	t.Run("merges two storages successfully", func(t *testing.T) {
		file1, err := os.CreateTemp("/tmp", "*.sqlite")
		assert.NoError(t, err)
		file2, err := os.CreateTemp("/tmp", "*.sqlite")
		assert.NoError(t, err)

		storage1, err := db.ConnectDB(file1.Name())
		assert.NoError(t, err)
		defer storage1.Close()
		storage2, err := db.ConnectDB(file2.Name())
		assert.NoError(t, err)
		defer storage2.Close()
		event1 := model.KeyEvent{
			Row: 5, Col: 100, Position: 5, Pressed: false,
		}
		assert.NoError(t, storage1.Store(&event1))
		event2 := model.KeyEvent{
			Row: 102, Col: 110, Position: 6, Pressed: true,
		}
		assert.NoError(t, storage2.Store(&event2))

		file3, err := os.CreateTemp("/tmp", "*.sqlite")
		assert.NoError(t, err)

		output, err := db.ConnectDB(file3.Name())

		assert.NoError(t, db.Merge([]*db.SQLiteStorage{storage1, storage2}, output))

		conn, err := sql.Open("sqlite3", file3.Name())
		rows, err := conn.Query(
			`select row, col, position, pressed, ts 
        from keypresses
        order by ts`)

		assert.True(t, rows.Next())
		var row, col, position int
		var pressed bool
		var ts time.Time

		assert.NoError(t, rows.Scan(&row, &col, &position, &pressed, &ts))
		assert.Equal(t,
			event1,
			model.KeyEvent{Row: row, Col: col, Position: position, Pressed: pressed},
		)

		assert.True(t, rows.Next())
		assert.NoError(t, rows.Scan(&row, &col, &position, &pressed, &ts))
		assert.Equal(t,
			event2,
			model.KeyEvent{Row: row, Col: col, Position: position, Pressed: pressed},
		)

		assert.False(t, rows.Next())
	})
}

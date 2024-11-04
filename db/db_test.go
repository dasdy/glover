package db_test

import (
	"cmp"
	"database/sql"
	"glover/db"
	"glover/keylog/parser"
	"log"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockEvents(keyPositions []int) []parser.KeyEvent {
	// Every position in array is an event. Repeating positions like 5,5 will
	// result in making two events: first with pressed = True, second with pressed=false
	// Also, row/col locations get jumbled up a bit because I don't really care here abt them, just make
	// them unique for the input range (0-79)
	state := make(map[int]parser.KeyEvent)
	values := make([]parser.KeyEvent, 0)

	for _, pos := range keyPositions {
		event, ok := state[pos]
		if !ok {
			event = parser.KeyEvent{Row: pos, Col: pos, Position: pos, Pressed: true}
		} else {
			event.Pressed = !event.Pressed
		}
		state[pos] = event
		values = append(values, event)

	}

	return values
}

func sortCombos(result []db.Combo) {
	// TODO: this sort might be not useful outside of tests, but maybe it's not that slow
	// (we are only looking at <200 rows here). Measure how long does it take.
	slices.SortFunc(result, func(a, b db.Combo) int {
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

func TestGatherCombos(t *testing.T) {
	t.Run("returns empty combos by default", func(t *testing.T) {
		storage, error := db.ConnectDB(":memory:")
		assert.NoError(t, error)

		items, error := storage.GatherCombos(2)

		assert.NoError(t, error)
		assert.Len(t, items, 0)
	})

	t.Run("returns one combo", func(t *testing.T) {
		conn, error := sql.Open("sqlite3", ":memory:")

		assert.NoError(t, error)
		error = db.InitDbStorage(conn)

		assert.NoError(t, error)

		storage := db.NewStorage(conn)

		assert.NoError(t, error)
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

		combos, error := storage.GatherCombos(2)
		assert.NoError(t, error)

		assert.Equal(t, []db.Combo{
			{
				[]db.ComboKey{
					{1},
					{2},
				},
				1,
			},
		}, combos)
	})
	t.Run("returns plain item count for complicated thing", func(t *testing.T) {
		conn, error := sql.Open("sqlite3", ":memory:")

		assert.NoError(t, error)
		error = db.InitDbStorage(conn)

		assert.NoError(t, error)
		storage := db.NewStorage(conn)

		assert.NoError(t, error)
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

		combos, error := storage.GatherCombos(2)
		assert.NoError(t, error)

		sortCombos(combos)

		assert.Equal(t, []db.Combo{
			{
				[]db.ComboKey{
					{1},
					{2},
				},
				1,
			},
			{
				[]db.ComboKey{
					{1},
					{3},
				},
				1,
			},
			{
				[]db.ComboKey{
					{1},
					{4},
				},
				1,
			},
			{
				[]db.ComboKey{
					{1},
					{3},
					{4},
				},
				1,
			},
		}, combos)
	})

	t.Run("show other combos", func(t *testing.T) {
		storage, error := db.ConnectDB("./../keypresses.sqlite")
		assert.NoError(t, error)

		_, error = storage.GatherCombos(2)
		assert.NoError(t, error)

		// log.Println("Combos:[")
		// for _, i := range items {
		// 	log.Printf("%v", i)
		// }
		// log.Println("]")
	})
}

func copyToMem(path string) (*sql.DB, error) {
	conn, error := sql.Open("sqlite3", path)
	if error != nil {
		return nil, error
	}
	memConn, error := sql.Open("sqlite3", ":memory:")
	if error != nil {
		return nil, error
	}
	error = db.InitDbStorage(memConn)

	if error != nil {
		return nil, error
	}

	rows, error := conn.Query(
		`select row, col, position, pressed, ts 
        from keypresses
        order by ts`)

	if error != nil {
		return nil, error
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

		_, error := memConn.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, ?)`,
			row, col, position, pressed, ts)

		if error != nil {
			return nil, error
		}

	}

	return memConn, error
}

func BenchmarkComboScan(b *testing.B) {
	conn, error := copyToMem("./../keypresses.sqlite")
	if error != nil {
		b.Fatal(error)
	}
	stmt, error := conn.Prepare(`select position, pressed, ts from keypresses order by ts`)
	if error != nil {
		b.Fatal(error)
	}
	for i := 0; i < b.N; i++ {
		rows, error := stmt.Query()
		if error != nil {
			b.Fatal(error)
		}

		_, error = db.ScanForCombos(rows, 2)
		if error != nil {
			b.Fatal(error)
		}
	}
}

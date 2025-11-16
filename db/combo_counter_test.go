package db_test

import (
	"cmp"
	"database/sql"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

			if keyCmp := cmp.Compare(ak, bk); keyCmp != 0 {
				return keyCmp
			}
		}

		return 0
	})
}

func TestGatherCombos(t *testing.T) {
	t.Run("returns empty combos by default", func(t *testing.T) {
		storage, err := db.NewStorageFromPath(":memory:", false)
		require.NoError(t, err)

		tracker, err := db.NewComboTrackerFromDB(storage)
		require.NoError(t, err)

		items := tracker.GatherCombos(1)

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
		slog.Info("testing events", "events", events)

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

		tracker, err := db.NewComboTrackerFromDB(storage)
		require.NoError(t, err)

		time.Sleep(120 * time.Millisecond)

		combos := tracker.GatherCombos(1)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(2),
				},
				Pressed: 1,
			},
		}, combos)
	})

	t.Run("returns one combo pressed twice", func(t *testing.T) {
		conn, err := sql.Open("sqlite3", ":memory:")

		require.NoError(t, err)

		require.NoError(t, db.InitDBStorage(conn))

		positions := []int{
			1, 2,
			1, 2,
			1, 2,
			1, 2,
		}
		events := mockEvents(positions)
		slog.Info("testing events", "events", events)

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

		tracker, err := db.NewComboTrackerFromDB(storage)
		require.NoError(t, err)

		time.Sleep(120 * time.Millisecond)

		combos := tracker.GatherCombos(1)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(2),
				},
				Pressed: 2,
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
		slog.Info("testing events", "events", events)

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

		tracker, err := db.NewComboTrackerFromDB(storage)
		require.NoError(t, err)

		time.Sleep(120 * time.Millisecond)

		combos := tracker.GatherCombos(1)

		sortCombos(combos)

		assert.Equal(t, []model.Combo{
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(2),
				},
				Pressed: 1,
			},
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(3),
				},
				Pressed: 1,
			},
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(4),
				},
				Pressed: 1,
			},
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(3),
					model.KeyPosition(4),
				},
				Pressed: 1,
			},
		}, combos)

		combos = tracker.GatherCombos(6)
		assert.Empty(t, combos)
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

		tracker, err := db.NewComboTrackerFromDB(storage)
		require.NoError(t, err)

		time.Sleep(120 * time.Millisecond)

		combos := tracker.GatherCombos(1)
		combos = append(combos, tracker.GatherCombos(3)...)

		assert.ElementsMatch(t, []model.Combo{
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(1),
					model.KeyPosition(2),
				},
				Pressed: 1,
			},
			{
				Keys: []model.KeyPosition{
					model.KeyPosition(3),
					model.KeyPosition(4),
				},
				Pressed: 1,
			},
		}, combos)
	})
}

func BenchmarkComboScan(b *testing.B) {
	conn, err := copyToMem("./../keypresses.sqlite")
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	storage, err := db.NewStorageFromConnection(conn, false)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err = db.NewComboTrackerFromDB(storage); err != nil {
			b.Fatal(err)
		}
	}
}

func TestComboKeyId(t *testing.T) {
	t.Run("empty key set should return empty bitmask", func(t *testing.T) {
		keys := []model.KeyPosition{}
		mask := db.ComboKeyID(keys)

		assert.Equal(t, db.ComboBitmask{Low: 0, High: 0}, mask)
	})

	t.Run("keys in low range only", func(t *testing.T) {
		keys := []model.KeyPosition{
			model.KeyPosition(1),
			model.KeyPosition(5),
			model.KeyPosition(63),
		}
		mask := db.ComboKeyID(keys)

		// Set bits at positions 1, 5, and 63
		expected := db.ComboBitmask{
			Low:  uint64(1<<1) | uint64(1<<5) | uint64(1<<63),
			High: 0,
		}
		assert.Equal(t, expected, mask)
	})

	t.Run("keys in high range only", func(t *testing.T) {
		keys := []model.KeyPosition{
			model.KeyPosition(64),
			model.KeyPosition(70),
			model.KeyPosition(127),
		}
		mask := db.ComboKeyID(keys)

		// Set bits at positions 0, 6, and 63 in high bitmask (after modulo)
		expected := db.ComboBitmask{
			Low:  0,
			High: uint64(1<<0) | uint64(1<<6) | uint64(1<<63),
		}
		assert.Equal(t, expected, mask)
	})

	t.Run("keys in both low and high ranges", func(t *testing.T) {
		keys := []model.KeyPosition{
			model.KeyPosition(3),
			model.KeyPosition(64),
			model.KeyPosition(10),
			model.KeyPosition(72),
		}
		mask := db.ComboKeyID(keys)

		expected := db.ComboBitmask{
			Low:  (1 << 3) | (1 << 10),
			High: (1 << 0) | (1 << 8), // 64%64=0, 72%64=8
		}
		assert.Equal(t, expected, mask)
	})

	t.Run("same keys should produce same bitmask", func(t *testing.T) {
		keys1 := []model.KeyPosition{
			model.KeyPosition(5),
			model.KeyPosition(20),
			model.KeyPosition(70),
		}
		keys2 := []model.KeyPosition{
			model.KeyPosition(70),
			model.KeyPosition(5),
			model.KeyPosition(20),
		}

		mask1 := db.ComboKeyID(keys1)
		mask2 := db.ComboKeyID(keys2)

		assert.Equal(t, mask1, mask2)
	})

	t.Run("different keys should produce different bitmasks", func(t *testing.T) {
		keys1 := []model.KeyPosition{
			model.KeyPosition(5),
			model.KeyPosition(20),
		}
		keys2 := []model.KeyPosition{
			model.KeyPosition(5),
			model.KeyPosition(21),
		}

		mask1 := db.ComboKeyID(keys1)
		mask2 := db.ComboKeyID(keys2)

		assert.NotEqual(t, mask1, mask2)
	})
}

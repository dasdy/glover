package db

import (
	"cmp"
	"database/sql"
	"fmt"
	"glover/keylog/parser"
	"log"
	"slices"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	Store(event *parser.KeyEvent) error
	GatherAll() ([]MinimalKeyEvent, error)
	GatherCombos(length int) ([]Combo, error)
	Close()
}

type SQLiteStorage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) SQLiteStorage {
	return SQLiteStorage{db}
}

func InitDbStorage(db *sql.DB) error {
	// TODO: add indices over row-col-position?
	sqlStmt := `
	create table if not exists keypresses(row int, col int, position int, pressed bool, ts datetime);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %v\n", err, sqlStmt)
		return err
	}
	return nil
}

func ConnectDB(path string) (Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = InitDbStorage(db)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db}, nil
}

func (s *SQLiteStorage) Store(event *parser.KeyEvent) error {
	_, err := s.db.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, datetime('now'))`,
		event.Row, event.Col, event.Position, event.Pressed)
	if err != nil {
		return err
	}
	return nil
}

type MinimalKeyEvent struct {
	Row, Col, Position, Count int
}

func (s *SQLiteStorage) GatherAll() ([]MinimalKeyEvent, error) {
	// TODO: position should be same for each row-col, in reality, maybe groupby can be simpler. But double-check that.
	rows, err := s.db.Query(
		`select row, col, position, count(*) as cnt
        from keypresses
        where pressed = false
        group by row, col, position
        order by row, position`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]MinimalKeyEvent, 0)

	for rows.Next() {
		var row, col, position, count int

		err = rows.Scan(&row, &col, &position, &count)
		if err != nil {
			return nil, err
		}

		result = append(result, MinimalKeyEvent{Row: row, Col: col, Position: position, Count: count})
	}

	return result, nil
}

type ComboKey struct {
	Row, Col, Position int
}

type keyState struct {
	pressed  bool
	timeWhen time.Time
}

type Combo struct {
	Keys    []ComboKey
	Pressed int
}

func (s *SQLiteStorage) GatherCombos(length int) ([]Combo, error) {
	rows, err := s.db.Query(
		`select row, col, position, pressed, ts 
        from keypresses
        order by ts`)
	if err != nil {
		return nil, err
	}

	comboKeys := make([][]ComboKey, 0)

	defer rows.Close()

	keys := make(map[int]ComboKey)
	// position -> "pressed"
	curState := make(map[int]*keyState)

	for rows.Next() {
		var row, col, position int
		var pressed bool
		var ts time.Time

		err = rows.Scan(&row, &col, &position, &pressed, &ts)
		if err != nil {
			return nil, err
		}
		key := ComboKey{Row: row, Col: col, Position: position}
		keys[position] = key

		curState[position] = &keyState{pressed, ts}

		pressedKeys := make([]ComboKey, 0)
		for k, p := range curState {
			// Ignore key states that have been "true" for too long - for cases when keypress was kost
			if p.pressed && ts.Sub(p.timeWhen) > 2*time.Second {
				p.pressed = false
			}

			if p.pressed {
				pressedKeys = append(pressedKeys, keys[k])
			}
		}

		log.Printf("Current key: %v; pressedKeys: %+v", key, pressedKeys)

		if len(pressedKeys) >= length {
			comboKeys = append(comboKeys, pressedKeys)
		}
	}

	return countCombos(comboKeys), nil
}

func countCombos(keys [][]ComboKey) []Combo {
	counter := make(map[string]Combo)

	for _, combo := range keys {
		id := comboKeyId(combo)

		v, ok := counter[id]
		if !ok {
			v = Combo{Keys: combo, Pressed: 1}
		} else {
			v.Pressed++
		}
		counter[id] = v
	}

	result := make([]Combo, 0)
	for _, v := range counter {
		result = append(result, v)
	}

	slices.SortFunc(result, func(a, b Combo) int {
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
			keyCmp := cmp.Or(
				cmp.Compare(ak.Row, bk.Row),
				cmp.Compare(ak.Col, bk.Col),
				cmp.Compare(ak.Position, bk.Position),
			)
			if keyCmp != 0 {
				return keyCmp
			}
		}
		return 0
	})

	return result
}

func comboKeyId(keys []ComboKey) string {
	slices.SortFunc(keys, func(a, b ComboKey) int {
		return cmp.Or(
			cmp.Compare(a.Position, b.Position),
			cmp.Compare(a.Row, b.Row),
			cmp.Compare(a.Col, b.Col),
		)
	})

	res := strings.Builder{}

	for _, key := range keys {
		res.WriteString(fmt.Sprintf("(%d|%d|%d)", key.Row, key.Col, key.Position))
	}

	return res.String()
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

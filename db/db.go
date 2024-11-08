package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/dasdy/glover/model"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	Store(event *model.KeyEvent) error
	GatherAll() ([]model.MinimalKeyEvent, error)
	GatherCombos(length int) ([]model.Combo, error)
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
	create table if not exists keypresses(row int, col int, position int, pressed bool, ts datetime);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %v\n", err, sqlStmt)
		return err
	}

	sqlStmt = ` create index if not exists keypresses_tsix on keypresses (ts ASC);`
	_, err = db.Exec(sqlStmt)
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

func (s *SQLiteStorage) Store(event *model.KeyEvent) error {
	_, err := s.db.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, datetime('now', 'subsec'))`,
		event.Row, event.Col, event.Position, event.Pressed)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) GatherAll() ([]model.MinimalKeyEvent, error) {
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

	result := make([]model.MinimalKeyEvent, 0)

	for rows.Next() {
		var row, col, position, count int

		err = rows.Scan(&row, &col, &position, &count)
		if err != nil {
			return nil, err
		}

		result = append(result, model.MinimalKeyEvent{Row: row, Col: col, Position: position, Count: count})
	}

	return result, nil
}

type keyState struct {
	pressed  bool
	timeWhen time.Time
}

func (s *SQLiteStorage) GatherCombos(length int) ([]model.Combo, error) {
	rows, err := s.db.Query(
		`select position, pressed, ts 
        from keypresses
        order by ts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ScanForCombos(rows, length)
}

func ScanForCombos(cursor *sql.Rows, length int) ([]model.Combo, error) {
	counter := make(map[keyHash]*model.Combo)
	keys := make([]*model.ComboKey, 100)
	curState := make([]*keyState, 100)

	for cursor.Next() {
		var position int
		var pressed bool
		var ts time.Time

		err := cursor.Scan(&position, &pressed, &ts)
		if err != nil {
			return nil, err
		}

		if keys[position] == nil {
			key := model.ComboKey{Position: position}
			keys[position] = &key
		}

		curState[position] = &keyState{pressed, ts}

		pressedKeys := make([]model.ComboKey, 0)
		for k, p := range curState {
			if p == nil {
				continue
			}
			// Ignore key states that have been "true" for too long - for cases when keypress was kost
			if p.pressed && ts.Sub(p.timeWhen) > 2*time.Second {
				p.pressed = false
			}

			if p.pressed {
				pressedKeys = append(pressedKeys, *keys[k])
			}
		}

		if len(pressedKeys) >= length {
			id := comboKeyIdFast(pressedKeys)
			v, ok := counter[id]
			if !ok {
				counter[id] = &model.Combo{Keys: pressedKeys, Pressed: 1}
			} else {
				v.Pressed++
			}
		}
	}

	result := make([]model.Combo, 0, len(counter))
	for _, v := range counter {
		result = append(result, *v)
	}
	return result, nil
}

type keyHash struct {
	high int32
	low  int64
}

func comboKeyIdFast(keys []model.ComboKey) keyHash {
	result := keyHash{}

	for _, key := range keys {
		if key.Position < 64 {
			result.low |= (1 << key.Position)
		} else {
			position := key.Position % 64
			result.high |= (1 << position)
		}
	}

	return result
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

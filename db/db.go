package db

import (
	"database/sql"
	"iter"
	"log"
	"time"

	"github.com/dasdy/glover/model"
	// This registers sqlite3 as sql connection provider.
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type SQLiteStorage struct {
	db      *sql.DB
	verbose bool
}

func NewStorageFromConnection(db *sql.DB, verbose bool) (*SQLiteStorage, error) {
	// TODO: replace verbosity thing by structured logging config
	return &SQLiteStorage{db: db, verbose: verbose}, nil
}

// Given a path to storage, connect to it and initialize everything.
func NewStorageFromPath(path string, verbose bool) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	err = InitDBStorage(db)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db, verbose: verbose}, nil
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SQLiteStorage) AllIterator() (iter.Seq[model.KeyEventWithTimestamp], error) {
	rows, err := s.db.Query("select row, col, position, pressed, ts from keypresses order by ts")
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return func(yield func(model.KeyEventWithTimestamp) bool) {
		defer rows.Close()

		for rows.Next() {
			var row, col, position int

			var ts time.Time

			var pressed bool

			err = rows.Scan(&row, &col, &position, &pressed, &ts)

			item := model.KeyEventWithTimestamp{
				Row:       row,
				Col:       col,
				Position:  position,
				Pressed:   pressed,
				Timestamp: ts,
			}

			if !yield(item) {
				return
			}
		}
	}, nil
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

// count total amount of events in the db.
func (s *SQLiteStorage) count() (int, error) {
	rows, err := s.db.Query("select count(*) from keypresses")
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var count int

	rows.Next()

	if err := rows.Scan(&count); err != nil {
		return -1, err
	}

	if err = rows.Err(); err != nil {
		return -1, err
	}

	return count, nil
}

// Given a connection to db, set up needed tables and indices.
func InitDBStorage(db *sql.DB) error {
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

func Merge(inputs []*SQLiteStorage, out *SQLiteStorage) error {
	for i, input := range inputs {
		count, err := input.count()
		if err != nil {
			return err
		}

		rows, err := input.db.Query(`
            select row, col, position, pressed, ts from keypresses
        `)
		if err != nil {
			return err
		}
		defer rows.Close()

		log.Printf("processing input %d", i)

		bar := progressbar.Default(int64(count), "Writing...")

		for rows.Next() {
			err := bar.Add(1)
			if err != nil {
				return err
			}

			var (
				row, col, position int
				pressed            bool
				ts                 time.Time
			)

			err = rows.Scan(&row, &col, &position, &pressed, &ts)
			if err != nil {
				return err
			}

			_, err = out.db.Exec(`
                insert into keypresses(row, col, position, pressed, ts)
	            values(?, ?, ?, ?, ?)`,
				row, col, position, pressed, ts)
			if err != nil {
				return err
			}
		}

		if err = rows.Err(); err != nil {
			return err
		}
	}

	return nil
}

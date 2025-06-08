package db

import (
	"database/sql"
	"fmt"
	"iter"
	"log"
	"log/slog"
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

		return nil, fmt.Errorf("could not open path %s: got %w", path, err)
	}

	err = InitDBStorage(db)
	if err != nil {
		return nil, fmt.Errorf("could not initialize db storage: got %w", err)
	}

	return &SQLiteStorage{db: db, verbose: verbose}, nil
}

func (s *SQLiteStorage) Store(event *model.KeyEvent) error {
	_, err := s.db.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, datetime('now', 'subsec'))`,
		event.Row, event.Col, event.Position, event.Pressed)
	if err != nil {
		return fmt.Errorf("could not insert keypress %+v: got %w", event, err)
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
		return nil, fmt.Errorf("could not query keypresses: got %w", err)
	}

	defer rows.Close()

	result := make([]model.MinimalKeyEvent, 0)

	for rows.Next() {
		var row, col, position, count int

		err = rows.Scan(&row, &col, &position, &count)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: got %w", err)
		}

		result = append(result, model.MinimalKeyEvent{Row: row, Col: col, Position: model.KeyPosition(position), Count: count})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over rows: got %w", err)
	}

	return result, nil
}

func (s *SQLiteStorage) AllIterator() (iter.Seq[model.KeyEventWithTimestamp], error) {
	rows, err := s.db.Query("select row, col, position, pressed, ts from keypresses order by ts")
	if err != nil {
		return nil, fmt.Errorf("could not query keypresses: got %w", err)
	}

	// I'm not convinced that this actually does anything - the check should be done after the loop finishes(?)
	// If so, how do I return such error?
	if rows.Err() != nil {
		return nil, fmt.Errorf("error from getting rows: got %w", rows.Err())
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
				Position:  model.KeyPosition(position),
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
		return -1, fmt.Errorf("could not query keypresses count: got %w", err)
	}
	defer rows.Close()

	var count int

	rows.Next()

	if err := rows.Scan(&count); err != nil {
		return -1, fmt.Errorf("could not scan keypresses count: got %w", err)
	}

	if err = rows.Err(); err != nil {
		return -1, fmt.Errorf("error while iterating over keycount iterator: got %w", err)
	}

	return count, nil
}

// Given a connection to db, set up needed tables and indices.
func InitDBStorage(db *sql.DB) error {
	// TODO: add indices over row-col-position?
	sqlStmt := `create table if not exists keypresses(row int, col int, position int, pressed bool, ts datetime);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		slog.Error("failed to create table", "error", err, "sql", sqlStmt)

		return fmt.Errorf("could not create keypresses table: got %w", err)
	}

	sqlStmt = `create index if not exists keypresses_tsix on keypresses (ts ASC);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		slog.Error("failed to create index", "error", err, "sql", sqlStmt)

		return fmt.Errorf("could not create keypresses_tsix index: got %w", err)
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
			return fmt.Errorf("could not query keypresses from input %d: got %w", i, err)
		}
		defer rows.Close()

		slog.Info("processing input database", "index", i)

		bar := progressbar.Default(int64(count), "Writing...")

		for rows.Next() {
			err := bar.Add(1)
			if err != nil {
				return fmt.Errorf("could not update progress bar: got %w", err)
			}

			var (
				row, col, position int
				pressed            bool
				ts                 time.Time
			)

			err = rows.Scan(&row, &col, &position, &pressed, &ts)
			if err != nil {
				return fmt.Errorf("could not scan row from input %d: got %w", i, err)
			}

			_, err = out.db.Exec(`
                insert into keypresses(row, col, position, pressed, ts)
	            values(?, ?, ?, ?, ?)`,
				row, col, position, pressed, ts)
			if err != nil {
				return fmt.Errorf("could not insert keypress from input %d: got %w", i, err)
			}
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error while iterating over rows from input %d: got %w", i, err)
		}
	}

	return nil
}

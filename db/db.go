package db

import (
	"database/sql"
	"glover/keylog/parser"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	Store(event *parser.KeyEvent) error
	GatherAll() ([]MinimalKeyEvent, error)
	Close()
}

type SQLiteStorage struct {
	db *sql.DB
}

func ConnectDB(path string) (Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// TODO: add indices over row-col-position?
	sqlStmt := `
	create table if not exists keypresses(row int, col int, position int, pressed bool, ts datetime);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %v\n", err, sqlStmt)
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

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

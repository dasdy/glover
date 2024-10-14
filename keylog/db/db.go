package db

import (
	"database/sql"
	"glover/keylog/parser"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	Store(event *parser.KeyEvent) error
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

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

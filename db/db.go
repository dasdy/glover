package db

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/dasdy/glover/model"
	// This registers sqlite3 as sql connection provider.
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type Storage interface {
	Store(event *model.KeyEvent) error
	GatherAll() ([]model.MinimalKeyEvent, error)
	GatherCombos() []model.Combo
	Close()
}

type SQLiteStorage struct {
	db           *sql.DB
	comboTracker *ComboTracker
}

func NewStorage(db *sql.DB) (*SQLiteStorage, error) {
	tracker, err := NewComboTrackerFromDb(db)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db, comboTracker: tracker}, nil
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

func ConnectDB(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	err = InitDbStorage(db)
	if err != nil {
		return nil, err
	}

	tracker, err := NewComboTrackerFromDb(db)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db, tracker}, nil
}

func (s *SQLiteStorage) Store(event *model.KeyEvent) error {
	_, err := s.db.Exec(`insert into keypresses(row, col, position, pressed, ts)
	    values(?, ?, ?, ?, datetime('now', 'subsec'))`,
		event.Row, event.Col, event.Position, event.Pressed)
	if err != nil {
		return err
	}

	s.comboTracker.HandleKeyNow(event.Position, event.Pressed)

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
	timeWhen time.Time
	pressed  bool
}

type ComboTracker struct {
	comboCounts map[keyHash]*model.Combo
	curState    []*keyState
	keys        []*model.ComboKey
	minComboLen int
	l           sync.RWMutex
}

func newComboTracker(keyCount, minComboLen int) *ComboTracker {
	return &ComboTracker{
		comboCounts: make(map[keyHash]*model.Combo),
		curState:    make([]*keyState, keyCount),
		keys:        make([]*model.ComboKey, keyCount),
		minComboLen: minComboLen,

		l: sync.RWMutex{},
	}
}

func NewComboTrackerFromDb(db *sql.DB) (*ComboTracker, error) {
	nullTracker := newComboTracker(100, 2)

	err := nullTracker.initComboCounter(db)
	if err != nil {
		return nil, err
	}

	return nullTracker, nil
}

func (c *ComboTracker) HandleKey(position int, pressed bool, timeWhen time.Time) {
	c.l.Lock()
	defer c.l.Unlock()

	if c.keys[position] == nil {
		key := model.ComboKey{Position: position}
		c.keys[position] = &key
	}

	c.curState[position] = &keyState{pressed: pressed, timeWhen: timeWhen}

	pressedKeys := make([]model.ComboKey, 0)

	for k, p := range c.curState {
		if p == nil {
			continue
		}
		// Ignore key states that have been "true" for too long - for cases when keypress was kost
		if p.pressed && timeWhen.Sub(p.timeWhen) > 2*time.Second {
			p.pressed = false
		}

		if p.pressed {
			pressedKeys = append(pressedKeys, *c.keys[k])
		}
	}

	if len(pressedKeys) >= c.minComboLen {
		id := comboKeyID(pressedKeys)

		v, ok := c.comboCounts[id]
		if !ok {
			c.comboCounts[id] = &model.Combo{Keys: pressedKeys, Pressed: 1}
		} else {
			v.Pressed++
		}
	}
}

func (c *ComboTracker) HandleKeyNow(position int, pressed bool) {
	c.HandleKey(position, pressed, time.Now())
}

func (c *ComboTracker) GatherCombos() []model.Combo {
	c.l.RLock()
	defer c.l.RUnlock()

	result := make([]model.Combo, 0, len(c.comboCounts))
	for _, v := range c.comboCounts {
		result = append(result, *v)
	}

	return result
}

func (s *SQLiteStorage) GatherCombos() []model.Combo {
	return s.comboTracker.GatherCombos()
}

func (c *ComboTracker) initComboCounter(db *sql.DB) error {
	rows, err := db.Query(
		`select position, pressed, ts 
        from keypresses
        order by ts`)
	if err != nil {
		return err
	}
	defer rows.Close()

	return c.initialRead(rows)
}

func (c *ComboTracker) initialRead(cursor *sql.Rows) error {
	for cursor.Next() {
		var (
			position int
			pressed  bool
			ts       time.Time
		)

		if err := cursor.Scan(&position, &pressed, &ts); err != nil {
			return err
		}

		c.HandleKey(position, pressed, ts)
	}

	return nil
}

type keyHash struct {
	high int32
	low  int64
}

func comboKeyID(keys []model.ComboKey) keyHash {
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

func (s *SQLiteStorage) count() (int, error) {
	rows, err := s.db.Query("select count(*) from keypresses")
	if err != nil {
		return -1, err
	}

	var count int

	rows.Next()

	if err := rows.Scan(&count); err != nil {
		return -1, err
	}

	return count, nil
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
			bar.Add(1)

			var (
				row, col, position int
				pressed            bool
				ts                 time.Time
			)

			err = rows.Scan(&row, &col, &position, &pressed, &ts)
			if err != nil {
				return err
			}

			_, err := out.db.Exec(`
                insert into keypresses(row, col, position, pressed, ts)
	            values(?, ?, ?, ?, ?)`,
				row, col, position, pressed, ts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

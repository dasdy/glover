package db

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/dasdy/glover/model"
)

type keyState struct {
	timeWhen time.Time
	pressed  bool
}

type ComboTracker struct {
	comboCounts map[keyHash]*model.Combo
	curState    []*keyState
	keys        []*model.ComboKey
	minComboLen int
	stateLock   sync.RWMutex
}

func newComboTracker(keyCount, minComboLen int) *ComboTracker {
	return &ComboTracker{
		comboCounts: make(map[keyHash]*model.Combo),
		curState:    make([]*keyState, keyCount),
		keys:        make([]*model.ComboKey, keyCount),
		minComboLen: minComboLen,

		stateLock: sync.RWMutex{},
	}
}

func NewComboTrackerFromDB(db *sql.DB) (*ComboTracker, error) {
	nullTracker := newComboTracker(100, 2)

	err := nullTracker.initComboCounter(db)
	if err != nil {
		return nil, err
	}

	return nullTracker, nil
}

func (c *ComboTracker) HandleKeyNow(position int, pressed bool, verbose bool) {
	c.handleKey(position, pressed, time.Now(), verbose)
}

func (c *ComboTracker) GatherCombos() []model.Combo {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()

	result := make([]model.Combo, 0, len(c.comboCounts))
	for _, v := range c.comboCounts {
		result = append(result, *v)
	}

	return result
}

func (c *ComboTracker) handleKey(position int, pressed bool, timeWhen time.Time, verbose bool) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

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
		if p.pressed && timeWhen.Sub(p.timeWhen) > 10*time.Second {
			log.Printf("Ignoring key in position %d: pressed %v ago", k, timeWhen.Sub(p.timeWhen))
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
			v = &model.Combo{Keys: pressedKeys, Pressed: 1}
			c.comboCounts[id] = v
		} else {
			v.Pressed++
		}

		if verbose {
			log.Printf("combo counting (%d keys, pressed: %d): %+v", len(pressedKeys), v.Pressed, pressedKeys)
		}
	}
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

	for rows.Next() {
		var (
			position int
			pressed  bool
			ts       time.Time
		)

		if err := rows.Scan(&position, &pressed, &ts); err != nil {
			return err
		}

		c.handleKey(position, pressed, ts, false)
	}

	if err := rows.Err(); err != nil {
		return err
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

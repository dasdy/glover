package db

import (
	"iter"
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

func NewComboTrackerFromDB(storage Storage) (*ComboTracker, error) {
	tracker := newComboTracker(100, 2)

	// TODO: make the init stage a bit clearer? forbid combos page calls until this init is done?
	go func() {
		iterator, err := storage.AllIterator()
		if err != nil {
			panic(err)
		}

		tracker.initComboCounter(iterator)
	}()

	return tracker, nil
}

func (c *ComboTracker) HandleKeyNow(position int, pressed bool, verbose bool) {
	c.handleKey(position, pressed, time.Now(), verbose)
}

func (c *ComboTracker) GatherCombos(position int) []model.Combo {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()

	result := make([]model.Combo, 0, len(c.comboCounts))

	for _, v := range c.comboCounts {
		for _, k := range v.Keys {
			if k.Position == position {
				result = append(result, *v)

				break
			}
		}
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

func (c *ComboTracker) initComboCounter(items iter.Seq[model.KeyEventWithTimestamp]) {
	for item := range items {
		c.handleKey(item.Position, item.Pressed, item.Timestamp, false)
	}
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

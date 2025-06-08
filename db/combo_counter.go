package db

import (
	"iter"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/dasdy/glover/model"
	"github.com/schollz/progressbar/v3"
)

type keyState struct {
	timeWhen time.Time
	pressed  bool
}

type ComboTracker struct {
	comboCounts map[ComboBitmask]*model.Combo
	curState    []*keyState
	keys        []*model.KeyPosition
	minComboLen int
	stateLock   sync.RWMutex
}

func newComboTracker(keyCount, minComboLen int) *ComboTracker {
	return &ComboTracker{
		comboCounts: make(map[ComboBitmask]*model.Combo),
		curState:    make([]*keyState, keyCount),
		keys:        make([]*model.KeyPosition, keyCount),
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

func (c *ComboTracker) HandleKeyNow(position model.KeyPosition, pressed bool, verbose bool) {
	c.handleKey(position, pressed, time.Now(), verbose)
}

func (c *ComboTracker) GatherCombos(position model.KeyPosition) []model.Combo {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()

	result := make([]model.Combo, 0, len(c.comboCounts))

	for _, v := range c.comboCounts {
		if slices.Contains(v.Keys, position) {
			result = append(result, *v)
		}
	}

	return result
}

func (c *ComboTracker) handleKey(position model.KeyPosition, pressed bool, timeWhen time.Time, verbose bool) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	if c.keys[position] == nil {
		key := position
		c.keys[position] = &key
	}

	c.curState[position] = &keyState{pressed: pressed, timeWhen: timeWhen}

	pressedKeys := make([]model.KeyPosition, 0)

	for k, p := range c.curState {
		if p == nil {
			continue
		}
		// Ignore key states that have been "true" for too long - for cases when keypress was kost
		if p.pressed && timeWhen.Sub(p.timeWhen) > 10*time.Second {
			if verbose {
				slog.Info("ignoring stale key",
					"position", k,
					"staleness", timeWhen.Sub(p.timeWhen))
			}

			p.pressed = false
		}

		if p.pressed {
			pressedKeys = append(pressedKeys, *c.keys[k])
		}
	}

	if len(pressedKeys) >= c.minComboLen {
		id := ComboKeyID(pressedKeys)

		v, ok := c.comboCounts[id]
		if !ok {
			v = &model.Combo{Keys: pressedKeys, Pressed: 1}
			c.comboCounts[id] = v
		} else {
			v.Pressed++
		}

		if verbose {
			slog.Info("combo counting",
				"keyCount", len(pressedKeys),
				"pressed", v.Pressed,
				"keys", pressedKeys)
		}
	}
}

func (c *ComboTracker) initComboCounter(items iter.Seq[model.KeyEventWithTimestamp]) {
	bar := progressbar.Default(-1, "Scanning history...")
	for item := range items {
		err := bar.Add(1)
		if err != nil {
			slog.Error("could not update progress bar", "error", err)
		}

		c.handleKey(item.Position, item.Pressed, item.Timestamp, false)
	}

	err := bar.Finish()
	if err != nil {
		slog.Error("could not finish progress bar", "error", err)
	}
}

type ComboBitmask struct {
	High uint64
	Low  uint64
}

// Represent combo by a bitmask. Each key present in the combo
// will have its' bit set to 1 in the mask. Key.position is used for that.
// Assert that Keyboard has at most 128 keys because I don't really care
// about keyboards larger than that.
func ComboKeyID(keys []model.KeyPosition) ComboBitmask {
	result := ComboBitmask{}

	for _, key := range keys {
		intKey := int(key)
		if intKey < 64 {
			result.Low |= (1 << intKey)
		} else {
			position := intKey % 64
			result.High |= (1 << position)
		}
	}

	return result
}

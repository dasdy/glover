package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dasdy/glover/model"
)

var ErrEmptyLine = errors.New("line does not contain something we can parse as KeyEvent")

func ParseLine(line string) (*model.KeyEvent, error) {
	splits := strings.Split(line, " ")

	var (
		row, col, position, foundCount int
		pressed                        bool
		err                            error
	)

	ix := 0
	limit := len(splits) - 1 // We always care about the next token, so stop before it's too late

	for ix < limit {
		curItem := splits[ix]
		nextItem := strings.TrimRight(splits[ix+1], ",")

		switch curItem {
		case "Row:":
			row, err = strconv.Atoi(nextItem)
			if err != nil {
				return nil, fmt.Errorf("could not parse row: %w. Full line: '%s'", err, line)
			}

			ix++
			foundCount++

		case "col:":
			col, err = strconv.Atoi(nextItem)
			if err != nil {
				return nil, fmt.Errorf("could not parse col: %w. Full line: '%s'", err, line)
			}

			ix++
			foundCount++

		case "position:":
			position, err = strconv.Atoi(nextItem)
			if err != nil {
				return nil, fmt.Errorf("could not parse position: %w. Full line: '%s'", err, line)
			}

			foundCount++
			ix++

		case "pressed:":
			// Trim the reset escape code from the output. Maybe we can do it another way implicitly?
			nextItem = strings.TrimSuffix(nextItem, "\x1b[0m")
			switch nextItem {
			case "true":
				pressed = true
			case "false":
				pressed = false
			default:
				return nil, fmt.Errorf("pressed value unexpected: '%s'", nextItem)
			}

			ix++
			foundCount++
		default:
		}

		ix++
	}

	if foundCount == 4 {
		return &model.KeyEvent{Row: row, Col: col, Position: model.KeyPosition(position), Pressed: pressed}, nil
	}

	return nil, ErrEmptyLine
}

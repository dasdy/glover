package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type KeyEvent struct {
	Row      int
	Col      int
	Position int
	Pressed  bool
}

func ParseLine(line string) (*KeyEvent, error) {
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
		return &KeyEvent{row, col, position, pressed}, nil
	}
	return nil, nil
}

var pattern = regexp.MustCompile(`Row: (?P<row>\d+), col: (?P<col>\d+), position: (?P<pos>\d+), pressed: (?P<pres>true|false)`)

func ParseLineRegex(line string) (*KeyEvent, error) {
	matches := pattern.FindStringSubmatch(line)

	if matches == nil {
		return nil, nil
	}

	row, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}
	col, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}
	pos, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, err
	}
	pressed := matches[4] == "true"

	return &KeyEvent{row, col, pos, pressed}, nil
}

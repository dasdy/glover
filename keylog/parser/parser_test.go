package parser_test

import (
	"fmt"
	"glover/keylog/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseLineTest struct {
	name           string
	line           string
	expectedResult *parser.KeyEvent
}

func TestParseLine(t *testing.T) {
	testCases := []parseLineTest{
		{"empty", "", nil},
		{
			"correct full line",
			`[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false`,
			&parser.KeyEvent{2, 1, 23, false},
		},
		{
			"trims escape code at end",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false\x1b[0m",
			&parser.KeyEvent{2, 1, 23, false},
		},
	}

	for _, item := range testCases {
		t.Run(fmt.Sprintf("parses %s", item.name), func(t *testing.T) {
			res, err := parser.ParseLine(item.line)

			assert.NoError(t, err)

			assert.Equal(t, res, item.expectedResult)
		})

		t.Run(fmt.Sprintf("regex parses %s", item.name), func(t *testing.T) {
			res, err := parser.ParseLineRegex(item.line)

			assert.NoError(t, err)

			assert.Equal(t, res, item.expectedResult)
		})

	}
}

var result *parser.KeyEvent

func BenchmarkParseLine(b *testing.B) {
	line := "[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false\x1b[0m"
	var r *parser.KeyEvent
	for i := 0; i < b.N; i++ {
		r, _ = parser.ParseLine(line)
	}

	result = r
}

func BenchmarkParseLineRegex(b *testing.B) {
	line := "[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false\x1b[0m"
	var r *parser.KeyEvent
	for i := 0; i < b.N; i++ {
		r, _ = parser.ParseLineRegex(line)
	}

	result = r
}

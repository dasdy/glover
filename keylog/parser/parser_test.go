package parser_test

import (
	"testing"

	"github.com/dasdy/glover/keylog/parser"
	"github.com/dasdy/glover/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parseLineTest struct {
	name           string
	line           string
	expectedResult *model.KeyEvent
}
type errorLineTest struct {
	name string
	line string
}

func TestParseLine(t *testing.T) {
	testCases := []parseLineTest{
		{
			"correct full line",
			`[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false`,
			&model.KeyEvent{Row: 2, Col: 1, Position: 23, Pressed: false},
		},
		{
			"trims escape code at end",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false\x1b[0m",
			&model.KeyEvent{Row: 2, Col: 1, Position: 23, Pressed: false},
		},
		{
			"pressed=true",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: true",
			&model.KeyEvent{Row: 2, Col: 1, Position: 23, Pressed: true},
		},
	}

	for _, item := range testCases {
		t.Run("parses "+item.name, func(t *testing.T) {
			res, err := parser.ParseLine(item.line)

			require.NoError(t, err)

			assert.Equal(t, item.expectedResult, res)
		})
	}

	errorTestCases := []errorLineTest{
		{"empty", ""},
		{
			"pressed=gobble",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: t",
		},
		{
			"row malformed",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: , col: 1, position: 23, pressed: true",
		},
		{
			"col malformed",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: k, position: 23, pressed: true",
		},
		{
			"pos malformed",
			"[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: :, pressed: true",
		},
	}

	for _, item := range errorTestCases {
		t.Run("does not parse "+item.name, func(t *testing.T) {
			res, err := parser.ParseLine(item.line)

			require.Error(t, err)
			assert.Nil(t, res)
		})
	}
}

var result *model.KeyEvent

func BenchmarkParseLine(b *testing.B) {
	line := "[23:09:36.886,444] <dbg> zmk: zmk_kscan_process_msgq: Row: 2, col: 1, position: 23, pressed: false\x1b[0m"

	var r *model.KeyEvent

	for range b.N {
		r, _ = parser.ParseLine(line)
	}

	result = r
}

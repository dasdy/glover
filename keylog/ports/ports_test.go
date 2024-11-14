package ports_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/dasdy/glover/keylog/ports"
	"github.com/stretchr/testify/assert"
)

func readChanLines(c <-chan string) []string {
	result := make([]string, 0)

	for line := range c {
		result = append(result, line)
	}
	return result
}

func TestReadFile(t *testing.T) {
	t.Run("should handle non-empty file", func(t *testing.T) {
		r := strings.NewReader("a\nb\nc\n")

		c := ports.ReadFile(r)

		lines := readChanLines(c)

		assert.Equal(t, []string{"a", "b", "c"}, lines)
	})

	t.Run("should handle empty file", func(t *testing.T) {
		r := strings.NewReader("")

		c := ports.ReadFile(r)

		lines := readChanLines(c)

		assert.Equal(t, []string{}, lines)
	})
}

func TestReadTwoFiles(t *testing.T) {
	t.Run("should handle non-empty files", func(t *testing.T) {
		r1 := strings.NewReader("aa\nbb\ncc\n")
		r2 := strings.NewReader("ab\nba\ncd\n")

		c := ports.ReadTwoFiles(r1, r2)

		lines := readChanLines(c)

		sort.Strings(lines)

		assert.Equal(t, []string{
			"aa", "ab", "ba", "bb", "cc", "cd",
		}, lines)
	})

	t.Run("should handle when one file is empty", func(t *testing.T) {
		r1 := strings.NewReader("aa\nbb\ncc\n")
		r2 := strings.NewReader("")

		c := ports.ReadTwoFiles(r1, r2)

		lines := readChanLines(c)

		sort.Strings(lines)

		assert.Equal(t, []string{
			"aa", "bb", "cc",
		}, lines)
	})
	t.Run("should handle when other file is empty", func(t *testing.T) {
		r1 := strings.NewReader("")
		r2 := strings.NewReader("aa\nbb\ncc\n")

		c := ports.ReadTwoFiles(r1, r2)

		lines := readChanLines(c)

		sort.Strings(lines)

		assert.Equal(t, []string{
			"aa", "bb", "cc",
		}, lines)
	})
}

package layout

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

func GetBinaryPath() string {
	// TODO: parameterize;
	//nolint:dogsled
	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	fp := filepath.Join(filepath.Dir(b), "..")

	return fp
}

func OpenPath(path string) (*os.File, error) {
	var err error

	var file *os.File

	if filepath.IsAbs(path) {
		slog.Info("Opening absolute path", "path", path)
		file, err = os.Open(path)
	} else {
		slog.Info("Opening relative path", "path", path)
		file, err = os.Open(filepath.Join(GetBinaryPath(), path))
	}

	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", path, err)
	}

	return file, nil
}

var labels = map[string]string{
	"LEFT_SHIFT":  "⇧",
	"LSHFT":       "⇧",
	"RIGHT_SHIFT": "R⇧",
	"RSHFT":       "R⇧",
	"LCTRL":       "^",
	"RCTRL":       "⌃",
	"RET":         "↵",
	"LCMD":        "⌘",
	"RCMD":        "⌘",
	"LALT":        "⌥",
	"RALT":        "⌥",
	"BSPC":        "⌫",
	"SPACE":       "␣",
	"TAB":         "⇥",

	"RIGHT_ARROW": "→",
	"RIGHT":       "→",
	"LEFT_ARROW":  "←",
	"LEFT":        "←",
	"UP_ARROW":    "↑",
	"DOWN_ARROW":  "↓",
	"EQUAL":       "=",
	"N1":          "1",
	"N2":          "2",
	"N3":          "3",
	"N4":          "4",
	"N5":          "5",
	"N6":          "6",
	"N7":          "7",
	"N8":          "8",
	"N9":          "9",
	"N0":          "0",
	"COMMA":       ",",
	"LBKT":        "[",
	"RBKT":        "]",
	"DOT":         ".",
	"SEMI":        ":",
	"BSLH":        "\\",
	"FSLH":        "/",
	"SQT":         "'",
	"MINUS":       "-",
	"GRAVE":       "`",

	// TODO: this is a hack, need to rethink combo parsing.
	"LS(LALT)": "⇧+⌥",
}

func GetKeyLabels(filename string) ([]string, error) {
	file, err := OpenPath(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open keymap file %s. %w", filename, err)
	}
	defer file.Close()

	keymap, err := Parse(file)
	if err != nil {
		return nil, fmt.Errorf("could not parse keymap file. %w", err)
	}

	if len(keymap.Layers) < 1 {
		return nil, errors.New("expected at least 1 layer in layout")
	}

	results := make([]string, 0, len(keymap.Layers[0].Bindings))

	for _, b := range keymap.Layers[0].Bindings {
		switch b.Action {
		case "&kp":
			for i := range b.Modifiers {
				if v, ok := labels[b.Modifiers[i]]; ok {
					b.Modifiers[i] = v
				}
			}

			if len(b.Modifiers) > 1 {
				results = append(results, fmt.Sprintf("%+v", b.Modifiers))
			} else {
				results = append(results, b.Modifiers[0])
			}
		case "&magic":
			results = append(results, "🪄")
		default:
			results = append(results, fmt.Sprintf("%s %+v", b.Action, b.Modifiers))
		}
	}

	return results, nil
}

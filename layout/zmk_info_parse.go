package layout

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/dasdy/glover/model"
)

type ZMKKeyDescriptor struct {
	Row   int     `json:"row"`
	Col   int     `json:"col"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	R     float64 `json:"r"`
	Rx    float64 `json:"rx"`
	Ry    float64 `json:"ry"`
	Label string  `json:"label"`
}

type ZMKLayoutCollection struct {
	Layout []ZMKKeyDescriptor `json:"layout"`
}

type ZmkInfoJSON struct {
	ID      string                         `json:"id"`
	Name    string                         `json:"name"`
	Layouts map[string]ZMKLayoutCollection `json:"layouts"`
}

func LoadZmkLocationsJSON(reader io.Reader) (*model.KeyboardLayout, error) {
	locations := make(map[int]model.Location)

	decoder := json.NewDecoder(reader)

	var info ZmkInfoJSON

	if err := decoder.Decode(&info); err != nil {
		return nil, fmt.Errorf("could not decode ZMK info JSON: %w", err)
	}

	if len(info.Layouts) != 1 {
		return nil, fmt.Errorf("expected exactly one layout, got %d", len(info.Layouts))
	}

	rows := 0
	cols := 0

	keyID := 0

	for _, layout := range info.Layouts {
		for _, key := range layout.Layout {
			loc := model.Location{}
			loc.Col = key.Col
			loc.Row = key.Row

			if key.Row > rows {
				rows = key.Row
			}

			if key.Col > cols {
				cols = key.Col
			}

			loc.X = key.X
			loc.Y = key.Y
			loc.R = key.R
			loc.Rx = key.Rx
			loc.Ry = key.Ry

			// log.Printf("Key %d: %+v", keyID, loc)
			// loc.Label = key.Label
			locations[keyID] = loc

			keyID++
		}
	}

	return &model.KeyboardLayout{
		Locations: locations,
		Rows:      rows + 1,
		Cols:      cols + 1,
	}, nil
}

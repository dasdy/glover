package model

import (
	"time"
)

// Position of the key in a combo.
type KeyPosition int

type KeyEvent struct {
	Row      int
	Col      int
	Position KeyPosition
	Pressed  bool
}

type KeyEventWithTimestamp struct {
	Row       int
	Col       int
	Position  KeyPosition
	Pressed   bool
	Timestamp time.Time
}

type MinimalKeyEvent struct {
	Row, Col, Count int
	Position        KeyPosition
}

type MinimalKeyEventWithLabel struct {
	Position KeyPosition
	Count    int
	KeyLabel string
	Location Location
}

type Combo struct {
	Keys    []KeyPosition
	Pressed int
}

type KeyboardLayout struct {
	Locations map[KeyPosition]Location
	Rows      int
	Cols      int
}

type RowCol struct {
	Row int
	Col int
}

type Location struct {
	RowCol
	X  float64
	Y  float64
	R  float64
	Rx float64
	Ry float64
}

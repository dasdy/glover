package model

import "time"

type KeyEvent struct {
	Row      int
	Col      int
	Position int
	Pressed  bool
}

type KeyEventWithTimestamp struct {
	Row       int
	Col       int
	Position  int
	Pressed   bool
	Timestamp time.Time
}

type MinimalKeyEvent struct {
	Row, Col, Position, Count int
}

type MinimalKeyEventWithLabel struct {
	Position, Count int
	KeyLabel        string
	Location        Location
}

type ComboKey struct {
	Position int
}

type Combo struct {
	Keys    []ComboKey
	Pressed int
}

type KeyboardLayout struct {
	Locations map[int]Location
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

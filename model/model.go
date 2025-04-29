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
	Row, Col, Position, Count int
	KeyLabel                  string
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

type Location struct {
	Row int
	Col int
}

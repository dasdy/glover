package model

type KeyEvent struct {
	Row      int
	Col      int
	Position int
	Pressed  bool
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

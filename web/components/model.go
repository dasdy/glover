package components

import "github.com/dasdy/glover/model"

type Item struct {
	Position       int
	Location       model.Location
	KeyName        string
	KeypressAmount string
	Highlight      bool
}

type ComboConnection struct {
	FromPosition int
	ToPosition   int
	PressCount   int
}
type PageType string

const (
	PageTypeStats     PageType = "stats"
	PageTypeCombo     PageType = "combo"
	PageTypeNeighbors PageType = "neighbors"
)

const (
	KeySize           = 80
	KeySizeWithoutGap = 70
	KeyCenterOffset   = KeySizeWithoutGap / 2
)

type RenderContext struct {
	TotalCols int
	TotalRows int
	Items     []Item
	MaxVal    int
	Page      PageType

	HighlightPosition int               // The position being highlighted
	ComboConnections  []ComboConnection // Top 5 combo connections for highlighted key
}

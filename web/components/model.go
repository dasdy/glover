package components

type Item struct {
	Position       int
	Row            int
	Col            int
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

type RenderContext struct {
	TotalCols int
	Items     []Item
	MaxVal    int
	Page      PageType

	HighlightPosition int               // The position being highlighted
	ComboConnections  []ComboConnection // Top 5 combo connections for highlighted key
}

type Location struct {
	Row int
	Col int
}

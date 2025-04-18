package components

type Item struct {
	Position       int
	Row            int
	Col            int
	KeyName        string
	KeypressAmount string
	Highlight      bool
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
}

type Location struct {
	Row int
	Col int
}

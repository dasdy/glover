package components

type Item struct {
	Position       int
	Row            int
	Col            int
	KeyName        string
	KeypressAmount string
	Highlight      bool
}

type RenderContext struct {
	TotalCols int
	Items     []Item
	MaxVal    int
}

type Location struct {
	Row int
	Col int
}

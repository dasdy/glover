package components

type Item struct {
	Position       int
	KeyName        string
	KeypressAmount string
	Visible        bool
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

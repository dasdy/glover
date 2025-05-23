package components

import (
	"fmt"
	"log"
	"math"

	"github.com/dasdy/glover/model"
)

// getLinkForPosition returns the appropriate URL based on the page type.
func getLinkForPosition(position int, pageType PageType) string {
	switch pageType {
	case PageTypeCombo, PageTypeStats:
		return fmt.Sprintf("/combo?position=%d", position)
	case PageTypeNeighbors:
		return fmt.Sprintf("/neighbors?position=%d", position)
	default:
		return fmt.Sprintf("/combo?position=%d", position)
	}
}

// getSwitchModeLink returns the appropriate URL to switch between combo and neighbors modes.
func getSwitchModeLink(position int, currentPageType PageType) string {
	switch currentPageType {
	case PageTypeCombo, PageTypeStats:
		return fmt.Sprintf("/neighbors?position=%d", position)
	case PageTypeNeighbors:
		return fmt.Sprintf("/combo?position=%d", position)
	default:
		return "/"
	}
}

// getSwitchModeButtonText returns the appropriate button text for switching modes.
func getSwitchModeButtonText(currentPageType PageType) string {
	switch currentPageType {
	case PageTypeCombo, PageTypeStats:
		return "View Neighbors"
	case PageTypeNeighbors:
		return "View Combos"
	default:
		return ""
	}
}

func (c *RenderContext) ViewBoxSize() string {
	maxX := float64(c.TotalCols)
	maxY := float64(c.TotalRows)

	for _, item := range c.Items {
		if item.Location.X > maxX {
			maxX = item.Location.X
		}

		if item.Location.Y > maxY {
			maxY = item.Location.Y
		}
	}

	// TODO: figure out how to account for keys with Rx/Ry properly.
	maxY += 2

	return fmt.Sprintf(
		"0 0 %d %d",
		int(math.Ceil(KeySize*(maxX))),
		int(math.Ceil(KeySize*(1+maxY))),
	)
}

func FindConnectionKey(c *RenderContext, conn *ComboConnection) (*Item, *Item) {
	var fromKey, toKey *Item
	for i := range c.Items {
		if c.Items[i].Position == conn.FromPosition {
			fromKey = &c.Items[i]
		}
		if c.Items[i].Position == conn.ToPosition {
			toKey = &c.Items[i]
		}
	}

	return fromKey, toKey
}

func KeyPath(fromKey, toKey *Item) string {
	fromX := fromKey.Location.Col*KeySize + KeyCenterOffset
	fromY := fromKey.Location.Row*KeySize + KeyCenterOffset
	toX := toKey.Location.Col*KeySize + KeyCenterOffset
	toY := toKey.Location.Row*KeySize + KeyCenterOffset

	// Calculate control points for a curved path
	midX := (fromX + toX) / 2
	midY := (fromY+toY)/2 - 40 // Curve upward

	return fmt.Sprintf("M %d %d Q %d %d %d %d", fromX, fromY, midX, midY, toX, toY)
}

func KeyPathStrokeWidth(conn *ComboConnection) string {
	// Draw a curved path with thickness based on press count
	// TODO: scaling by 100 is arbirary, we should adjust it based on total keystrokes in db
	// or something.
	strokeWidth := math.Min(10, 1+float64(conn.PressCount)/100)

	return fmt.Sprintf("%f", strokeWidth)
}

func ToTransform(l *model.Location) string {
	log.Printf("ToTransform: %+v", l)
	translate := fmt.Sprintf("translate(%.2f, %.2f)", l.X*KeySize, l.Y*KeySize)
	if l.R != 0 {
		translate += fmt.Sprintf(" rotate(%.2f)", l.R)
	}

	return translate
}

func ToTransformOrigin(l *model.Location) string {
	rx := l.Rx
	if l.Rx != 0 {
		rx -= l.X
	}

	ry := l.Ry
	if l.Ry != 0 {
		ry -= l.Y
	}

	return fmt.Sprintf("%.2f %.2f", rx*KeySize, ry*KeySize)
}

package components

import (
	"fmt"
	"log/slog"
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

// Calculate how big coordinate space needs to be to fit all keys.
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

// Attempt at a linear algebra. Seems to be correct, but for some reason the output is not what I expect.
// and svg points created by this function do not match the key boxes.
func RotatePoint(x, y, cx, cy, angle float64) (float64, float64) {
	// Convert angle from degrees to radians
	angleRad := angle * math.Pi / 180.0

	// Translate point to origin
	x -= cx
	y -= cy

	// Rotate point
	xNew := x*math.Cos(angleRad) - y*math.Sin(angleRad)
	yNew := x*math.Sin(angleRad) + y*math.Cos(angleRad)

	// Translate point back
	return xNew + cx, yNew + cy
}

// Try to calculate center of the key based on its location. Can be used to calculate paths.
func KeyCenter(key *Item) (float64, float64) {
	x := key.Location.X * KeySize // + KeyCenterOffset
	y := key.Location.Y * KeySize // + KeyCenterOffset

	// if key.Location.R != 0 {
	slog.Info("key center calculation",
		"key", key.KeyName,
		"location", key.Location,
		"x", x,
		"y", y,
		"rotation", key.Location.R)

	// Rotate the point if it has a rotation. Should work fine, but gives completely wrong results.
	// cx, cy := ToTransformOrigin(&key.Location)
	// log.Printf("KeyCenter (%s): %+v, x: %.2f, y: %.2f, cx: %.2f, cy: %.2f r: %.2f", key.KeyName, key.Location, x, y, cx, cy, key.Location.R)
	// x, y = RotatePoint(x, y, cx, cy, key.Location.R)

	// log.Printf("KeyCenter (%s): %+v, x: %.2f, y: %.2f, r: %.2f", key.KeyName, key.Location, x, y, key.Location.R)

	// Gives somewhat similar results, but I dunno how to make them fit exactly.
	x, y = RotatePoint(x, y, key.Location.Rx*KeySize, key.Location.Ry*KeySize, key.Location.R)
	// }

	slog.Info("key center after rotation",
		"key", key.KeyName,
		"x", x,
		"y", y)

	return x, y
}

// KeyPath is not used at the moment since the 'rotate' part does not work properly, and points
// to different coordinates than key boxes, and I couldn't figure out how to fix it. Instead I used
// JS approach to calculate coords dynamically, but maybe this can be fixed somehow.
func KeyPath(fromKey, toKey *Item) string {
	fromX, fromY := KeyCenter(fromKey)
	toX, toY := KeyCenter(toKey)

	// Calculate control points for a curved path
	midX := (fromX + toX) / 2
	midY := (fromY+toY)/2 - 40 // Curve upward

	return fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f", fromX, fromY, midX, midY, toX, toY)
}

func KeyPathStrokeWidth(conn *ComboConnection) string {
	// Draw a curved path with thickness based on press count
	// TODO: scaling by 100 is arbirary, we should adjust it based on total keystrokes in db
	// or something.
	strokeWidth := math.Min(10, 1+float64(conn.PressCount)/100)

	return fmt.Sprintf("%f", strokeWidth)
}

// Make svg transform expressions for keys. Usually they only have offset, sometimes rotations with a pivot point.
func ToTransform(l *model.Location) string {
	// log.Printf("ToTransform: %+v", l)
	translate := fmt.Sprintf("translate(%.2f, %.2f)", l.X*KeySize, l.Y*KeySize)

	if l.R != 0 {
		rx, ry := ToTransformOrigin(l)
		translate += fmt.Sprintf(" rotate(%.2f %.2f %.2f)", l.R, rx*KeySize, ry*KeySize)
	}

	return translate
}

// Make up a pivot point for rotated keys.
func ToTransformOrigin(l *model.Location) (float64, float64) {
	rx := l.Rx

	// I have no idea why this has to be done, but for json files in the contrib repo,
	// this seems to be necessary.
	if l.Rx != 0 {
		rx -= l.X
	}

	ry := l.Ry
	if l.Ry != 0 {
		ry -= l.Y
	}

	return rx, ry
}

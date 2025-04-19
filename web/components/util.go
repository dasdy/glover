package components

import "fmt"

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

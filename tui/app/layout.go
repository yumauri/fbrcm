package app

import "cmp"

// panelLayout holds panel layout state used by the app package.
type panelLayout struct {
	// topHeight stores top height for panelLayout.
	topHeight int
	// bottomHeight stores bottom height for panelLayout.
	bottomHeight int
	// leftWidth stores left width for panelLayout.
	leftWidth int
	// rightWidth stores right width for panelLayout.
	rightWidth int
	// bottomWidth stores bottom width for panelLayout.
	bottomWidth int
}

const (
	defaultLogsPanelHeight   = 7
	minLogsPanelHeight       = 3
	collapsedLogsPanelHeight = 1
	collapsedProjectsWidth   = 3
	minTopPanelsHeight       = 7
)

// newPanelLayout constructs new panel layout and returns the resulting value or error.
func newPanelLayout(width, height, preferredLeftWidth, preferredLogsHeight int, projectsMode projectsPanelMode) panelLayout {
	availableWidth := cmp.Or(width, 80)
	availableHeight := max(cmp.Or(height, 24)-helpLineHeight, 1)

	maxBottomHeight := max(availableHeight-minTopPanelsHeight, collapsedLogsPanelHeight)
	bottomHeight := min(preferredLogsHeight, maxBottomHeight)
	if bottomHeight == collapsedLogsPanelHeight {
		// Keep explicit collapsed state even when there is room for content.
	} else if availableHeight >= minTopPanelsHeight+minLogsPanelHeight {
		bottomHeight = max(bottomHeight, minLogsPanelHeight)
	} else {
		bottomHeight = max(bottomHeight, collapsedLogsPanelHeight)
	}
	topHeight := max(availableHeight-bottomHeight, 1)
	if topHeight+bottomHeight > availableHeight {
		topHeight = max(availableHeight-bottomHeight, 1)
	}

	leftWidth := availableWidth / 2
	if projectsMode == projectsPanelModeCollapsed {
		leftWidth = min(collapsedProjectsWidth, max(availableWidth-1, 1))
	} else {
		if preferredLeftWidth > 0 {
			leftWidth = min(preferredLeftWidth, max(availableWidth-1, 1))
		} else if availableWidth >= 48 {
			leftWidth = max(leftWidth, 24)
		}
		if leftWidth >= availableWidth {
			leftWidth = max(availableWidth-1, 1)
		}
	}
	rightWidth := max(availableWidth-leftWidth, 1)

	return panelLayout{
		topHeight:    topHeight,
		bottomHeight: bottomHeight,
		leftWidth:    leftWidth,
		rightWidth:   rightWidth,
		bottomWidth:  availableWidth,
	}
}

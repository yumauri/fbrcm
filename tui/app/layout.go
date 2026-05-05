package app

import "cmp"

type panelLayout struct {
	topHeight    int
	bottomHeight int
	leftWidth    int
	rightWidth   int
	bottomWidth  int
}

const (
	defaultLogsPanelHeight = 7
	minLogsPanelHeight     = 3
)

func newPanelLayout(width, height, preferredLeftWidth, preferredLogsHeight int) panelLayout {
	availableWidth := cmp.Or(width, 80)
	availableHeight := cmp.Or(height, 24)

	maxBottomHeight := max(availableHeight-1, 1)
	bottomHeight := min(preferredLogsHeight, maxBottomHeight)
	if availableHeight >= minLogsPanelHeight+1 {
		bottomHeight = max(bottomHeight, minLogsPanelHeight)
	}
	topHeight := max(availableHeight-bottomHeight, 1)
	if topHeight+bottomHeight > availableHeight {
		topHeight = max(availableHeight-bottomHeight, 1)
	}

	leftWidth := availableWidth / 2
	if preferredLeftWidth > 0 {
		leftWidth = min(preferredLeftWidth, max(availableWidth-1, 1))
	} else if availableWidth >= 48 {
		leftWidth = max(leftWidth, 24)
	}
	if leftWidth >= availableWidth {
		leftWidth = max(availableWidth-1, 1)
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

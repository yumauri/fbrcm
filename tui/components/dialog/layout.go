package dialog

import (
	"charm.land/lipgloss/v2"
	"strings"
)

func (m Model) maxScroll() int { return max(len(m.bodyLines())-m.bodyHeight(), 0) }

func (m Model) scrollbar() scrollbarState {
	contentHeight, totalLines := m.bodyHeight(), len(m.bodyLines())
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}
	thumbHeight := min(max(2, (contentHeight*contentHeight)/totalLines), contentHeight)
	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.scroll * maxThumbStart) / maxOffset
	return scrollbarState{visible: true, thumbStart: thumbStart, thumbEnd: min(thumbStart+thumbHeight-1, contentHeight-1)}
}

func (m Model) bodyHeight() int { return max(min(max(m.height-10, 3), len(m.bodyLines())), 1) }

func (m Model) contentWidth() int {
	width := len([]rune(m.title))
	for _, line := range m.body {
		width = max(width, printableWidth(line))
	}
	width = max(width, printableWidth(m.renderButtons())) + 2
	return min(max(width, 38), min(max(m.width-12, 38), 88))
}

func (m Model) boxGeometry() (x, y, width, height int) {
	contentWidth, bodyHeight := m.contentWidth(), m.bodyHeight()
	width, height = contentWidth+7, bodyHeight+m.buttonHeight()+4
	if m.positioned {
		x = clamp(m.manualX, m.x, max(m.x+m.width-width, m.x))
		y = clamp(m.manualY, m.y, max(m.y+m.height-height, m.y))
		return
	}
	x = max(m.x+(m.width-width)/2, m.x)
	y = max(m.y+(m.height-height)/2, m.y)
	return
}

func (m *Model) setManualPosition(x, y int) {
	_, _, boxWidth, boxHeight := m.boxGeometry()
	m.manualX = clamp(x, m.x, max(m.x+m.width-boxWidth, m.x))
	m.manualY = clamp(y, m.y, max(m.y+m.height-boxHeight, m.y))
	m.positioned = true
}

func (m Model) titleHit(x, y int) bool {
	boxX, boxY, boxWidth, _ := m.boxGeometry()
	return y == boxY && x >= boxX && x < boxX+boxWidth
}

func (m Model) buttonHeight() int { return max(lipgloss.Height(m.renderButtons()), 1) }

func clamp(value, low, high int) int { return max(low, min(value, high)) }

func (m Model) buttonIndexAt(x, y int) (int, bool) {
	if !m.open || len(m.buttons) == 0 {
		return -1, false
	}
	boxX, boxY, _, _ := m.boxGeometry()
	contentWidth, bodyHeight := m.contentWidth(), m.bodyHeight()
	buttonLines := strings.Split(m.renderButtons(), "\n")
	if len(buttonLines) == 0 {
		return -1, false
	}
	buttonX := boxX + 4 + max(contentWidth-printableWidth(buttonLines[0]), 0)
	buttonY := boxY + bodyHeight + 3
	if y < buttonY || y >= buttonY+len(buttonLines) {
		return -1, false
	}
	return m.buttonBar().IndexAt(x-buttonX, y-buttonY)
}

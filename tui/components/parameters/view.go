package parameters

import (
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
)

func (m Model) View(active bool) string {
	projectLine, groupLine, bodyStart := m.stickyHeaderLines(m.offset)
	bodyLines := m.visibleBodyLines(bodyStart)
	lines := make([]string, 0, len(bodyLines)+2)
	if projectLine != "" {
		lines = append(lines, projectLine)
	}
	if groupLine != "" {
		lines = append(lines, groupLine)
	}
	lines = append(lines, bodyLines...)
	return renderPanel(lines, m.width, m.height, active, m.scrollbar(), m.filter.View(max(m.width-1, 1), active, m.filteredParameterCount()))
}

func (m Model) renderBody() []string {
	if len(m.visible) == 0 {
		return []string{
			"Select project in Projects panel.",
			"",
			"Selected project will appear here immediately.",
		}
	}

	height := m.contentHeight()
	if height <= 0 {
		return nil
	}

	lines := make([]string, 0, len(m.visible)+4)
	for i := 0; i < len(m.visible); i++ {
		lines = append(lines, m.renderNodeBlock(i, false)...)
	}
	return lines
}

func (m Model) visibleBodyLines(startLine int) []string {
	height := m.bodyVisibleLinesForOffset(m.offset)
	if height <= 0 {
		return nil
	}

	if len(m.visible) == 0 {
		width := m.viewportWidth()
		lines := m.renderBody()
		for i := range lines {
			lines[i] = viewutil.PadRight(lipgloss.NewStyle().MaxWidth(width).Render(lines[i]), width)
		}
		for len(lines) < height {
			lines = append(lines, "")
		}
		return lines[:height]
	}

	width := m.viewportWidth()
	endLine := startLine + height
	lines := make([]string, 0, height)

	for i := 0; i < len(m.visible); i++ {
		rowStart := m.lineIndexByNode[i]
		rowHeight := m.nodeBlockLineCount(i)
		rowEnd := rowStart + rowHeight
		if rowEnd <= startLine {
			continue
		}
		if rowStart >= endLine {
			break
		}

		blockLines := m.renderNodeBlock(i, i == m.cursor)
		sliceStart := max(0, startLine-rowStart)
		sliceEnd := min(len(blockLines), endLine-rowStart)
		for _, line := range blockLines[sliceStart:sliceEnd] {
			lines = append(lines, lipgloss.NewStyle().MaxWidth(width).Render(line))
		}
		if len(lines) >= height {
			break
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

func (m Model) scrollbar() scrollbarState {
	contentHeight := m.viewportHeight()
	totalLines := m.totalLines
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}

	thumbHeight := max(2, (contentHeight*contentHeight)/totalLines)
	thumbHeight = min(thumbHeight, contentHeight)

	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.offset * maxThumbStart) / maxOffset

	return scrollbarState{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   min(thumbStart+thumbHeight-1, contentHeight-1),
	}
}

func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
}

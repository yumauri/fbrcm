// Package contextpanel renders the two-line facts and choices used by Context.
package contextpanel

import (
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/mouseutil"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Row struct {
	Key, Label, Value string
}

type Layout struct {
	X, Y, Width, Height int
	Title, Empty        string
	TitleKey            string
	Sidebar             bool
	Count               int
	Footer              string
}

type Model struct {
	rows      []Row
	cursor    int
	offset    int
	lastClick mouseutil.ClickTracker
}

func (m Model) SetRows(rows []Row, selectedKey string) Model {
	if selectedKey == "" {
		selectedKey = m.Current().Key
	}
	m.rows = rows
	m.lastClick.Reset()
	m.cursor = min(m.cursor, max(len(rows)-1, 0))
	if i := slices.IndexFunc(rows, func(row Row) bool { return row.Key == selectedKey }); i >= 0 {
		m.cursor = i
	}
	return m
}

func (m Model) Len() int { return len(m.rows) }

func (m Model) Current() Row {
	if len(m.rows) == 0 {
		return Row{}
	}
	return m.rows[m.cursor]
}

func (m *Model) Move(delta int, layout Layout) {
	m.moveTo(m.cursor+delta, layout)
}

func (m *Model) Jump(end bool, layout Layout) {
	index := 0
	if end {
		index = len(m.rows) - 1
	}
	m.moveTo(index, layout)
}

func (m *Model) moveTo(index int, layout Layout) {
	m.lastClick.Reset()
	m.cursor = min(max(index, 0), max(len(m.rows)-1, 0))
	m.offset = m.firstRow(layout)
}

func (m Model) firstRow(layout Layout) int {
	count := layout.rowCount()
	return min(max(m.offset, m.cursor-count+1, 0), m.cursor, max(len(m.rows)-count, 0))
}

func (l Layout) rowCount() int {
	return max(l.bodyHeight()/2, 0)
}

func (l Layout) bodyHeight() int {
	height := l.Height - 2
	if l.Footer != "" {
		height--
	}
	return max(height, 0)
}

func (l Layout) leftBorderWidth() int {
	if l.Sidebar {
		return 0
	}
	return 1
}

func (l Layout) Contains(x, y int) bool {
	return x >= l.X && x < l.X+l.Width && y >= l.Y && y < l.Y+l.Height
}

// Click selects the item covering either of its two lines and reports activation on double-click.
func (m *Model) Click(l Layout, x, y int, at time.Time) bool {
	if !l.Contains(x, y) {
		return false
	}
	x, y = x-l.X, y-l.Y
	if x < l.leftBorderWidth() || x == l.Width-1 || y == 0 || y > 2*l.rowCount() {
		m.lastClick.Reset()
		return false
	}
	index := m.firstRow(l) + (y-1)/2
	if index >= len(m.rows) {
		return false
	}
	m.cursor = index
	return m.lastClick.Register(1, index, at)
}

func (m Model) View(l Layout, active bool) string {
	if l.Height < 1 || l.Width < 3 {
		return ""
	}
	border := styles.BorderStyle(active)
	top, left, bottomLeft := "", "", ""
	count := ""
	if l.Count > 0 {
		count = strconv.Itoa(l.Count)
	}
	if l.Sidebar {
		top = viewutil.SidebarHeader(l.Width, l.TitleKey, l.Title, active, border, count, styles.SecondaryTitleCount)
	} else {
		top = viewutil.PopupHeader(l.Width, l.TitleKey, cleanLine(l.Title), active, border, count, styles.SecondaryTitleCount)
		left, bottomLeft = border.Render("│"), border.Render("╰")
	}
	lines := []string{top}
	if l.Height == 1 {
		return lines[0]
	}

	inner := l.Width - l.leftBorderWidth() - 1
	frame := func(line string) string {
		return left + viewutil.PadRight(line, inner) + border.Render("│")
	}
	branch := ansi.Truncate(" ╰ ", inner, "")
	valueWidth := inner - lipgloss.Width(branch)
	first := m.firstRow(l)
	for i := first; i < min(first+l.rowCount(), len(m.rows)); i++ {
		row := m.rows[i]
		nameStyle := styles.PanelTitle.Bold(true)
		valueStyle, branchStyle := styles.PanelBody, styles.TreeBranch
		if active && i == m.cursor {
			valueStyle = styles.TreeItemSelectionStyle()
			nameStyle = valueStyle.Bold(true)
			if styles.NoColorEnabled() {
				branchStyle = valueStyle
			} else {
				branchStyle = branchStyle.Background(valueStyle.GetBackground())
			}
		}
		label := ansi.Truncate(" "+cleanLine(row.Label), inner, "…")
		value := ansi.Truncate(cleanLine(row.Value), valueWidth, "…")
		lines = append(lines,
			frame(nameStyle.Render(viewutil.PadRight(label, inner))),
			frame(branchStyle.Render(branch)+valueStyle.Render(viewutil.PadRight(value, valueWidth))),
		)
	}
	if len(m.rows) == 0 && l.bodyHeight() > 0 {
		lines = append(lines, frame(styles.PanelMuted.Render(ansi.Truncate(" "+l.Empty, inner, "…"))))
	}
	for len(lines) <= l.bodyHeight() {
		lines = append(lines, frame(""))
	}
	if l.Footer != "" {
		lines = append(lines, frame(l.Footer))
	}
	lines = append(lines, bottomLeft+border.Render(strings.Repeat("─", inner)+"╯"))
	return strings.Join(lines, "\n")
}

func cleanLine(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}

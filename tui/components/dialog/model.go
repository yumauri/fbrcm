package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type ButtonVariant int

const (
	ButtonVariantNeutral ButtonVariant = iota
	ButtonVariantDanger
	ButtonVariantAccent
)

type Button struct {
	Label   string
	Variant ButtonVariant
	OnPress tea.Cmd
}

type Config struct {
	Title   string
	Body    []string
	Buttons []Button
}

type Model struct {
	x          int
	y          int
	width      int
	height     int
	manualX    int
	manualY    int
	title      string
	body       []string
	buttons    []Button
	selected   int
	scroll     int
	open       bool
	positioned bool
	dragging   bool
	dragOffX   int
	dragOffY   int
}

func New() Model {
	return Model{}
}

// SetBounds sets bounds for Model and returns the resulting state or error.
func (m Model) SetBounds(x, y, width, height int) Model {
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	return m
}

func (m Model) CenterWithin(x, y, width, height int) Model {
	_, _, boxWidth, boxHeight := m.boxGeometry()
	m.manualX = max(x+(width-boxWidth)/2, x)
	m.manualY = max(y+(height-boxHeight)/2, y)
	m.positioned = true
	return m
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(cfg Config) Model {
	m.open = true
	m.title = cfg.Title
	m.body = append([]string(nil), cfg.Body...)
	m.buttons = append([]Button(nil), cfg.Buttons...)
	m.selected = 0
	m.scroll = 0
	m.positioned = false
	m.dragging = false
	m.dragOffX = 0
	m.dragOffY = 0
	return m
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.title = ""
	m.body = nil
	m.buttons = nil
	m.selected = 0
	m.scroll = 0
	m.positioned = false
	m.dragging = false
	m.dragOffX = 0
	m.dragOffY = 0
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Contains(x, y int) bool {
	if !m.open {
		return false
	}
	boxX, boxY, boxWidth, boxHeight := m.boxGeometry()
	return x >= boxX && x < boxX+boxWidth && y >= boxY && y < boxY+boxHeight
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionPrev, k):
			m.move(-1)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionNext, k):
			m.move(1)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionUp, k):
			m.scrollBy(-1)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionDown, k):
			m.scrollBy(1)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionPageUp, k):
			m.scrollBy(-5)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionPageDown, k):
			m.scrollBy(5)
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionHome, k):
			m.scroll = 0
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionEnd, k):
			m.scroll = m.maxScroll()
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionSubmit, k):
			if m.selected >= 0 && m.selected < len(m.buttons) {
				cmd := m.buttons[m.selected].OnPress
				m = m.Close()
				return m, cmd
			}
		}
	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if mouse.Button == tea.MouseLeft && m.titleHit(mouse.X, mouse.Y) {
			boxX, boxY, _, _ := m.boxGeometry()
			m.dragging = true
			m.dragOffX = mouse.X - boxX
			m.dragOffY = mouse.Y - boxY
			return m, nil
		}
		index, ok := m.buttonIndexAt(msg.Mouse().X, msg.Mouse().Y)
		if !ok {
			return m, nil
		}
		m.selected = index
		if index >= 0 && index < len(m.buttons) {
			cmd := m.buttons[index].OnPress
			m = m.Close()
			return m, cmd
		}
	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		if m.dragging {
			m.setManualPosition(mouse.X-m.dragOffX, mouse.Y-m.dragOffY)
			return m, nil
		}
		index, ok := m.buttonIndexAt(mouse.X, mouse.Y)
		if ok {
			m.selected = index
		}
	case tea.MouseReleaseMsg:
		m.dragging = false
	}

	return m, nil
}

// move moves move for Model and returns the resulting state or error.
func (m *Model) move(delta int) {
	if len(m.buttons) == 0 {
		return
	}
	m.selected = (m.selected + delta + len(m.buttons)) % len(m.buttons)
}

func (m *Model) scrollBy(delta int) {
	m.scroll = max(0, min(m.scroll+delta, m.maxScroll()))
}

func (m Model) maxScroll() int {
	return max(len(m.body)-m.bodyHeight(), 0)
}

func (m Model) scrollbar() scrollbarState {
	contentHeight := m.bodyHeight()
	totalLines := len(m.body)
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}

	thumbHeight := max(2, (contentHeight*contentHeight)/totalLines)
	thumbHeight = min(thumbHeight, contentHeight)

	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.scroll * maxThumbStart) / maxOffset

	return scrollbarState{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   min(thumbStart+thumbHeight-1, contentHeight-1),
	}
}

func (m Model) bodyHeight() int {
	return max(min(max(m.height-10, 3), len(m.body)), 1)
}

func (m Model) contentWidth() int {
	width := len([]rune(m.title))
	for _, line := range m.body {
		width = max(width, printableWidth(line))
	}
	buttonRow := m.renderButtons()
	width = max(width, printableWidth(buttonRow))
	width += 2
	return min(max(width, 38), min(max(m.width-12, 38), 88))
}

func (m Model) boxGeometry() (x, y, width, height int) {
	contentWidth := m.contentWidth()
	bodyHeight := m.bodyHeight()
	width = contentWidth + 6
	height = bodyHeight + m.buttonHeight() + 4
	if m.positioned {
		x = clamp(m.manualX, m.x, max(m.x+m.width-width, m.x))
		y = clamp(m.manualY, m.y, max(m.y+m.height-height, m.y))
		return
	}
	x = max(m.x+(m.width-width)/2, m.x)
	y = max(m.y+(m.height-height)/2, m.y)
	return
}

// setManualPosition sets set manual position for Model and returns the resulting state or error.
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

func (m Model) buttonHeight() int {
	return max(lipgloss.Height(m.renderButtons()), 1)
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func (m Model) buttonIndexAt(x, y int) (int, bool) {
	if !m.open || len(m.buttons) == 0 {
		return -1, false
	}

	boxX, boxY, _, _ := m.boxGeometry()
	contentWidth := m.contentWidth()
	bodyHeight := m.bodyHeight()
	buttonBlock := m.renderButtons()
	buttonLines := strings.Split(buttonBlock, "\n")
	if len(buttonLines) == 0 {
		return -1, false
	}

	contentX := boxX + 3
	buttonX := contentX + max(contentWidth-printableWidth(buttonLines[0]), 0)
	buttonY := boxY + bodyHeight + 3
	if y < buttonY || y >= buttonY+len(buttonLines) {
		return -1, false
	}

	offsetX := buttonX
	for i, button := range m.renderedButtons() {
		w := printableWidth(button)
		h := lipgloss.Height(button)
		if x >= offsetX && x < offsetX+w && y >= buttonY && y < buttonY+h {
			return i, true
		}
		offsetX += w
		if i < len(m.buttons)-1 {
			offsetX++
		}
	}

	return -1, false
}

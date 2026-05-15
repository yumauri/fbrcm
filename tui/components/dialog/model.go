package dialog

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"strings"
)

type ButtonVariant int

const (
	ButtonVariantNeutral ButtonVariant = iota
	ButtonVariantDanger
	ButtonVariantAccent
)

// Button holds button state used by the dialog package.
type Button struct {
	// Label stores label for Button.
	Label string
	// Variant stores variant for Button.
	Variant ButtonVariant
	// OnPress stores on press for Button.
	OnPress tea.Cmd
}

// Config holds config state used by the dialog package.
type Config struct {
	// Title stores title for Config.
	Title string
	// Body stores body for Config.
	Body []string
	// Buttons stores buttons for Config.
	Buttons []Button
}

// Model holds model state used by the dialog package.
type Model struct {
	// x stores x for Model.
	x int
	// y stores y for Model.
	y int
	// width stores width for Model.
	width int
	// height stores height for Model.
	height int
	// manualX stores manual x for Model.
	manualX int
	// manualY stores manual y for Model.
	manualY int
	// title stores title for Model.
	title string
	// body stores body for Model.
	body []string
	// buttons stores buttons for Model.
	buttons []Button
	// selected stores selected for Model.
	selected int
	// scroll stores scroll for Model.
	scroll int
	// open stores open for Model.
	open bool
	// positioned stores positioned for Model.
	positioned bool
	// dragging stores dragging for Model.
	dragging bool
	// dragOffX stores drag off x for Model.
	dragOffX int
	// dragOffY stores drag off y for Model.
	dragOffY int
}

// New constructs new and returns the resulting value or error.
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

// CenterWithin handles center within for Model and returns the resulting state or error.
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

// Contains handles contains for Model and returns the resulting state or error.
func (m Model) Contains(x, y int) bool {
	if !m.open {
		return false
	}
	boxX, boxY, boxWidth, boxHeight := m.boxGeometry()
	return x >= boxX && x < boxX+boxWidth && y >= boxY && y < boxY+boxHeight
}

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "shift+tab":
			m.move(-1)
		case "right", "l", "tab":
			m.move(1)
		case "up", "k":
			m.scrollBy(-1)
		case "down", "j":
			m.scrollBy(1)
		case "pgup":
			m.scrollBy(-5)
		case "pgdown":
			m.scrollBy(5)
		case "home":
			m.scroll = 0
		case "end":
			m.scroll = m.maxScroll()
		case "enter":
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

// scrollBy handles scroll by for Model and returns the resulting state or error.
func (m *Model) scrollBy(delta int) {
	m.scroll = max(0, min(m.scroll+delta, m.maxScroll()))
}

// maxScroll handles max scroll for Model and returns the resulting state or error.
func (m Model) maxScroll() int {
	return max(len(m.body)-m.bodyHeight(), 0)
}

// bodyHeight handles body height for Model and returns the resulting state or error.
func (m Model) bodyHeight() int {
	return max(min(max(m.height-10, 3), len(m.body)), 1)
}

// contentWidth handles content width for Model and returns the resulting state or error.
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

// boxGeometry handles box geometry for Model and returns the resulting state or error.
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

// titleHit handles title hit for Model and returns the resulting state or error.
func (m Model) titleHit(x, y int) bool {
	boxX, boxY, boxWidth, _ := m.boxGeometry()
	return y == boxY && x >= boxX && x < boxX+boxWidth
}

// buttonHeight handles button height for Model and returns the resulting state or error.
func (m Model) buttonHeight() int {
	return max(lipgloss.Height(m.renderButtons()), 1)
}

// clamp handles clamp and returns the resulting value or error.
func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// buttonIndexAt handles button index at for Model and returns the resulting state or error.
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

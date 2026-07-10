package dialog

import (
	tea "charm.land/bubbletea/v2"

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

func (m *Model) move(delta int) {
	if len(m.buttons) == 0 {
		return
	}
	m.selected = (m.selected + delta + len(m.buttons)) % len(m.buttons)
}

func (m *Model) scrollBy(delta int) {
	m.scroll = max(0, min(m.scroll+delta, m.maxScroll()))
}

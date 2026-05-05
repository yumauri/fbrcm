package logs

import (
	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.LogLineMsg:
		if msg.Line != "" {
			m.lines = append(m.lines, msg.Line)
			m.refreshViewport()
		}
		return m, waitForLogCmd(m.sub)

	case tea.KeyMsg:
		if !m.active {
			break
		}
		switch msg.String() {
		case "[":
			m.moveLevel(-1)
		case "]":
			m.moveLevel(1)
		case "enter":
			m.lines = append(m.lines, "")
			m.refreshViewport()
		case "up", "k":
			m.viewport.ScrollUp(1)
			m.follow = false
		case "down", "j":
			m.viewport.ScrollDown(1)
			m.follow = m.viewport.AtBottom()
		case "pgup", "h":
			m.viewport.PageUp()
			m.follow = false
		case "pgdown", "l":
			m.viewport.PageDown()
			m.follow = m.viewport.AtBottom()
		case "home":
			m.viewport.GotoTop()
			m.follow = false
		case "end":
			m.viewport.GotoBottom()
			m.follow = true
		}

	case tea.MouseWheelMsg:
		if !m.isMouseInside(msg.Mouse()) {
			break
		}
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.viewport.ScrollUp(1)
			m.follow = false
		case tea.MouseWheelDown:
			m.viewport.ScrollDown(1)
			m.follow = m.viewport.AtBottom()
		}
	}

	return m, nil
}

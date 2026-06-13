package logs

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.LogLineMsg:
		if msg.Line != "" {
			m.lines = append(m.lines, msg.Line)
			m.refreshViewport()
			if isErrorLogLine(msg.Line) && m.statusFlashLeft == 0 && !m.statusFlashOn {
				m.statusFlashOn = true
				m.statusFlashLeft = statusFlashToggles - 1
				return m, tea.Batch(waitForLogCmd(m.sub), statusFlashTickCmd())
			}
		}
		return m, waitForLogCmd(m.sub)

	case statusFlashTickMsg:
		if m.statusFlashLeft <= 0 {
			m.statusFlashOn = false
			m.statusFlashLeft = 0
			return m, nil
		}
		m.statusFlashOn = !m.statusFlashOn
		m.statusFlashLeft--
		if m.statusFlashLeft > 0 {
			return m, statusFlashTickCmd()
		}
		m.statusFlashOn = false
		return m, nil

	case tea.KeyMsg:
		if !m.active {
			break
		}
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionLevelDown, k):
			m.moveLevel(-1)
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionLevelUp, k):
			m.moveLevel(1)
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionBlankLine, k):
			m.lines = append(m.lines, "")
			m.refreshViewport()
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionUp, k):
			m.viewport.ScrollUp(1)
			m.follow = false
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionDown, k):
			m.viewport.ScrollDown(1)
			m.follow = m.viewport.AtBottom()
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionPageUp, k):
			m.viewport.PageUp()
			m.follow = false
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionPageDown, k):
			m.viewport.PageDown()
			m.follow = m.viewport.AtBottom()
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionHome, k):
			m.viewport.GotoTop()
			m.follow = false
		case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionEnd, k):
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

// isErrorLogLine reports is error log line and returns the resulting value or error.
func isErrorLogLine(line string) bool {
	plain := ansiOSCRe.ReplaceAllString(line, "")
	plain = ansiCSIRe.ReplaceAllString(plain, "")
	upper := strings.ToUpper(plain)
	return strings.Contains(upper, " ERROR ") || strings.Contains(upper, " ERRO ")
}

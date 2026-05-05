package logs

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	charmlog "github.com/charmbracelet/log"

	"fbrcm/core"
	corelog "fbrcm/core/log"
	"fbrcm/tui/messages"
)

type Model struct {
	svc *core.Core

	viewport viewport.Model
	lines    []string
	active   bool
	follow   bool
	level    charmlog.Level
	sub      <-chan string
	x        int
	y        int
	width    int
	height   int
}

func New(svc *core.Core) Model {
	vp := viewport.New(
		viewport.WithWidth(1),
		viewport.WithHeight(1),
	)
	vp.SoftWrap = true
	vp.MouseWheelEnabled = false

	ch, _ := corelog.Subscribe()

	m := Model{
		svc:      svc,
		viewport: vp,
		follow:   true,
		level:    corelog.CurrentLevel(),
		lines:    corelog.Snapshot(),
		sub:      ch,
	}
	m.refreshViewport()
	return m
}

func (m Model) Init() tea.Cmd {
	return waitForLogCmd(m.sub)
}

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.refreshViewport()
	return m
}

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.refreshViewport()
	return m
}

func (m Model) SetActive(active bool) Model {
	m.active = active
	return m
}

func (m *Model) moveLevel(delta int) {
	levels := corelog.AvailableLevels()
	current := 0
	for i, level := range levels {
		if level == m.level {
			current = i
			break
		}
	}

	next := current + delta
	if next < 0 || next >= len(levels) {
		return
	}

	m.level = levels[next]
	corelog.SetLevel(m.level)
}

func (m *Model) refreshViewport() {
	m.viewport.SetWidth(max(m.width, 1))
	m.viewport.SetHeight(max(m.height-2, 1))

	content := "No logs yet."
	if len(m.lines) > 0 {
		content = strings.Join(m.lines, "\n")
	}

	offset := m.viewport.YOffset()
	m.viewport.SetContent(content)
	if m.follow {
		m.viewport.GotoBottom()
		return
	}

	maxOffset := max(m.viewport.TotalLineCount()-m.viewport.Height(), 0)
	m.viewport.SetYOffset(min(offset, maxOffset))
}

func (m Model) isMouseInside(mouse tea.Mouse) bool {
	return mouse.X >= m.x && mouse.X < m.x+m.width && mouse.Y >= m.y && mouse.Y < m.y+m.height
}

func waitForLogCmd(ch <-chan string) tea.Cmd {
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return messages.LogLineMsg{Line: line}
	}
}

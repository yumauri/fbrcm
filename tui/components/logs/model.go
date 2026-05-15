package logs

import (
	"regexp"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	charmlog "github.com/charmbracelet/log"

	"fbrcm/core"
	corelog "fbrcm/core/log"
	"fbrcm/tui/messages"
)

// Model holds model state used by the logs package.
type Model struct {
	// svc stores svc for Model.
	svc *core.Core

	// viewport stores viewport for Model.
	viewport viewport.Model
	// lines stores lines for Model.
	lines []string
	// active stores active for Model.
	active bool
	// follow stores follow for Model.
	follow bool
	// level stores level for Model.
	level charmlog.Level
	// sub stores sub for Model.
	sub <-chan string
	// x stores x for Model.
	x int
	// y stores y for Model.
	y int
	// width stores width for Model.
	width int
	// height stores height for Model.
	height int
	// statusFlashOn stores status flash on for Model.
	statusFlashOn bool
	// statusFlashLeft stores status flash left for Model.
	statusFlashLeft int
}

const (
	statusFlashToggles = 6
	statusFlashStep    = 250 * time.Millisecond
)

var ansiCSIRe = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var ansiOSCRe = regexp.MustCompile(`\x1b]8;;.*?\x1b\\`)

// New constructs new and returns the resulting value or error.
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

// Init initializes init for Model and returns the resulting state or error.
func (m Model) Init() tea.Cmd {
	return waitForLogCmd(m.sub)
}

// SetSize sets size for Model and returns the resulting state or error.
func (m Model) SetSize(width, height int) Model {
	if m.width == width && m.height == height {
		return m
	}
	m.width = width
	m.height = height
	m.refreshViewport()
	return m
}

// SetBounds sets bounds for Model and returns the resulting state or error.
func (m Model) SetBounds(x, y, width, height int) Model {
	if m.x == x && m.y == y && m.width == width && m.height == height {
		return m
	}
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.refreshViewport()
	return m
}

// SetActive sets active for Model and returns the resulting state or error.
func (m Model) SetActive(active bool) Model {
	m.active = active
	return m
}

// moveLevel moves move level for Model and returns the resulting state or error.
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

// refreshViewport handles refresh viewport for Model and returns the resulting state or error.
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

// isMouseInside reports is mouse inside for Model and returns the resulting state or error.
func (m Model) isMouseInside(mouse tea.Mouse) bool {
	return mouse.X >= m.x && mouse.X < m.x+m.width && mouse.Y >= m.y && mouse.Y < m.y+m.height
}

// waitForLogCmd handles wait for log cmd and returns the resulting value or error.
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

// statusFlashTickMsg holds status flash tick msg state used by the logs package.
type statusFlashTickMsg struct{}

// statusFlashTickCmd handles status flash tick cmd and returns the resulting value or error.
func statusFlashTickCmd() tea.Cmd {
	return tea.Tick(statusFlashStep, func(time.Time) tea.Msg {
		return statusFlashTickMsg{}
	})
}

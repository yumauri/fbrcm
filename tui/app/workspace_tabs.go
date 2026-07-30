package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

var workspacePanelOrder = [...]panels.ID{
	panels.Parameters,
	panels.Conditions,
	panels.History,
	panels.ABTests,
	panels.Personalizations,
	panels.Rollouts,
}

func (m Model) workspaceTabAt(x, y int) (panels.ID, bool) {
	if m.promote.WorkspaceOpen() {
		return panels.None, false
	}
	if y != 0 || (m.detailsVisible && m.details.Contains(x, y)) {
		return panels.None, false
	}
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok || x < panelX {
		return panels.None, false
	}
	index, ok := header.TabAt(x - panelX)
	if !ok {
		return panels.None, false
	}
	return workspacePanel(index)
}

func (m Model) workspaceHeaderLayout() (workspaceheader.Layout, int, bool) {
	header, panelX, ok := m.workspaceBaseHeaderLayout()
	if !ok {
		return header, panelX, false
	}
	if preview, exists := m.workspaceMenuPreview(header); exists {
		header = header.WithPreview(preview)
	}
	return header, panelX, true
}

func (m Model) workspaceBaseHeaderLayout() (workspaceheader.Layout, int, bool) {
	if m.promote.WorkspaceOpen() {
		return workspaceheader.Layout{}, 0, false
	}
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	selected := workspaceTabIndex(m.selectedParametersTab())
	header := workspaceheader.LayoutForRightReserve(layout.rightWidth, selected, m.workspaceHeaderRightReserve())
	return header, layout.leftWidth, true
}

func (m Model) workspaceMenuPreview(header workspaceheader.Layout) (int, bool) {
	if !m.workspaceMenu {
		return 0, false
	}
	menuTabs := header.MenuTabs()
	if len(menuTabs) == 0 {
		return 0, false
	}
	cursor := min(max(m.workspaceCursor, 0), len(menuTabs)-1)
	return menuTabs[cursor], true
}

func (m Model) workspaceOverflowAt(x, y int) bool {
	if y != 0 || (m.detailsVisible && m.details.Contains(x, y)) {
		return false
	}
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok || !header.HasOverflow() {
		return false
	}
	buttonX, buttonWidth, ok := header.OverflowButtonBounds()
	return ok && x >= panelX+buttonX && x < panelX+buttonX+buttonWidth
}

func (m Model) workspaceMenuAvailable() bool {
	header, _, ok := m.workspaceHeaderLayout()
	return ok && header.HasOverflow()
}

func (m Model) openWorkspaceMenu() Model {
	header, _, ok := m.workspaceBaseHeaderLayout()
	if !ok || !header.HasOverflow() {
		return m
	}
	m.workspaceMenu = true
	m.workspaceCursor = header.MenuCursor()
	return m
}

func (m Model) closeWorkspaceMenu() Model {
	m.workspaceMenu = false
	m.workspaceCursor = 0
	return m
}

func (m Model) updateWorkspaceMenu(msg tea.Msg) (Model, tea.Cmd, bool) {
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok || !header.HasOverflow() {
		return m.closeWorkspaceMenu(), nil, true
	}
	menuTabs := header.MenuTabs()
	m.workspaceCursor = min(max(m.workspaceCursor, 0), len(menuTabs)-1)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		preview := menuTabs[m.workspaceCursor]
		m.updateWindowSize(msg)
		if !m.workspaceMenuAvailable() {
			m = m.closeWorkspaceMenu()
		} else if nextHeader, _, exists := m.workspaceHeaderLayout(); exists {
			nextTabs := nextHeader.MenuTabs()
			m.workspaceCursor = nextHeader.MenuCursor()
			for index, tab := range nextTabs {
				if tab == preview {
					m.workspaceCursor = index
					break
				}
			}
		}
		return m, nil, true
	case tea.KeyMsg:
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
			m = m.closeWorkspaceMenu()
			return m, m.requestQuit(), true
		case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionWorkspaceMenu, k),
			tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionCancel, k):
			return m.closeWorkspaceMenu(), nil, true
		case tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionUp, k):
			m.workspaceCursor = max(m.workspaceCursor-1, 0)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionDown, k):
			m.workspaceCursor = min(m.workspaceCursor+1, len(menuTabs)-1)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionHome, k):
			m.workspaceCursor = 0
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionEnd, k):
			m.workspaceCursor = len(menuTabs) - 1
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockWorkspaceMenu, tuiconfig.ActionSubmit, k):
			panel, exists := workspacePanel(menuTabs[m.workspaceCursor])
			m = m.closeWorkspaceMenu()
			if !exists {
				return m, nil, true
			}
			return m.activateWorkspacePanel(panel)
		default:
			if next, cmd, handled := m.updateGlobalFocusKey(k); handled {
				next = next.closeWorkspaceMenu()
				return next, cmd, true
			}
			return m, nil, true
		}
	case tea.MouseClickMsg:
		if msg.Mouse().Button != tea.MouseLeft {
			return m, nil, true
		}
		localX := msg.Mouse().X - panelX
		buttonX, buttonWidth, buttonExists := header.OverflowButtonBounds()
		if buttonExists && localX >= buttonX && localX < buttonX+buttonWidth && msg.Mouse().Y == 0 {
			return m.closeWorkspaceMenu(), nil, true
		}
		if tab, exists := header.PopupTabAt(m.workspaceCursor, localX, msg.Mouse().Y); exists {
			panel, panelExists := workspacePanel(tab)
			m = m.closeWorkspaceMenu()
			if !panelExists {
				return m, nil, true
			}
			return m.activateWorkspacePanel(panel)
		}
		if tab, exists := header.TabAt(localX); exists && msg.Mouse().Y == 0 {
			panel, panelExists := workspacePanel(tab)
			m = m.closeWorkspaceMenu()
			if !panelExists {
				return m, nil, true
			}
			return m.activateWorkspacePanel(panel)
		}
		return m.closeWorkspaceMenu(), nil, true
	case tea.MouseMsg:
		return m, nil, true
	}
	return m, nil, true
}

func (m Model) activateWorkspacePanel(panel panels.ID) (Model, tea.Cmd, bool) {
	if m.promote.WorkspaceOpen() {
		return m, nil, true
	}
	if workspaceTabIndex(panel) < 0 {
		return m, nil, false
	}
	m.setActive(panel)
	switch panel {
	case panels.History:
		var cmd tea.Cmd
		m.parameters, cmd = m.parameters.LoadHistory()
		return m, cmd, true
	case panels.ABTests:
		var cmd tea.Cmd
		m.abTests, cmd = m.abTests.Activate()
		return m, cmd, true
	case panels.Personalizations:
		var cmd tea.Cmd
		m.personalizations, cmd = m.personalizations.Activate()
		return m, cmd, true
	case panels.Rollouts:
		var cmd tea.Cmd
		m.rollouts, cmd = m.rollouts.Activate()
		return m, cmd, true
	}
	return m, nil, true
}

func workspaceTabIndex(panel panels.ID) int {
	for index, candidate := range workspacePanelOrder {
		if candidate == panel {
			return index
		}
	}
	return -1
}

func workspacePanel(index int) (panels.ID, bool) {
	if index < 0 || index >= len(workspacePanelOrder) {
		return panels.None, false
	}
	return workspacePanelOrder[index], true
}

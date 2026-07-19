package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	logscmp "github.com/yumauri/fbrcm/tui/components/logs"
	"github.com/yumauri/fbrcm/tui/components/minsize"
	"github.com/yumauri/fbrcm/tui/components/setup"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionForceQuit, keyMsg.String()) {
		return m, tea.Quit
	}
	if logscmp.IsBackgroundMessage(msg) {
		var cmd tea.Cmd
		m.logs, cmd = m.logs.Update(msg)
		return m, cmd
	}
	if ready, ok := msg.(setup.WorkspaceReadyMsg); ok {
		var cmds []tea.Cmd
		if ready.Reset {
			cmds = appendCmd(cmds, m.resetWorkspaceForProfile())
		}
		m.authCount = m.setup.AuthCount()
		m.setup = m.setup.Close()
		notice := ""
		if ready.CachedOnly {
			notice = "Cached only · press " + tuiconfig.Label(tuiconfig.BlockGlobal, tuiconfig.ActionAccounts) + " to add authentication"
		}
		m.projects = m.projects.SetNotice(notice)
		var cmd tea.Cmd
		m.projects, cmd = m.projects.Update(messages.ProjectsLoadedMsg{
			Projects: ready.Projects,
			Source:   ready.Source,
		})
		cmds = appendCmd(cmds, cmd)
		if m.width > 0 && m.height > 0 {
			m.applyLayout()
		}
		return m, tea.Batch(cmds...)
	}
	if _, ok := msg.(setup.CanceledMsg); ok {
		m.authCount = m.setup.AuthCount()
		m.setup = m.setup.Close()
		return m, nil
	}
	if _, ok := msg.(setup.QuitRequestedMsg); ok {
		m.setup = m.setup.Close()
		return m, m.requestQuit()
	}
	switch msg := msg.(type) {
	case setup.AuthPurgeRequestedMsg:
		m.openAuthPurgeDialog(msg)
		return m, nil
	case setup.ProfilePurgeRequestedMsg:
		m.openProfilePurgeDialog(msg)
		return m, nil
	case setup.ProfileRenameRequestedMsg:
		return m, m.openProfileRenameInput(msg)
	case setup.ErrorRequestedMsg:
		m.openSetupErrorDialog(msg)
		return m, nil
	case profileRenameCompletedMsg:
		return m.updateProfileRenameCompleted(msg)
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok && tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionHelp, keyMsg.String()) && (m.helpPalette.IsOpen() || m.helpShortcutAvailable(keyMsg.String())) {
		if m.helpPalette.IsOpen() {
			m.helpPalette = m.helpPalette.Close()
			return m, nil
		}
		var cmd tea.Cmd
		m.helpPalette, cmd = m.helpPalette.Open()
		return m, cmd
	}
	if m.helpPalette.IsOpen() {
		if size, ok := msg.(tea.WindowSizeMsg); ok {
			m.updateWindowSize(size)
			if size.Width >= minsize.MinWidth && size.Height >= minsize.MinHeight {
				actions := m.helpPalette.filtered(m.helpPaletteActions())
				m.helpPalette.ensureVisible(len(actions), helpPaletteListHeight(size.Height))
			}
			return m.updateChildPanels(msg)
		}
		next, cmd, _ := m.updateHelpPalette(msg)
		return next, cmd
	}
	if next, cmd, ok := m.updateOpenModal(msg); ok {
		return next, cmd
	}
	if m.setup.IsOpen() {
		if size, ok := msg.(tea.WindowSizeMsg); ok {
			m.updateWindowSize(size)
		}
		var cmd tea.Cmd
		m.setup, cmd = m.setup.Update(msg)
		return m, cmd
	}
	next, cmd, ok := m.updateAppMessage(msg)
	m = next
	if ok {
		return next, cmd
	}

	return m.updateChildPanels(msg)
}

func (m Model) helpShortcutAvailable(key string) bool {
	setupAvailable := !m.setup.IsOpen()
	if m.setup.IsOpen() {
		_, setupAvailable = m.setup.HelpBlock()
	}
	return setupAvailable &&
		m.width >= minsize.MinWidth &&
		m.height >= minsize.MinHeight &&
		(strings.HasPrefix(key, "ctrl+") || m.helpPlainKeyAvailable())
}

func (m Model) helpPlainKeyAvailable() bool {
	return !m.keyboardCaptured() &&
		!m.jsonInput.IsOpen() &&
		!m.numberInput.IsOpen() &&
		!m.stringInput.IsOpen() &&
		!m.moveParam.IsOpen() &&
		!m.authPicker.IsOpen() &&
		!m.renameInput.IsOpen() &&
		!m.projectIO.IsOpen() &&
		(m.active != panels.Details || !m.detailsVisible || !m.details.TextInputActive())
}

func (m *Model) closeDetailsIfOrphaned() {
	if !m.detailsVisible {
		return
	}
	data := m.details.Data()
	if data != nil {
		if m.parameters.HasProject(data.Project.ProjectID) {
			return
		}
		m.closeDetailsPanel()
		return
	}
	groupData := m.details.GroupData()
	if groupData != nil {
		if m.parameters.HasProject(groupData.Project.ProjectID) && m.parameters.HasGroup(groupData.Project.ProjectID, groupData.Group.Key) {
			return
		}
		m.closeDetailsPanel()
		return
	}
	conditionData := m.details.ConditionData()
	if conditionData != nil && !m.conditions.HasProject(conditionData.Project.ProjectID) {
		m.closeDetailsPanel()
	}
}

func (m *Model) applyLayout() {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)

	m.projects = m.projects.SetCollapsed(m.projectsMode == projectsPanelModeCollapsed)
	m.projects = m.projects.SetBounds(0, 0, layout.leftWidth, layout.topHeight)
	m.parameters = m.parameters.SetBounds(layout.leftWidth, 0, layout.rightWidth, layout.topHeight)
	m.conditions = m.conditions.SetBounds(layout.leftWidth, 0, layout.rightWidth, layout.topHeight)
	m.dialog = m.dialog.SetBounds(0, 0, m.width, m.height)
	m.authPicker = m.authPicker.SetBounds(0, 0, m.width, m.height)
	m.projectIO = m.projectIO.SetBounds(0, 0, m.width, m.height)
	detailsWidth := m.detailsWidthForLayout(layout)
	m.details = m.details.SetBounds(layout.bottomWidth-detailsWidth, 0, detailsWidth, layout.topHeight)
	m.logs = m.logs.SetBounds(0, layout.topHeight, layout.bottomWidth, layout.bottomHeight)
}

func (m Model) nextTabPanel() panels.ID {
	if m.active == panels.Logs {
		if m.detailsVisible {
			if m.prevTop == panels.Details || m.prevTop == panels.Parameters || m.prevTop == panels.Conditions || m.prevTop == panels.History {
				return m.prevTop
			}
			return m.selectedParametersTab()
		}
		return m.prevTop
	}

	if m.detailsVisible {
		if m.active == panels.Details {
			return m.selectedParametersTab()
		}
		if m.active == panels.Parameters || m.active == panels.Conditions {
			return panels.Details
		}
		return m.selectedParametersTab()
	}

	if m.active == panels.Parameters || m.active == panels.Conditions || m.active == panels.History {
		return panels.Projects
	}

	return m.selectedParametersTab()
}

func (m Model) selectedParametersTab() panels.ID {
	if m.parametersTab == panels.History || m.parametersTab == panels.Conditions {
		return m.parametersTab
	}
	return panels.Parameters
}

func (m *Model) setActive(panel panels.ID) {
	if panel != panels.Logs {
		m.prevTop = panel
	}
	m.active = panel
	if panel == panels.Parameters || panel == panels.Conditions || panel == panels.History {
		m.parametersTab = panel
		m.parameters = m.parameters.SetHistory(panel == panels.History)
	}
	if m.capture != panels.None && m.capture != panel {
		m.capture = panels.None
	}
	m.projects = m.projects.SetActive(panel == panels.Projects)
	m.parameters = m.parameters.SetActive(panel == panels.Parameters || panel == panels.History)
	m.conditions = m.conditions.SetActive(panel == panels.Conditions)
	m.details = m.details.SetActive(panel == panels.Details)
	m.details = m.details.SetBridgeActive(panel == panels.Parameters || panel == panels.Conditions)
	m.logs = m.logs.SetActive(panel == panels.Logs)
}

func (m Model) keyboardCaptured() bool {
	return m.capture != panels.None
}

func (m Model) panelAt(x, y int) (panels.ID, bool) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	if x < 0 || y < 0 || x >= layout.bottomWidth || y >= layout.topHeight+layout.bottomHeight {
		return 0, false
	}

	if m.detailsVisible && m.details.Contains(x, y) {
		return panels.Details, true
	}

	if y < layout.topHeight {
		if x < layout.leftWidth {
			return panels.Projects, true
		}
		return m.selectedParametersTab(), true
	}

	return panels.None, false
}

func (m *Model) resizeLogsHeight(delta int) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	m.logsHeight = nextLogsPanelHeight(layout.bottomHeight, delta)
	m.logsHeight = min(m.logsHeight, newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode).bottomHeight)
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

func nextLogsPanelHeight(current, delta int) int {
	if delta == 0 {
		return current
	}
	if delta > 0 {
		if current == collapsedLogsPanelHeight {
			return minLogsPanelHeight
		}
		return current + 1
	}
	if current == minLogsPanelHeight {
		return collapsedLogsPanelHeight
	}
	if current == collapsedLogsPanelHeight {
		return collapsedLogsPanelHeight
	}
	return current - 1
}

func initialLogsPanelHeight(terminalHeight int) int {
	if terminalHeight <= 35 {
		return collapsedLogsPanelHeight
	}
	if terminalHeight >= 40 {
		return defaultLogsPanelHeight
	}
	return terminalHeight - 33
}

func (m Model) detailsWidthForLayout(layout panelLayout) int {
	minWidth := max(layout.rightWidth/2, 1)
	maxWidth := max(layout.rightWidth-11, 1)

	// Details content width = panel width - 5:
	// bridge spacer, left border, left padding, right padding, scrollbar lane.
	nameFitWidth := max(m.parameters.LongestParameterNameWidth(), m.conditions.LongestConditionNameWidth()) + 5

	desired := max(minWidth, nameFitWidth)
	return min(desired, maxWidth)
}

package app

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) updateAppMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case messages.HistoryRollbackRequestedMsg:
		return m.beginHistoryRollback(msg)

	case messages.HistoryRollbackPreviewLoadedMsg:
		return m.updateHistoryRollbackPreview(msg)

	case messages.HistoryRollbackConfirmedMsg:
		return m.confirmHistoryRollback()

	case messages.HistoryRollbackCanceledMsg:
		m.cancelHistoryRollback()
		return m, nil, true

	case messages.HistoryRollbackCompletedMsg:
		return m.updateHistoryRollbackCompleted(msg)

	case messages.KeyboardCaptureMsg:
		if msg.Enabled {
			m.capture = m.active
		} else {
			m.capture = panels.None
		}

	case messages.SetActivePanelMsg:
		panel := msg.Panel
		if panel == panels.Parameters && m.active == panels.Projects && !msg.ResetParametersTab {
			panel = m.selectedParametersTab()
		}
		m.setActive(panel)
		if panel == panels.History {
			var cmd tea.Cmd
			m.parameters, cmd = m.parameters.LoadHistory()
			return m, cmd, true
		}

	case messages.DialogCanceledMsg:
		if m.historyRollback != nil {
			m.cancelHistoryRollback()
			return m, nil, true
		}
		m.dialogQueue = nil
		m.pendingDetails = nil

	case messages.DetailsEditCanceledMsg:
		return m.updateDetailsEditCanceled(msg)

	case messages.DetailsInvalidFixMsg:
		m.updateDetailsInvalidFix()

	case messages.DetailsInvalidDiscardMsg:
		return m.updateDetailsInvalidDiscard(msg)

	case messages.DetailsValueEditRequestedMsg:
		if m.active == panels.Details && m.detailsVisible && m.details.ValueSelected() {
			return m, m.openDetailsValueEditor(), true
		}

	case messages.ParameterSelectionChangedMsg:
		return m, m.handleParameterSelection(msg), true

	case messages.ConditionSelectionChangedMsg:
		m.applyConditionSelection(msg)
		return m, nil, true

	case messages.ParametersLoadedMsg:
		m.updateParametersLoadedMessage(msg)

	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)

	case tea.KeyMsg:
		return m.updateKeyMessage(msg)

	case tea.PasteMsg, tea.ClipboardMsg:
		return m.updatePasteMessage(msg)

	case tea.MouseClickMsg:
		return m.updatePanelMouseMessage(msg)

	case tea.MouseWheelMsg:
		return m.updatePanelMouseMessage(msg)

	case tea.MouseMotionMsg:
	case tea.MouseReleaseMsg:
	}
	return m, nil, false
}

func (m Model) updateChildPanels(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.projects, cmd = m.projects.Update(msg)
	cmds = appendCmd(cmds, cmd)
	if _, ok := msg.(tea.WindowSizeMsg); !ok && m.width > 0 && m.height > 0 {
		m.applyLayout()
	}

	m.parameters, cmd = m.parameters.Update(msg)
	cmds = appendCmd(cmds, cmd)
	m.conditions, cmd = m.conditions.Update(msg)
	cmds = appendCmd(cmds, cmd)
	m.closeDetailsIfOrphaned()

	m.details, cmd = m.details.Update(msg)
	cmds = appendCmd(cmds, cmd)

	m.logs, cmd = m.logs.Update(msg)
	cmds = appendCmd(cmds, cmd)
	m.updateParametersLoadedPanelState(msg)

	return m, tea.Batch(cmds...)
}

func appendCmd(cmds []tea.Cmd, cmd tea.Cmd) []tea.Cmd {
	if cmd == nil {
		return cmds
	}
	return append(cmds, cmd)
}

func (m *Model) updateParametersLoadedPanelState(msg tea.Msg) {
	loadedMsg, ok := msg.(messages.ParametersLoadedMsg)
	if !ok || loadedMsg.Err != nil {
		return
	}
	m.updateDetailsAfterParametersLoaded(loadedMsg)
	m.updateDuplicateAfterParametersLoaded(loadedMsg)
}

func (m *Model) updateParametersLoadedMessage(msg messages.ParametersLoadedMsg) {
	if msg.CloseDetails && msg.Err == nil {
		m.detailsVisible = false
		m.details = m.details.SetData(nil)
		if m.active == panels.Details {
			m.setActive(panels.Parameters)
		}
	}
	if !m.dialog.IsOpen() && len(m.dialogQueue) > 0 {
		next := m.dialogQueue[0]
		m.dialogQueue = m.dialogQueue[1:]
		m.openDraftDialog(next.project, next.mode, nil)
	}
	m.closeRenameIfOrphaned()
	m.closeBoolPickerIfOrphaned()
	m.closeJSONInputIfOrphaned()
	m.closeNumberInputIfOrphaned()
	m.closeStringInputIfOrphaned()
	m.closeMoveIfOrphaned()
}

func (m *Model) updateWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	if !m.logsSized {
		m.logsHeight = initialLogsPanelHeight(msg.Height)
		m.logsSized = true
	}
	m.applyLayout()
}

func (m Model) updatePasteMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	if m.active == panels.Details && m.detailsVisible && m.details.FieldActive() {
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		return m, cmd, true
	}
	return m, nil, false
}

func (m Model) updatePanelMouseMessage(msg tea.MouseMsg) (Model, tea.Cmd, bool) {
	if m.active == panels.Logs {
		return m, nil, false
	}
	panel, ok := m.panelAt(msg.Mouse().X, msg.Mouse().Y)
	if !ok {
		return m, nil, false
	}
	m.setActive(panel)
	var cmd tea.Cmd
	switch panel {
	case panels.Projects:
		m.projects, cmd = m.projects.Update(msg)
		if m.width > 0 && m.height > 0 {
			m.applyLayout()
		}
	case panels.Parameters, panels.History:
		m.parameters, cmd = m.parameters.Update(msg)
	case panels.Conditions:
		m.conditions, cmd = m.conditions.Update(msg)
	case panels.Details:
		m.details, cmd = m.details.Update(msg)
	default:
		return m, nil, false
	}
	return m, cmd, true
}

func (m Model) updateKeyMessage(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	k := msg.String()
	if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionForceQuit, k) {
		return m, tea.Quit, true
	}
	if m.active == panels.Details && m.detailsVisible {
		if next, cmd, ok := m.updateDetailsKeyMessage(msg, k); ok {
			return next, cmd, true
		}
	}
	if !m.keyboardCaptured() {
		return m.updateGlobalKeyMessage(k)
	}
	return m, nil, false
}

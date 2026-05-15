package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/messages"
	"fbrcm/tui/panels"
)

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dialog.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				m.closeDialog(false)
				return m, nil
			}
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}
	}

	if m.boolPicker.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.closeBoolPicker()
				return m, nil
			case "ctrl+y":
				if value, ok := m.boolPicker.CurrentString(); ok {
					return m, copyToClipboardCmd(value)
				}
				return m, nil
			case "enter":
				return m, m.submitBoolPicker()
			case "up", "k":
				m.boolPicker.Move(-1)
				return m, nil
			case "down", "j":
				m.boolPicker.Move(1)
				return m, nil
			}
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	if m.jsonInput.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.closeJSONInput()
				return m, nil
			case "ctrl+y":
				return m, copyToClipboardCmd(m.jsonInput.PrettyValue())
			case "ctrl+f":
				if m.jsonInput.Valid() {
					m.jsonInput = m.jsonInput.Reformat()
				}
				return m, nil
			case "ctrl+s":
				if m.jsonInput.Valid() {
					return m, m.submitJSONInput()
				}
				return m, nil
			case "ctrl+enter":
				if m.jsonInput.Valid() {
					return m, m.submitJSONInput()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.jsonInput, cmd = m.jsonInput.Update(msg)
			return m, cmd
		case tea.PasteMsg, tea.ClipboardMsg:
			var cmd tea.Cmd
			m.jsonInput, cmd = m.jsonInput.Update(msg)
			return m, cmd
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	if m.numberInput.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.closeNumberInput()
				return m, nil
			case "ctrl+y":
				return m, copyToClipboardCmd(m.numberInput.Value())
			case "enter":
				if m.numberInput.Valid() {
					return m, m.submitNumberInput()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.numberInput, cmd = m.numberInput.Update(msg)
			if m.valueEditSource == panels.Details {
				m.details = m.details.SetValuesInvalid(!m.numberInput.Valid())
			}
			return m, cmd
		case tea.PasteMsg, tea.ClipboardMsg:
			var cmd tea.Cmd
			m.numberInput, cmd = m.numberInput.Update(msg)
			if m.valueEditSource == panels.Details {
				m.details = m.details.SetValuesInvalid(!m.numberInput.Valid())
			}
			return m, cmd
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	if m.stringInput.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.closeStringInput()
				return m, nil
			case "ctrl+y":
				return m, copyToClipboardCmd(m.stringInput.Value())
			case "f4":
				return m, m.toggleStringInputMode()
			case "ctrl+s":
				return m, m.submitStringInput()
			case "ctrl+enter":
				return m, m.submitStringInput()
			case "enter":
				if !m.stringInput.IsExpanded() {
					return m, m.submitStringInput()
				}
			}
			var cmd tea.Cmd
			m.stringInput, cmd = m.stringInput.Update(msg)
			return m, cmd
		case tea.PasteMsg, tea.ClipboardMsg:
			var cmd tea.Cmd
			m.stringInput, cmd = m.stringInput.Update(msg)
			return m, cmd
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	if m.moveParam.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.closeMoveParam()
				return m, nil
			case "enter":
				if _, ok := m.moveParam.Current(); ok {
					return m, m.submitMoveParam()
				}
				return m, nil
			case "up", "k":
				return m, m.moveParam.Move(-1)
			case "down", "j":
				return m, m.moveParam.Move(1)
			}
			if m.moveParam.InputSelected() {
				return m, m.moveParam.Update(msg)
			}
			if m.moveParam.Typeahead(msg.String(), time.Now()) {
				return m, nil
			}
		case tea.PasteMsg, tea.ClipboardMsg:
			if m.moveParam.InputSelected() {
				return m, m.moveParam.Update(msg)
			}
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	if m.renameInput.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				return m, m.cancelRenameInput()
			case "enter":
				return m, m.submitRenameInput()
			}
			var cmd tea.Cmd
			m.renameInput, cmd = m.renameInput.Update(msg)
			return m, cmd
		case tea.PasteMsg, tea.ClipboardMsg:
			var cmd tea.Cmd
			m.renameInput, cmd = m.renameInput.Update(msg)
			return m, cmd
		case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case messages.KeyboardCaptureMsg:
		if msg.Enabled {
			m.capture = m.active
		} else {
			m.capture = panels.None
		}

	case messages.SetActivePanelMsg:
		m.setActive(msg.Panel)

	case messages.DialogCanceledMsg:
		m.dialogQueue = nil
		m.pendingDetails = nil

	case messages.DetailsEditCanceledMsg:
		if msg.CloseDetails {
			m.pendingDetails = nil
			m.closeDetailsPanel()
			return m, nil
		}
		if m.pendingDetails != nil {
			m.newParameter = nil
			m.parameters.ClearTransientNewParameter()
			m.applyPendingDetailsSelection()
		}

	case messages.DetailsInvalidFixMsg:
		if m.pendingDetails != nil && m.newParameter != nil {
			m.newParameter = nil
			m.parameters.ClearTransientNewParameter()
		}
		m.pendingDetails = nil
		if data := m.details.Data(); data != nil {
			m.parameters.FocusParameter(data.Project.ProjectID, data.GroupKey, data.Parameter.Key)
		}
		if m.detailsVisible {
			m.setActive(panels.Details)
		}

	case messages.DetailsInvalidDiscardMsg:
		if msg.CloseDetails {
			m.pendingDetails = nil
			m.closeDetailsPanel()
			return m, nil
		}
		if m.pendingDetails != nil {
			m.newParameter = nil
			m.parameters.ClearTransientNewParameter()
			m.applyPendingDetailsSelection()
			return m, nil
		}
		if data := m.details.Data(); data != nil {
			m.details = m.details.SetData(data)
			m.setActive(panels.Details)
		}

	case messages.DetailsValueEditRequestedMsg:
		if m.active == panels.Details && m.detailsVisible && m.details.ValueSelected() {
			return m, m.openDetailsValueEditor()
		}

	case messages.ParameterSelectionChangedMsg:
		return m, m.handleParameterSelection(msg)

	case messages.ParametersLoadedMsg:
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.logsSized {
			m.logsHeight = initialLogsPanelHeight(msg.Height)
			m.logsSized = true
		}
		m.applyLayout()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.active == panels.Details && m.detailsVisible {
			switch msg.String() {
			case "tab":
				m.setActive(m.nextTabPanel())
				return m, nil
			case "esc":
				if m.details.FieldActive() || m.details.ValueSelected() {
					var cmd tea.Cmd
					m.details, cmd = m.details.Update(msg)
					return m, cmd
				}
				return m, m.requestCloseDetails()
			case "ctrl+enter":
				return m, m.submitDetailsForm()
			case "right", "f4":
				if m.details.ValueSelected() {
					return m, m.openDetailsValueEditor()
				}
			}
			if m.details.FieldActive() {
				var cmd tea.Cmd
				m.details, cmd = m.details.Update(msg)
				return m, cmd
			}
		}
		if !m.keyboardCaptured() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "esc":
				if m.active == panels.Details && m.detailsVisible {
					m.detailsVisible = false
					m.setActive(panels.Parameters)
				}
			case "1":
				m.setActive(panels.Projects)
			case "2":
				m.setActive(panels.Parameters)
			case "3":
				if m.detailsVisible {
					m.setActive(panels.Details)
				}
			case "0":
				m.setActive(panels.Logs)
			case "f9":
				if m.active == panels.Projects {
					m.toggleProjectsMode()
				}
				if m.active == panels.Logs {
					m.toggleLogsMode()
				}
			case "f11":
				if m.active == panels.Parameters {
					m.toggleParametersMaximize()
				}
			case "=", "+":
				if m.active == panels.Logs {
					if m.logsMode == logsPanelModeCollapsed {
						m.growLogsFromCollapsed()
						break
					}
					m.resizeLogsHeight(1)
				}
			case "-", "_":
				if m.active == panels.Logs {
					m.resizeLogsHeight(-1)
				}
			case "tab":
				m.setActive(m.nextTabPanel())
			case "f8":
				if m.active == panels.Parameters {
					project, groupKey, groupLabel, ok := m.parameters.CurrentGroupRef()
					if ok {
						if m.parameters.HasDraft(project.ProjectID) {
							return m, m.deleteGroupCmd(project, groupKey, false)
						}
						m.openDeleteGroupDialog(project, groupKey, groupLabel)
						return m, nil
					}
					project, groupKey, paramKey, ok := m.parameters.CurrentParameterRef()
					if ok {
						if m.parameters.HasDraft(project.ProjectID) {
							return m, m.deleteParameterCmd(project, groupKey, paramKey, false, false)
						}
						m.openDeleteDialog(project, groupKey, paramKey, false)
						return m, nil
					}
				}
				if m.active == panels.Details && m.detailsVisible {
					data := m.details.Data()
					if data != nil {
						if m.parameters.HasDraft(data.Project.ProjectID) {
							return m, m.deleteParameterCmd(data.Project, data.GroupKey, data.Parameter.Key, false, true)
						}
						m.openDeleteDialog(data.Project, data.GroupKey, data.Parameter.Key, true)
						x, y, width, height := m.details.Bounds()
						m.dialog = m.dialog.CenterWithin(x, y, width, height)
						return m, nil
					}
				}
			case "f2":
				if m.active == panels.Parameters {
					return m, m.openRenameInput()
				}
			case "f5":
				if m.active == panels.Parameters {
					return m, m.openDuplicateInput()
				}
			case "shift+f4":
				if m.active == panels.Parameters {
					return m, m.openNewParameterDetails()
				}
			case "f4":
				if m.active == panels.Parameters {
					if _, ok := m.parameters.CurrentBoolValueAnchor(); ok {
						return m, m.openBoolPicker()
					}
					if _, ok := m.parameters.CurrentNumberValueAnchor(); ok {
						return m, m.openNumberInput()
					}
					if _, ok := m.parameters.CurrentJSONValueAnchor(); ok {
						return m, m.openJSONInput()
					}
					if _, ok := m.parameters.CurrentStringValueAnchor(); ok {
						return m, m.openStringInput()
					}
					if m.parameters.FocusCurrentParameterDefaultValue() {
						if _, ok := m.parameters.CurrentBoolValueAnchor(); ok {
							return m, m.openBoolPicker()
						}
						if _, ok := m.parameters.CurrentNumberValueAnchor(); ok {
							return m, m.openNumberInput()
						}
						if _, ok := m.parameters.CurrentJSONValueAnchor(); ok {
							return m, m.openJSONInput()
						}
						if _, ok := m.parameters.CurrentStringValueAnchor(); ok {
							return m, m.openStringInput()
						}
					}
					return m, nil
				}
			case "f6":
				if m.active == panels.Parameters {
					return m, m.openMoveParam()
				}
			case "p":
				if m.active == panels.Parameters {
					project, ok := m.parameters.CurrentProject()
					if ok && m.parameters.HasDraft(project.ProjectID) {
						m.openDraftDialog(project, dialogModePublishDraft, nil)
						return m, nil
					}
				}
			case "P":
				if m.active == panels.Parameters {
					projects := m.parameters.DraftProjects()
					if len(projects) > 0 {
						queue := make([]pendingDialog, 0, len(projects)-1)
						for _, project := range projects[1:] {
							queue = append(queue, pendingDialog{project: project, mode: dialogModePublishDraft})
						}
						m.openDraftDialog(projects[0], dialogModePublishDraft, queue)
						return m, nil
					}
				}
			case "d":
				if m.active == panels.Parameters {
					project, ok := m.parameters.CurrentProject()
					if ok && m.parameters.HasDraft(project.ProjectID) {
						m.openDraftDialog(project, dialogModeDiscardDraft, nil)
						return m, nil
					}
				}
			case "D":
				if m.active == panels.Parameters {
					projects := m.parameters.DraftProjects()
					if len(projects) > 0 {
						queue := make([]pendingDialog, 0, len(projects)-1)
						for _, project := range projects[1:] {
							queue = append(queue, pendingDialog{project: project, mode: dialogModeDiscardDraft})
						}
						m.openDraftDialog(projects[0], dialogModeDiscardDraft, queue)
						return m, nil
					}
				}
			}
		}

	case tea.PasteMsg, tea.ClipboardMsg:
		if m.active == panels.Details && m.detailsVisible && m.details.FieldActive() {
			var cmd tea.Cmd
			m.details, cmd = m.details.Update(msg)
			return m, cmd
		}

	case tea.MouseClickMsg:
		if m.active == panels.Logs {
			break
		}
		if panel, ok := m.panelAt(msg.Mouse().X, msg.Mouse().Y); ok {
			m.setActive(panel)
			if panel == panels.Details {
				var cmd tea.Cmd
				m.details, cmd = m.details.Update(msg)
				return m, cmd
			}
		}

	case tea.MouseWheelMsg:
		if m.active == panels.Logs {
			break
		}
		if panel, ok := m.panelAt(msg.Mouse().X, msg.Mouse().Y); ok {
			m.setActive(panel)
			if panel == panels.Details {
				var cmd tea.Cmd
				m.details, cmd = m.details.Update(msg)
				return m, cmd
			}
		}

	case tea.MouseMotionMsg:
	case tea.MouseReleaseMsg:
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.projects, cmd = m.projects.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.WindowSizeMsg); !ok && m.width > 0 && m.height > 0 {
		m.applyLayout()
	}

	m.parameters, cmd = m.parameters.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.closeDetailsIfOrphaned()

	m.details, cmd = m.details.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.logs, cmd = m.logs.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if loadedMsg, ok := msg.(messages.ParametersLoadedMsg); ok && loadedMsg.Err == nil {
		if loadedMsg.DetailsSaved {
			if m.newParameter != nil && m.newParameter.projectID == loadedMsg.Project.ProjectID {
				m.newParameter = nil
				if loadedMsg.SelectParamKey != "" {
					m.parameters.ClearTransientNewParameterAndFocus(loadedMsg.Project.ProjectID, loadedMsg.SelectGroupKey, loadedMsg.SelectParamKey)
				} else {
					m.parameters.ClearTransientNewParameter()
				}
			}
			if loadedMsg.CloseDetails {
				m.pendingDetails = nil
			} else if m.pendingDetails != nil {
				m.applyPendingDetailsSelection()
			} else {
				m.details = m.details.MarkSaved()
			}
		} else if m.detailsVisible && loadedMsg.SelectParamKey != "" {
			if data, ok := m.parameters.CurrentParameterViewData(); ok && data.Project.ProjectID == loadedMsg.Project.ProjectID {
				m.details = m.details.SetData(data)
			}
		}
		if m.duplicate != nil && m.duplicate.project.ProjectID == loadedMsg.Project.ProjectID {
			m.duplicate = nil
			if loadedMsg.SelectParamKey != "" {
				m.parameters.ClearTransientDuplicateAndFocus(loadedMsg.Project.ProjectID, loadedMsg.SelectGroupKey, loadedMsg.SelectParamKey)
			} else {
				m.parameters.ClearTransientDuplicate()
			}
			m.closeRenameInput()
		}
	}

	return m, tea.Batch(cmds...)
}

// closeDetailsIfOrphaned closes close details if orphaned for Model and returns the resulting state or error.
func (m *Model) closeDetailsIfOrphaned() {
	if !m.detailsVisible {
		return
	}
	data := m.details.Data()
	if data == nil {
		return
	}
	if m.parameters.HasProject(data.Project.ProjectID) {
		return
	}
	m.closeDetailsPanel()
}

// applyLayout handles apply layout for Model and returns the resulting state or error.
func (m *Model) applyLayout() {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)

	m.projects = m.projects.SetCollapsed(m.projectsMode == projectsPanelModeCollapsed)
	m.projects = m.projects.SetBounds(0, 0, layout.leftWidth, layout.topHeight)
	m.parameters = m.parameters.SetBounds(layout.leftWidth, 0, layout.rightWidth, layout.topHeight)
	m.dialog = m.dialog.SetBounds(0, 0, m.width, m.height)
	detailsWidth := m.detailsWidthForLayout(layout)
	m.details = m.details.SetBounds(layout.bottomWidth-detailsWidth, 0, detailsWidth, layout.topHeight)
	m.logs = m.logs.SetBounds(0, layout.topHeight, layout.bottomWidth, layout.bottomHeight)
}

// nextTabPanel handles next tab panel for Model and returns the resulting state or error.
func (m Model) nextTabPanel() panels.ID {
	if m.active == panels.Logs {
		if m.detailsVisible {
			if m.prevTop == panels.Details || m.prevTop == panels.Parameters {
				return m.prevTop
			}
			return panels.Parameters
		}
		return m.prevTop
	}

	if m.detailsVisible {
		if m.active == panels.Details {
			return panels.Parameters
		}
		if m.active == panels.Parameters {
			return panels.Details
		}
		return panels.Parameters
	}

	if m.active == panels.Parameters {
		return panels.Projects
	}

	return panels.Parameters
}

// setActive sets set active for Model and returns the resulting state or error.
func (m *Model) setActive(panel panels.ID) {
	if panel != panels.Logs {
		m.prevTop = panel
	}
	m.active = panel
	if m.capture != panels.None && m.capture != panel {
		m.capture = panels.None
	}
	m.projects = m.projects.SetActive(panel == panels.Projects)
	m.parameters = m.parameters.SetActive(panel == panels.Parameters)
	m.details = m.details.SetActive(panel == panels.Details)
	m.details = m.details.SetBridgeActive(panel == panels.Parameters)
	m.logs = m.logs.SetActive(panel == panels.Logs)
}

// keyboardCaptured handles keyboard captured for Model and returns the resulting state or error.
func (m Model) keyboardCaptured() bool {
	return m.capture != panels.None
}

// panelAt handles panel at for Model and returns the resulting state or error.
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
		return panels.Parameters, true
	}

	return panels.None, false
}

// resizeLogsHeight handles resize logs height for Model and returns the resulting state or error.
func (m *Model) resizeLogsHeight(delta int) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	m.logsHeight = nextLogsPanelHeight(layout.bottomHeight, delta)
	m.logsHeight = min(m.logsHeight, newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode).bottomHeight)
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

// nextLogsPanelHeight handles next logs panel height and returns the resulting value or error.
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

// initialLogsPanelHeight initializes initial logs panel height and returns the resulting value or error.
func initialLogsPanelHeight(terminalHeight int) int {
	if terminalHeight <= 35 {
		return collapsedLogsPanelHeight
	}
	if terminalHeight >= 40 {
		return defaultLogsPanelHeight
	}
	return terminalHeight - 33
}

// detailsWidthForLayout handles details width for layout for Model and returns the resulting state or error.
func (m Model) detailsWidthForLayout(layout panelLayout) int {
	minWidth := max(layout.rightWidth/2, 1)
	maxWidth := max(layout.rightWidth-11, 1)

	// Details content width = panel width - 5:
	// bridge spacer, left border, left padding, right padding, scrollbar lane.
	nameFitWidth := m.parameters.LongestParameterNameWidth() + 5

	desired := max(minWidth, nameFitWidth)
	return min(desired, maxWidth)
}

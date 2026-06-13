package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dialog.IsOpen() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionCancel, msg.String()) {
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionCancel, k):
				m.closeBoolPicker()
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionCopyValue, k):
				if value, ok := m.boolPicker.CurrentString(); ok {
					return m, copyToClipboardCmd(value)
				}
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionSubmit, k):
				return m, m.submitBoolPicker()
			case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionUp, k):
				m.boolPicker.Move(-1)
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionDown, k):
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockJSONInput, tuiconfig.ActionCancel, k):
				m.closeJSONInput()
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockJSONInput, tuiconfig.ActionCopyValue, k):
				return m, copyToClipboardCmd(m.jsonInput.PrettyValue())
			case tuiconfig.Matches(tuiconfig.BlockJSONInput, tuiconfig.ActionFormat, k):
				if m.jsonInput.Valid() {
					m.jsonInput = m.jsonInput.Reformat()
				}
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockJSONInput, tuiconfig.ActionSave, k):
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockNumberInput, tuiconfig.ActionCancel, k):
				m.closeNumberInput()
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockNumberInput, tuiconfig.ActionCopyValue, k):
				return m, copyToClipboardCmd(m.numberInput.Value())
			case tuiconfig.Matches(tuiconfig.BlockNumberInput, tuiconfig.ActionSubmit, k):
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionCancel, k):
				m.closeStringInput()
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionCopyValue, k):
				return m, copyToClipboardCmd(m.stringInput.Value())
			case tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionToggleExpanded, k):
				return m, m.toggleStringInputMode()
			case tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionSave, k):
				return m, m.submitStringInput()
			case tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionSubmit, k):
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionCancel, k):
				m.closeMoveParam()
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionSubmit, k):
				if _, ok := m.moveParam.Current(); ok {
					return m, m.submitMoveParam()
				}
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionUp, k):
				return m, m.moveParam.Move(-1)
			case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionDown, k):
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
			k := msg.String()
			switch {
			case tuiconfig.Matches(tuiconfig.BlockRenameInput, tuiconfig.ActionCancel, k):
				return m, m.cancelRenameInput()
			case tuiconfig.Matches(tuiconfig.BlockRenameInput, tuiconfig.ActionSubmit, k):
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
		k := msg.String()
		if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionForceQuit, k) {
			return m, tea.Quit
		}
		if m.active == panels.Details && m.detailsVisible {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusNext, k):
				m.setActive(m.nextTabPanel())
				return m, nil
			case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionClose, k):
				if m.details.FieldActive() || m.details.ValueSelected() {
					var cmd tea.Cmd
					m.details, cmd = m.details.Update(msg)
					return m, cmd
				}
				return m, m.requestCloseDetails()
			case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionSubmit, k):
				return m, m.submitDetailsForm()
			case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, k):
				if m.details.ValueSelected() {
					return m, m.openDetailsValueEditor()
				}
			}
			if !m.details.TextInputActive() {
				switch {
				case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
					return m, tea.Quit
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionMove, k):
					return m, m.activateDetailsGroup()
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionRename, k):
					var cmd tea.Cmd
					m.details, cmd = m.details.ActivateName()
					return m, cmd
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyName, k):
					return m, m.copyDetailsNameCmd()
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath, k):
					return m, m.copyDetailsPathCmd()
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, k):
					if m.details.ValueSelected() {
						return m, m.copyDetailsSelectedValueCmd()
					}
				case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionDelete, k):
					return m, m.requestDeleteDetails()
				}
			}
			if m.details.FieldActive() {
				var cmd tea.Cmd
				m.details, cmd = m.details.Update(msg)
				return m, cmd
			}
		}
		if !m.keyboardCaptured() {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
				return m, tea.Quit
			case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionClose, k):
				if m.active == panels.Details && m.detailsVisible {
					m.detailsVisible = false
					m.setActive(panels.Parameters)
				}
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusProjects, k):
				m.setActive(panels.Projects)
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusParameters, k):
				m.setActive(panels.Parameters)
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusDetails, k):
				if m.detailsVisible {
					m.setActive(panels.Details)
				}
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusLogs, k):
				m.setActive(panels.Logs)
			case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionToggleMode, k), tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionToggleMode, k), tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDuplicate, k):
				if m.active == panels.Projects {
					m.toggleProjectsMode()
				}
				if m.active == panels.Logs {
					m.toggleLogsMode()
				}
				if m.active == panels.Parameters {
					return m, m.openDuplicateInput()
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionToggleMaximize, k):
				if m.active == panels.Parameters {
					m.toggleParametersMaximize()
				}
			case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionResizeGrow, k):
				if m.active == panels.Logs {
					if m.logsMode == logsPanelModeCollapsed {
						m.growLogsFromCollapsed()
						break
					}
					m.resizeLogsHeight(1)
				}
			case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionResizeShrink, k):
				if m.active == panels.Logs {
					m.resizeLogsHeight(-1)
				}
			case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusNext, k):
				m.setActive(m.nextTabPanel())
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDelete, k), tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionDelete, k):
				if m.active == panels.Parameters {
					if anchor, ok := m.parameters.CurrentConditionalValueAnchor(); ok {
						if m.parameters.HasDraft(anchor.Project.ProjectID) {
							return m, m.deleteConditionalValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, false)
						}
						m.openDeleteConditionalValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel)
						return m, nil
					}
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
					return m, m.requestDeleteDetails()
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionRename, k):
				if m.active == panels.Parameters {
					return m, m.openRenameInput()
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionNew, k):
				if m.active == panels.Parameters {
					return m, m.openNewParameterDetails()
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionEdit, k):
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
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionMove, k):
				if m.active == panels.Parameters {
					return m, m.openMoveParam()
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionPublish, k):
				if m.active == panels.Parameters {
					project, ok := m.parameters.CurrentProject()
					if ok && m.parameters.HasDraft(project.ProjectID) {
						m.openDraftDialog(project, dialogModePublishDraft, nil)
						return m, nil
					}
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionPublishAll, k):
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
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDiscard, k):
				if m.active == panels.Parameters {
					project, ok := m.parameters.CurrentProject()
					if ok && m.parameters.HasDraft(project.ProjectID) {
						m.openDraftDialog(project, dialogModeDiscardDraft, nil)
						return m, nil
					}
				}
			case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDiscardAll, k):
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

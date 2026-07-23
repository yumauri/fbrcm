package app

import (
	"fmt"
	"slices"
	"strings"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) helpPaletteActions() []helpPaletteAction {
	actions := helpPaletteCatalog()
	active, _ := m.activeHelpBlock()
	for i := range actions {
		actions[i].keys = tuiconfig.Keys(actions[i].block, actions[i].action)
		if sharedHelpActionAvailableIn(actions[i], active) {
			actions[i].group = helpPaletteBlockTitle(active)
		}
		actions[i].enabled, actions[i].reason = m.helpPaletteActionAvailability(actions[i])
		if len(actions[i].keys) == 0 {
			actions[i].enabled = false
			actions[i].reason = "shortcut is unbound or has a key conflict"
		}
	}
	slices.SortStableFunc(actions, func(left, right helpPaletteAction) int {
		leftRank := helpPaletteGroupRank(left, active)
		rightRank := helpPaletteGroupRank(right, active)
		if leftRank != rightRank {
			return leftRank - rightRank
		}
		if leftRank == 2 {
			return strings.Compare(left.group, right.group)
		}
		return 0
	})
	return actions
}

func helpPaletteGroupRank(item helpPaletteAction, active tuiconfig.Block) int {
	if active != "" && item.group == helpPaletteBlockTitle(active) {
		return 0
	}
	if item.block == tuiconfig.BlockGlobal {
		return 1
	}
	return 2
}

func (m Model) helpPaletteActionAvailability(item helpPaletteAction) (bool, string) {
	if item.block == tuiconfig.BlockHelp {
		return true, ""
	}
	if item.block == tuiconfig.BlockGlobal {
		return m.globalHelpActionAvailability(item.action)
	}
	if item.block == tuiconfig.BlockAccounts || item.block == tuiconfig.BlockProfiles {
		return m.setup.HelpActionAvailability(item.block, item.action)
	}
	if item.block == tuiconfig.BlockFilter {
		return m.filterHelpActionAvailability(item.action)
	}

	if active, reason := m.activeHelpBlock(); item.block != active {
		if sharedHelpActionAvailableIn(item, active) {
			return m.sharedHelpActionAvailability(active, item.action)
		}
		return false, reasonForInactiveHelpBlock(item.block, active, reason)
	}
	return m.contextualHelpActionAvailability(item.block, item.action)
}

func sharedHelpActionAvailableIn(item helpPaletteAction, active tuiconfig.Block) bool {
	if item.block != tuiconfig.BlockParameters {
		return false
	}
	switch active {
	case tuiconfig.BlockConditions:
		switch item.action {
		case tuiconfig.ActionToggleMaximize, tuiconfig.ActionReload, tuiconfig.ActionReloadAll:
			return true
		}
	case tuiconfig.BlockHistory:
		switch item.action {
		case tuiconfig.ActionToggleMaximize, tuiconfig.ActionToggle, tuiconfig.ActionCopyName, tuiconfig.ActionCopyPath:
			return true
		}
	case tuiconfig.BlockPromote:
		return item.action == tuiconfig.ActionToggleMaximize
	}
	return false
}

func (m Model) sharedHelpActionAvailability(active tuiconfig.Block, action tuiconfig.Action) (bool, string) {
	if active == tuiconfig.BlockHistory {
		return m.parametersHelpActionAvailability(action)
	}
	if active == tuiconfig.BlockPromote && action == tuiconfig.ActionToggleMaximize {
		return true, ""
	}
	if action == tuiconfig.ActionToggleMaximize {
		return true, ""
	}
	if _, ok := m.conditions.CurrentProject(); !ok {
		if action == tuiconfig.ActionReloadAll {
			return false, "no projects are selected"
		}
		return false, "no project is selected"
	}
	return true, ""
}

func (m Model) globalHelpActionAvailability(action tuiconfig.Action) (bool, string) {
	switch action {
	case tuiconfig.ActionAccounts, tuiconfig.ActionProfiles:
		if m.setup.IsOpen() {
			if active, ok := m.setup.HelpBlock(); ok {
				target := tuiconfig.BlockAccounts
				if action == tuiconfig.ActionProfiles {
					target = tuiconfig.BlockProfiles
				}
				if active == target {
					return false, helpPaletteBlockTitle(target) + " is already active"
				}
				return true, ""
			}
		}
		if m.details.Dirty() {
			return false, "save or discard the open Details changes first"
		}
		if m.contextOverlayOpen() {
			return false, "close the current dialog or editor first"
		}
		if m.keyboardCaptured() {
			return false, "finish or close the current filter first"
		}
		return true, ""
	case tuiconfig.ActionHelp:
		return true, ""
	case tuiconfig.ActionForceQuit:
		return true, ""
	case tuiconfig.ActionQuit:
		if m.setup.IsOpen() {
			if _, ok := m.setup.HelpBlock(); ok {
				return true, ""
			}
		}
	case tuiconfig.ActionFocusDetails:
		if m.promote.WorkspaceOpen() {
			return false, "the Promote panel replaces the workspace panels"
		}
		if !m.detailsVisible {
			return false, "details panel is not open"
		}
	case tuiconfig.ActionFocusParameters, tuiconfig.ActionFocusConditions, tuiconfig.ActionFocusHistory:
		if m.promote.WorkspaceOpen() {
			return false, "the Promote panel replaces the workspace panels"
		}
	case tuiconfig.ActionFocusPromote:
		if !m.promote.WorkspaceOpen() {
			return false, "the Promote panel is not open"
		}
	}
	if m.contextOverlayOpen() {
		return false, "close the current dialog or editor first"
	}
	if m.keyboardCaptured() {
		return false, "finish or close the current filter first"
	}
	return true, ""
}

func (m Model) filterHelpActionAvailability(action tuiconfig.Action) (bool, string) {
	if m.contextOverlayOpen() {
		return false, "close the current dialog or editor first"
	}
	filterPanel := m.active == panels.Projects || m.active == panels.Parameters || m.active == panels.Conditions || m.active == panels.History || m.active == panels.Promote
	if !filterPanel {
		return false, "the focused panel does not support filtering"
	}
	filterOpen := m.keyboardCaptured()
	if action == tuiconfig.ActionFilterApply || action == tuiconfig.ActionFilterCancel || action == tuiconfig.ActionFilterUp || action == tuiconfig.ActionFilterDown {
		if !filterOpen {
			return false, "no panel filter is open"
		}
		return true, ""
	}
	if filterOpen {
		return false, "a panel filter is already open"
	}
	return true, ""
}

func (m Model) activeHelpBlock() (tuiconfig.Block, string) {
	switch {
	case m.setup.IsOpen():
		if block, ok := m.setup.HelpBlock(); ok {
			return block, ""
		}
		return "", "account or profile workflow is active"
	case m.keyboardCaptured():
		return tuiconfig.BlockFilter, "panel filter has keyboard focus"
	case m.parameters.HistoryPickerOpen():
		return tuiconfig.BlockHistoryPicker, "history version picker is open"
	case m.conditions.MoveActive(), m.moveParam.IsOpen():
		return tuiconfig.BlockMoveInput, "move editor is open"
	case m.authPicker.IsOpen():
		return tuiconfig.BlockAuthPicker, "authentication picker is open"
	case m.dialog.IsOpen():
		return tuiconfig.BlockDialog, "confirmation dialog is open"
	case m.diffView.IsOpen():
		return tuiconfig.BlockDiffView, "diff viewer is open"
	case m.boolPicker.IsOpen():
		return tuiconfig.BlockBoolInput, "boolean editor is open"
	case m.jsonInput.IsOpen():
		return tuiconfig.BlockJSONInput, "JSON editor is open"
	case m.numberInput.IsOpen():
		return tuiconfig.BlockNumberInput, "number editor is open"
	case m.stringInput.IsOpen():
		return tuiconfig.BlockStringInput, "string editor is open"
	case m.renameInput.IsOpen():
		return tuiconfig.BlockRenameInput, "rename editor is open"
	case m.active == panels.Details && m.detailsVisible && m.details.FieldActive():
		return tuiconfig.BlockDetailsForm, "details form has keyboard focus"
	}
	switch m.active {
	case panels.Projects:
		return tuiconfig.BlockProjects, ""
	case panels.Parameters:
		return tuiconfig.BlockParameters, ""
	case panels.Conditions:
		return tuiconfig.BlockConditions, ""
	case panels.History:
		return tuiconfig.BlockHistory, ""
	case panels.Promote:
		return tuiconfig.BlockPromote, ""
	case panels.Details:
		return tuiconfig.BlockDetails, ""
	case panels.Logs:
		return tuiconfig.BlockLogs, ""
	default:
		return "", ""
	}
}

func reasonForInactiveHelpBlock(block, active tuiconfig.Block, activeReason string) string {
	if active == tuiconfig.BlockFilter {
		return "unavailable while " + activeReason
	}
	if activeReason != "" && isOverlayHelpBlock(active) {
		return "unavailable while " + activeReason
	}
	if isOverlayHelpBlock(block) {
		return helpPaletteBlockTitle(block) + " is not open"
	}
	return fmt.Sprintf("focus the %s first", helpPaletteBlockTitle(block))
}

func isOverlayHelpBlock(block tuiconfig.Block) bool {
	switch block {
	case tuiconfig.BlockHistoryPicker, tuiconfig.BlockDetailsForm, tuiconfig.BlockDialog,
		tuiconfig.BlockBoolInput, tuiconfig.BlockDiffView, tuiconfig.BlockJSONInput, tuiconfig.BlockNumberInput,
		tuiconfig.BlockStringInput, tuiconfig.BlockMoveInput, tuiconfig.BlockAuthPicker, tuiconfig.BlockRenameInput:
		return true
	default:
		return false
	}
}

func (m Model) contextualHelpActionAvailability(block tuiconfig.Block, action tuiconfig.Action) (bool, string) {
	switch block {
	case tuiconfig.BlockProjects:
		if action == tuiconfig.ActionBindAuth && m.authCount <= 1 {
			return false, "at least two authentication identities are required"
		}
		if action == tuiconfig.ActionRefresh || action == tuiconfig.ActionToggleMode {
			return true, ""
		}
		if !m.projects.HasCurrentProject() {
			return false, "no project is selected"
		}
		if action == tuiconfig.ActionPromote && len(m.projects.AllProjects()) < 2 {
			return false, "at least two projects are required"
		}
		if action == tuiconfig.ActionBindAuth && !m.projects.AuthBindingAvailable() {
			return false, "at least two authentication identities must discover every selected project"
		}
		if !m.projects.CurrentProjectEnabled() {
			switch action {
			case tuiconfig.ActionSelect, tuiconfig.ActionMark, tuiconfig.ActionOpen,
				tuiconfig.ActionImport, tuiconfig.ActionExport, tuiconfig.ActionDefaults:
				return false, "project is disabled"
			}
		}
	case tuiconfig.BlockParameters:
		return m.parametersHelpActionAvailability(action)
	case tuiconfig.BlockConditions:
		return m.conditionsHelpActionAvailability(action)
	case tuiconfig.BlockPromote:
		switch action {
		case tuiconfig.ActionSaveDraft, tuiconfig.ActionPublish:
			if !m.promote.HasSelection() {
				return false, "no promotion changes are selected"
			}
			if action == tuiconfig.ActionPublish && !m.promote.CanPublish() {
				return false, "Firebase verification is required before publication"
			}
		}
	case tuiconfig.BlockHistory:
		if _, ok := m.parameters.CurrentProject(); !ok {
			return false, "no project history is selected"
		}
		if action == tuiconfig.ActionSubmit && !m.parameters.HistoryDiffAvailable() {
			return false, "select a property in a loaded version pair"
		}
	case tuiconfig.BlockDetails:
		return m.detailsHelpActionAvailability(action)
	}
	return true, ""
}

func (m Model) parametersHelpActionAvailability(action tuiconfig.Action) (bool, string) {
	project, projectOK := m.parameters.CurrentProject()
	_, _, _, paramOK := m.parameters.CurrentParameterRef()
	_, _, _, groupOK := m.parameters.CurrentGroupRef()
	_, conditionalOK := m.parameters.CurrentConditionalValueAnchor()
	_, renameOK := m.parameters.CurrentRenameAnchor()
	_, moveOK := m.parameters.CurrentMoveAnchor()
	valueOK := m.currentParameterValueSelected()

	switch action {
	case tuiconfig.ActionToggleMaximize:
		return true, ""
	case tuiconfig.ActionReloadAll:
		if !projectOK {
			return false, "no projects are selected"
		}
	case tuiconfig.ActionPublishAll, tuiconfig.ActionDiscardAll:
		if len(m.parameters.DraftProjects()) == 0 {
			return false, "there are no unpublished drafts"
		}
	case tuiconfig.ActionPublish, tuiconfig.ActionDiscard:
		if !projectOK {
			return false, "no project is selected"
		}
		if !m.parameters.HasDraft(project.ProjectID) {
			return false, "the selected project has no draft"
		}
	case tuiconfig.ActionRename:
		if !renameOK {
			return false, "select a parameter or parameter group"
		}
	case tuiconfig.ActionEdit:
		if !valueOK && !paramOK {
			return false, "select a parameter value"
		}
	case tuiconfig.ActionDuplicate, tuiconfig.ActionOpenDetails, tuiconfig.ActionCopyName, tuiconfig.ActionCopyPath:
		if !paramOK {
			return false, "select a parameter"
		}
	case tuiconfig.ActionMove:
		if !moveOK {
			return false, "select a parameter or parameter group"
		}
	case tuiconfig.ActionDelete:
		if !paramOK && !groupOK && !conditionalOK {
			return false, "select a parameter, group, or conditional value"
		}
	case tuiconfig.ActionToggle:
		if !paramOK {
			return false, "select a parameter"
		}
	case tuiconfig.ActionNew, tuiconfig.ActionReload, tuiconfig.ActionFirst, tuiconfig.ActionLast,
		tuiconfig.ActionNextGroup, tuiconfig.ActionPrevGroup, tuiconfig.ActionExpandAll,
		tuiconfig.ActionCollapseAll, tuiconfig.ActionExpandGroups, tuiconfig.ActionCollapseGroups,
		tuiconfig.ActionExpand, tuiconfig.ActionCollapse, tuiconfig.ActionUp, tuiconfig.ActionDown:
		if !projectOK {
			return false, "no project is selected"
		}
	}
	return true, ""
}

func (m Model) currentParameterValueSelected() bool {
	if _, ok := m.parameters.CurrentBoolValueAnchor(); ok {
		return true
	}
	if _, ok := m.parameters.CurrentNumberValueAnchor(); ok {
		return true
	}
	if _, ok := m.parameters.CurrentStringValueAnchor(); ok {
		return true
	}
	_, ok := m.parameters.CurrentJSONValueAnchor()
	return ok
}

func (m Model) conditionsHelpActionAvailability(action tuiconfig.Action) (bool, string) {
	project, projectOK := m.conditions.CurrentProject()
	_, conditionOK := m.conditions.CurrentCondition()
	switch action {
	case tuiconfig.ActionPublishAll, tuiconfig.ActionDiscardAll:
		if len(m.parameters.DraftProjects()) == 0 {
			return false, "there are no unpublished drafts"
		}
	case tuiconfig.ActionPublish, tuiconfig.ActionDiscard:
		if !projectOK {
			return false, "no project is selected"
		}
		if !m.parameters.HasDraft(project.ProjectID) {
			return false, "the selected project has no draft"
		}
	case tuiconfig.ActionNew, tuiconfig.ActionReload:
		if !projectOK {
			return false, "no project is selected"
		}
	case tuiconfig.ActionReloadAll:
		if !projectOK {
			return false, "no projects are selected"
		}
	case tuiconfig.ActionToggleMaximize:
		return true, ""
	default:
		if !conditionOK {
			return false, "no condition is selected"
		}
	}
	return true, ""
}

func (m Model) detailsHelpActionAvailability(action tuiconfig.Action) (bool, string) {
	if !m.detailsVisible || (m.details.Data() == nil && m.details.ConditionData() == nil && m.details.GroupData() == nil) {
		return false, "details panel has no selected item"
	}
	if m.details.IsGroup() {
		switch action {
		case tuiconfig.ActionClose, tuiconfig.ActionSubmit, tuiconfig.ActionRename,
			tuiconfig.ActionDelete, tuiconfig.ActionCopyName, tuiconfig.ActionCopyPath:
			return true, ""
		default:
			return false, "parameter groups do not support this Details action"
		}
	}
	if m.details.IsCondition() {
		if action == tuiconfig.ActionNew {
			return false, "conditional values can only be added to parameters"
		}
		return true, ""
	}
	if action == tuiconfig.ActionColor {
		return false, "colors can only be edited for conditions"
	}
	if action == tuiconfig.ActionEditValue && !m.details.ValueSelected() {
		return false, "no parameter value is selected"
	}
	if action == tuiconfig.ActionCopyValue && !m.details.ValueSelected() && !m.details.UsageSelected() {
		return false, "no value or usage is selected"
	}
	return true, ""
}

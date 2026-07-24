package app

import (
	"slices"
	"strings"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type helpPaletteAction struct {
	block       tuiconfig.Block
	action      tuiconfig.Action
	group       string
	title       string
	description string
	keys        []string
	enabled     bool
	reason      string
}

var helpPaletteBlockOrder = []tuiconfig.Block{
	tuiconfig.BlockGlobal,
	tuiconfig.BlockAccounts,
	tuiconfig.BlockProfiles,
	tuiconfig.BlockProjects,
	tuiconfig.BlockParameters,
	tuiconfig.BlockConditions,
	tuiconfig.BlockPromote,
	tuiconfig.BlockHistory,
	tuiconfig.BlockDetails,
	tuiconfig.BlockLogs,
	tuiconfig.BlockFilter,
	tuiconfig.BlockHistoryPicker,
	tuiconfig.BlockDetailsForm,
	tuiconfig.BlockDialog,
	tuiconfig.BlockBoolInput,
	tuiconfig.BlockDiffView,
	tuiconfig.BlockJSONInput,
	tuiconfig.BlockNumberInput,
	tuiconfig.BlockStringInput,
	tuiconfig.BlockMoveInput,
	tuiconfig.BlockAuthPicker,
	tuiconfig.BlockRenameInput,
	tuiconfig.BlockHelp,
}

func helpPaletteCatalog() []helpPaletteAction {
	defaults := tuiconfig.DefaultKeyMap()
	out := make([]helpPaletteAction, 0)
	for _, block := range helpPaletteBlockOrder {
		actions := make([]tuiconfig.Action, 0, len(defaults[block]))
		for action := range defaults[block] {
			actions = append(actions, action)
		}
		slices.SortFunc(actions, func(left, right tuiconfig.Action) int {
			if byTitle := strings.Compare(helpPaletteActionTitle(block, left), helpPaletteActionTitle(block, right)); byTitle != 0 {
				return byTitle
			}
			return strings.Compare(string(left), string(right))
		})
		for _, action := range actions {
			title := helpPaletteActionTitle(block, action)
			out = append(out, helpPaletteAction{
				block:       block,
				action:      action,
				group:       helpPaletteBlockTitle(block),
				title:       title,
				description: helpPaletteActionDescription(block, action, title),
			})
		}
	}
	return out
}

func helpPaletteBlockTitle(block tuiconfig.Block) string {
	switch block {
	case tuiconfig.BlockGlobal:
		return "Global"
	case tuiconfig.BlockAccounts:
		return "Accounts panel"
	case tuiconfig.BlockProfiles:
		return "Profiles panel"
	case tuiconfig.BlockProjects:
		return "Projects panel"
	case tuiconfig.BlockParameters:
		return "Parameters panel"
	case tuiconfig.BlockConditions:
		return "Conditions panel"
	case tuiconfig.BlockPromote:
		return "Promote workspace"
	case tuiconfig.BlockHistory:
		return "History panel"
	case tuiconfig.BlockDetails:
		return "Details panel"
	case tuiconfig.BlockLogs:
		return "Logs panel"
	case tuiconfig.BlockFilter:
		return "Current panel filter"
	case tuiconfig.BlockHistoryPicker:
		return "History version picker"
	case tuiconfig.BlockDetailsForm:
		return "Details form"
	case tuiconfig.BlockDialog:
		return "Confirmation dialog"
	case tuiconfig.BlockBoolInput:
		return "Boolean editor"
	case tuiconfig.BlockDiffView:
		return "Diff viewer"
	case tuiconfig.BlockJSONInput:
		return "JSON editor"
	case tuiconfig.BlockNumberInput:
		return "Number editor"
	case tuiconfig.BlockStringInput:
		return "String editor"
	case tuiconfig.BlockMoveInput:
		return "Move editor"
	case tuiconfig.BlockAuthPicker:
		return "Authentication picker"
	case tuiconfig.BlockRenameInput:
		return "Rename editor"
	case tuiconfig.BlockHelp:
		return "Help palette"
	default:
		return titleWords(string(block))
	}
}

func helpPaletteActionTitle(block tuiconfig.Block, action tuiconfig.Action) string {
	if block == tuiconfig.BlockProjects {
		switch action {
		case tuiconfig.ActionToggleMode:
			return "Collapse or expand projects"
		case tuiconfig.ActionRefresh:
			return "Update projects"
		case tuiconfig.ActionSelect:
			return "Select project"
		case tuiconfig.ActionOpen:
			return "Open project in Firebase"
		case tuiconfig.ActionMark:
			return "Mark or unmark project"
		case tuiconfig.ActionDelete:
			return "Delete local project"
		case tuiconfig.ActionBindAuth:
			return "Bind authentication"
		case tuiconfig.ActionImport:
			return "Import Remote Config"
		case tuiconfig.ActionExport:
			return "Export Remote Config"
		case tuiconfig.ActionDefaults:
			return "Download application defaults"
		case tuiconfig.ActionPromote:
			return "Promote to another project"
		}
	}
	if block == tuiconfig.BlockParameters {
		switch action {
		case tuiconfig.ActionToggleMaximize:
			return "Maximize or restore workspace"
		case tuiconfig.ActionNew:
			return "Add new parameter"
		case tuiconfig.ActionNewGroup:
			return "Add new parameter group"
		case tuiconfig.ActionEdit:
			return "Edit selected value"
		case tuiconfig.ActionMove:
			return "Move selected item"
		case tuiconfig.ActionToggle:
			return "Expand or collapse selected item"
		case tuiconfig.ActionDelete:
			return "Delete selected item"
		case tuiconfig.ActionPublish:
			return "Publish current project draft"
		case tuiconfig.ActionPublishAll:
			return "Publish all project drafts"
		case tuiconfig.ActionDiscard:
			return "Discard current project draft"
		case tuiconfig.ActionDiscardAll:
			return "Discard all project drafts"
		case tuiconfig.ActionReload:
			return "Update current project"
		case tuiconfig.ActionReloadAll:
			return "Update all projects"
		case tuiconfig.ActionOpenDetails:
			return "Open parameter details"
		case tuiconfig.ActionFirst:
			return "First parameter or group"
		case tuiconfig.ActionLast:
			return "Last parameter or group"
		}
	}
	if block == tuiconfig.BlockConditions {
		switch action {
		case tuiconfig.ActionNew:
			return "Add new condition"
		case tuiconfig.ActionEdit:
			return "Edit condition expression"
		case tuiconfig.ActionMove:
			return "Change condition priority"
		case tuiconfig.ActionOpenDetails:
			return "Open condition details"
		case tuiconfig.ActionFirst:
			return "First condition"
		case tuiconfig.ActionLast:
			return "Last condition"
		}
	}
	if block == tuiconfig.BlockPromote {
		switch action {
		case tuiconfig.ActionClose:
			return "Close promotion"
		case tuiconfig.ActionToggle:
			return "Select or clear change"
		case tuiconfig.ActionSubmit:
			return "Open selected change diff"
		case tuiconfig.ActionSelectAll:
			return "Select visible changes"
		case tuiconfig.ActionSelectNone:
			return "Clear visible changes"
		case tuiconfig.ActionSwap:
			return "Swap source and target"
		case tuiconfig.ActionPrune:
			return "Toggle target-only removals"
		case tuiconfig.ActionSaveDraft:
			return "Save promotion draft"
		case tuiconfig.ActionSource:
			return "Toggle draft or published source"
		case tuiconfig.ActionPublish:
			return "Review and publish promotion"
		}
	}
	if block == tuiconfig.BlockGlobal && action == tuiconfig.ActionFocusPromote {
		return "Focus promote"
	}
	if block == tuiconfig.BlockHistory && action == tuiconfig.ActionSubmit {
		return "Open selected property diff"
	}
	if block == tuiconfig.BlockDetails {
		switch action {
		case tuiconfig.ActionNew:
			return "Add conditional value"
		case tuiconfig.ActionSubmit:
			return "Save details"
		case tuiconfig.ActionEditValue:
			return "Edit selected value or expression"
		case tuiconfig.ActionMove:
			return "Move selected item or priority"
		case tuiconfig.ActionCopyValue:
			return "Copy selected value or expression"
		}
	}
	if block == tuiconfig.BlockMoveInput && action == tuiconfig.ActionSubmit {
		return "Place at selected destination"
	}
	if block == tuiconfig.BlockLogs && action == tuiconfig.ActionToggleMode {
		return "Collapse or expand logs"
	}
	if block == tuiconfig.BlockAccounts {
		switch action {
		case tuiconfig.ActionCancel:
			return "Close Accounts"
		case tuiconfig.ActionSubmit:
			return "Validate or add authentication"
		case tuiconfig.ActionDelete:
			return "Delete authentication"
		case tuiconfig.ActionUp:
			return "Select previous authentication"
		case tuiconfig.ActionDown:
			return "Select next authentication"
		}
	}
	if block == tuiconfig.BlockProfiles {
		switch action {
		case tuiconfig.ActionCancel:
			return "Close Profiles"
		case tuiconfig.ActionSubmit:
			return "Switch or create profile"
		case tuiconfig.ActionRename:
			return "Rename profile"
		case tuiconfig.ActionDelete:
			return "Delete profile"
		case tuiconfig.ActionUp:
			return "Select previous profile"
		case tuiconfig.ActionDown:
			return "Select next profile"
		}
	}
	if block == tuiconfig.BlockFilter {
		switch action {
		case tuiconfig.ActionFilterFuzzy:
			return "Start fuzzy filter"
		case tuiconfig.ActionFilterStartsWith:
			return "Start prefix filter"
		case tuiconfig.ActionFilterIncludes:
			return "Start contains filter"
		case tuiconfig.ActionFilterExact:
			return "Start exact filter"
		case tuiconfig.ActionFilterExpression:
			return "Start expression filter"
		case tuiconfig.ActionFilterApply:
			return "Apply filter"
		case tuiconfig.ActionFilterCancel:
			return "Clear and close filter"
		case tuiconfig.ActionFilterUp:
			return "Close filter and move up"
		case tuiconfig.ActionFilterDown:
			return "Close filter and move down"
		}
	}
	if block == tuiconfig.BlockGlobal {
		switch action {
		case tuiconfig.ActionAccounts:
			return "Open accounts"
		case tuiconfig.ActionProfiles:
			return "Open profiles"
		case tuiconfig.ActionHelp:
			return "Open or close Actions"
		case tuiconfig.ActionForceQuit:
			return "Force quit"
		}
	}
	if block == tuiconfig.BlockHelp {
		switch action {
		case tuiconfig.ActionCancel:
			return "Close help"
		case tuiconfig.ActionSubmit:
			return "Run selected action"
		case tuiconfig.ActionUp:
			return "Move selection up"
		case tuiconfig.ActionDown:
			return "Move selection down"
		case tuiconfig.ActionPageUp:
			return "Previous page"
		case tuiconfig.ActionPageDown:
			return "Next page"
		case tuiconfig.ActionHome:
			return "First action"
		case tuiconfig.ActionEnd:
			return "Last action"
		}
	}
	if block == tuiconfig.BlockDiffView {
		switch action {
		case tuiconfig.ActionClose:
			return "Close diff"
		case tuiconfig.ActionToggle:
			return "Collapse or expand property"
		case tuiconfig.ActionLeft:
			return "Collapse property"
		case tuiconfig.ActionRight:
			return "Expand property"
		}
	}
	if block == tuiconfig.BlockHistory {
		switch action {
		case tuiconfig.ActionHistoryBothOlder:
			return "Move both versions older"
		case tuiconfig.ActionHistoryBothNewer:
			return "Move both versions newer"
		case tuiconfig.ActionHistoryChoose:
			return "Choose versions"
		case tuiconfig.ActionHistoryChanges:
			return "Toggle changed items only"
		}
	}
	if target := helpPaletteNavigationTarget(block); target != "" {
		switch action {
		case tuiconfig.ActionUp:
			return "Previous " + target
		case tuiconfig.ActionDown:
			return "Next " + target
		case tuiconfig.ActionPageUp:
			return "Previous page of " + helpPaletteNavigationCollection(block)
		case tuiconfig.ActionPageDown:
			return "Next page of " + helpPaletteNavigationCollection(block)
		case tuiconfig.ActionHome:
			return "First " + target
		case tuiconfig.ActionEnd:
			return "Last " + target
		}
	}
	return titleWords(string(action))
}

func helpPaletteNavigationCollection(block tuiconfig.Block) string {
	switch block {
	case tuiconfig.BlockParameters:
		return "parameters and groups"
	case tuiconfig.BlockAuthPicker, tuiconfig.BlockAccounts:
		return "authentication identities"
	default:
		return helpPaletteNavigationTarget(block) + "s"
	}
}

func helpPaletteNavigationTarget(block tuiconfig.Block) string {
	switch block {
	case tuiconfig.BlockProjects:
		return "project"
	case tuiconfig.BlockParameters:
		return "parameter or group"
	case tuiconfig.BlockConditions:
		return "condition"
	case tuiconfig.BlockPromote:
		return "change"
	case tuiconfig.BlockHistoryPicker:
		return "version"
	case tuiconfig.BlockDetailsForm:
		return "details field"
	case tuiconfig.BlockLogs:
		return "log entry"
	case tuiconfig.BlockDialog:
		return "dialog line"
	case tuiconfig.BlockBoolInput:
		return "boolean value"
	case tuiconfig.BlockDiffView:
		return "property"
	case tuiconfig.BlockMoveInput:
		return "destination"
	case tuiconfig.BlockAuthPicker:
		return "authentication identity"
	case tuiconfig.BlockAccounts:
		return "authentication identity"
	case tuiconfig.BlockProfiles:
		return "profile"
	case tuiconfig.BlockHelp:
		return "action"
	default:
		return ""
	}
}

func helpPaletteActionDescription(block tuiconfig.Block, action tuiconfig.Action, title string) string {
	switch action {
	case tuiconfig.ActionQuit:
		return "Quit fbrcm, prompting before discarding unsaved Details changes."
	case tuiconfig.ActionForceQuit:
		return "Quit immediately, even when Details has unsaved changes."
	case tuiconfig.ActionAccounts:
		return "Open authentication account management."
	case tuiconfig.ActionProfiles:
		return "Open profile management."
	case tuiconfig.ActionFocusProjects, tuiconfig.ActionFocusParameters, tuiconfig.ActionFocusConditions,
		tuiconfig.ActionFocusHistory, tuiconfig.ActionFocusPromote, tuiconfig.ActionFocusDetails, tuiconfig.ActionFocusLogs:
		return "Move keyboard focus to the " + strings.TrimPrefix(strings.ToLower(title), "focus ") + " panel."
	case tuiconfig.ActionRefresh:
		return "Update the cached project list from Firebase."
	case tuiconfig.ActionReload:
		return "Update (reload) Remote Config for the current project."
	case tuiconfig.ActionReloadAll:
		return "Update (reload) Remote Config for every selected project."
	case tuiconfig.ActionFirst, tuiconfig.ActionHome:
		return "Move the selection to the first " + helpPaletteNavigationDescriptionTarget(block) + "."
	case tuiconfig.ActionLast, tuiconfig.ActionEnd:
		return "Move the selection to the last " + helpPaletteNavigationDescriptionTarget(block) + "."
	case tuiconfig.ActionToggleMode:
		return "Collapse or expand the " + strings.ToLower(helpPaletteBlockTitle(block)) + "."
	case tuiconfig.ActionToggleMaximize:
		return "Maximize the workspace panel or restore the split layout."
	case tuiconfig.ActionFocusNext:
		return "Move keyboard focus to the next workspace panel."
	case tuiconfig.ActionHelp:
		return "Open or close this searchable Actions window."
	case tuiconfig.ActionOpenDetails:
		return "Open Details for the selected item."
	case tuiconfig.ActionMove:
		return "Move the selected item; for conditions this changes priority."
	case tuiconfig.ActionCopyValue:
		if block == tuiconfig.BlockDetails {
			return "Copy the selected parameter value, condition expression, or usage."
		}
	case tuiconfig.ActionSubmit:
		if block == tuiconfig.BlockDetails {
			return "Save Details. Enter also saves when the current Details selection has no other Enter action."
		}
		if block == tuiconfig.BlockMoveInput {
			return "Place the item at the selected destination."
		}
		if block == tuiconfig.BlockHistory {
			return "Compare the selected property across both versions."
		}
	}
	return title + " in the " + strings.ToLower(helpPaletteBlockTitle(block)) + "."
}

func helpPaletteNavigationDescriptionTarget(block tuiconfig.Block) string {
	if target := helpPaletteNavigationTarget(block); target != "" {
		return target
	}
	return "item"
}

func titleWords(value string) string {
	value = strings.ReplaceAll(value, "_", " ")
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

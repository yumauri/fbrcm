package app

import (
	"slices"
	"strings"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type helpPaletteAction struct {
	block   tuiconfig.Block
	action  tuiconfig.Action
	group   string
	title   string
	keys    []string
	enabled bool
	reason  string
}

var helpPaletteBlockOrder = []tuiconfig.Block{
	tuiconfig.BlockGlobal,
	tuiconfig.BlockAccounts,
	tuiconfig.BlockProfiles,
	tuiconfig.BlockProjects,
	tuiconfig.BlockParameters,
	tuiconfig.BlockConditions,
	tuiconfig.BlockHistory,
	tuiconfig.BlockDetails,
	tuiconfig.BlockLogs,
	tuiconfig.BlockFilter,
	tuiconfig.BlockHistoryPicker,
	tuiconfig.BlockDetailsForm,
	tuiconfig.BlockDialog,
	tuiconfig.BlockBoolInput,
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
			return strings.Compare(helpPaletteActionTitle(block, left), helpPaletteActionTitle(block, right))
		})
		for _, action := range actions {
			out = append(out, helpPaletteAction{
				block:  block,
				action: action,
				group:  helpPaletteBlockTitle(block),
				title:  helpPaletteActionTitle(block, action),
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
	if block == tuiconfig.BlockAccounts {
		switch action {
		case tuiconfig.ActionCancel:
			return "Close Accounts"
		case tuiconfig.ActionSubmit:
			return "Validate or add authentication"
		case tuiconfig.ActionDelete:
			return "Purge authentication"
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
			return "Purge profile"
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
			return "Open full help"
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
	if block == tuiconfig.BlockConditions && action == tuiconfig.ActionEdit {
		return "Edit expression"
	}
	if block == tuiconfig.BlockProjects && action == tuiconfig.ActionBindAuth {
		return "Bind authentication"
	}
	if block == tuiconfig.BlockDetails && action == tuiconfig.ActionEditValue {
		return "Edit selected value"
	}
	return titleWords(string(action))
}

func titleWords(value string) string {
	value = strings.ReplaceAll(value, "_", " ")
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

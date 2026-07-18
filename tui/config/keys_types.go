package config

// Block names one keybinding conflict scope.
type Block string

const (
	BlockGlobal        Block = "global"
	BlockHelp          Block = "help"
	BlockFilter        Block = "filter"
	BlockProjects      Block = "projects"
	BlockParameters    Block = "parameters"
	BlockConditions    Block = "conditions"
	BlockHistory       Block = "history"
	BlockHistoryPicker Block = "history_picker"
	BlockDetails       Block = "details"
	BlockDetailsForm   Block = "details_form"
	BlockLogs          Block = "logs"
	BlockDialog        Block = "dialog"
	BlockBoolInput     Block = "bool_input"
	BlockJSONInput     Block = "json_input"
	BlockNumberInput   Block = "number_input"
	BlockStringInput   Block = "string_input"
	BlockMoveInput     Block = "move_input"
	BlockRenameInput   Block = "rename_input"
)

// Action names one configurable TUI key action.
type Action string

const (
	ActionQuit      Action = "quit"
	ActionForceQuit Action = "force_quit"
	ActionHelp      Action = "help"
	ActionAccounts  Action = "accounts"

	ActionFilterFuzzy      Action = "fuzzy"
	ActionFilterStartsWith Action = "starts_with"
	ActionFilterIncludes   Action = "includes"
	ActionFilterExact      Action = "exact"
	ActionFilterApply      Action = "apply"
	ActionFilterCancel     Action = "cancel"
	ActionFilterUp         Action = "up"
	ActionFilterDown       Action = "down"

	ActionFocusProjects   Action = "focus_projects"
	ActionFocusParameters Action = "focus_parameters"
	ActionFocusConditions Action = "focus_conditions"
	ActionFocusHistory    Action = "focus_history"
	ActionFocusDetails    Action = "focus_details"
	ActionFocusLogs       Action = "focus_logs"
	ActionFocusNext       Action = "focus_next"

	ActionToggleMode Action = "toggle_mode"
	ActionRefresh    Action = "refresh"
	ActionSelect     Action = "select"
	ActionOpen       Action = "open"
	ActionMark       Action = "mark"

	ActionToggleMaximize   Action = "toggle_maximize"
	ActionRename           Action = "rename"
	ActionEdit             Action = "edit"
	ActionColor            Action = "color"
	ActionNew              Action = "new"
	ActionDuplicate        Action = "duplicate"
	ActionMove             Action = "move"
	ActionToggle           Action = "toggle"
	ActionDelete           Action = "delete"
	ActionPublish          Action = "publish"
	ActionPublishAll       Action = "publish_all"
	ActionDiscard          Action = "discard"
	ActionDiscardAll       Action = "discard_all"
	ActionReload           Action = "reload"
	ActionReloadAll        Action = "reload_all"
	ActionCopyName         Action = "copy_name"
	ActionCopyPath         Action = "copy_path"
	ActionOpenDetails      Action = "open_details"
	ActionExpand           Action = "expand"
	ActionCollapse         Action = "collapse"
	ActionExpandAll        Action = "expand_all"
	ActionCollapseAll      Action = "collapse_all"
	ActionExpandGroups     Action = "expand_groups"
	ActionCollapseGroups   Action = "collapse_groups"
	ActionNextGroup        Action = "next_group"
	ActionPrevGroup        Action = "prev_group"
	ActionFirst            Action = "first"
	ActionLast             Action = "last"
	ActionHistoryBothOlder Action = "pair_older"
	ActionHistoryBothNewer Action = "pair_newer"
	ActionHistoryChoose    Action = "choose_versions"
	ActionHistoryChanges   Action = "toggle_changes"
	ActionHistoryRollback  Action = "rollback"
	ActionReset            Action = "reset"

	ActionClose     Action = "close"
	ActionSubmit    Action = "submit"
	ActionSave      Action = "save"
	ActionCancel    Action = "cancel"
	ActionEditValue Action = "edit_value"
	ActionCopyValue Action = "copy_value"

	ActionLevelDown    Action = "level_down"
	ActionLevelUp      Action = "level_up"
	ActionResizeGrow   Action = "resize_grow"
	ActionResizeShrink Action = "resize_shrink"
	ActionBlankLine    Action = "blank_line"

	ActionUp       Action = "up"
	ActionDown     Action = "down"
	ActionPageUp   Action = "page_up"
	ActionPageDown Action = "page_down"
	ActionHome     Action = "home"
	ActionEnd      Action = "end"
	ActionLeft     Action = "left"
	ActionRight    Action = "right"

	ActionPrev           Action = "prev"
	ActionNext           Action = "next"
	ActionFormat         Action = "format"
	ActionToggleExpanded Action = "toggle_expanded"
)

type KeyMap map[Block]map[Action][]string

type actionRef struct {
	block  Block
	action Action
}

type State struct {
	keys     KeyMap
	disabled map[actionRef]struct{}
}

var (
	active          = validate(DefaultKeyMap())
	powerlineGlyphs = true
)

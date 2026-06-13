package config

import (
	"errors"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	corefilter "github.com/yumauri/fbrcm/core/filter"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// Block names one keybinding conflict scope.
type Block string

const (
	BlockGlobal      Block = "global"
	BlockFilter      Block = "filter"
	BlockProjects    Block = "projects"
	BlockParameters  Block = "parameters"
	BlockDetails     Block = "details"
	BlockDetailsForm Block = "details_form"
	BlockLogs        Block = "logs"
	BlockDialog      Block = "dialog"
	BlockBoolInput   Block = "bool_input"
	BlockJSONInput   Block = "json_input"
	BlockNumberInput Block = "number_input"
	BlockStringInput Block = "string_input"
	BlockMoveInput   Block = "move_input"
	BlockRenameInput Block = "rename_input"
)

// Action names one configurable TUI key action.
type Action string

const (
	ActionQuit      Action = "quit"
	ActionForceQuit Action = "force_quit"

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
	ActionFocusDetails    Action = "focus_details"
	ActionFocusLogs       Action = "focus_logs"
	ActionFocusNext       Action = "focus_next"

	ActionToggleMode Action = "toggle_mode"
	ActionRefresh    Action = "refresh"
	ActionSelect     Action = "select"
	ActionOpen       Action = "open"
	ActionMark       Action = "mark"

	ActionToggleMaximize Action = "toggle_maximize"
	ActionRename         Action = "rename"
	ActionEdit           Action = "edit"
	ActionNew            Action = "new"
	ActionDuplicate      Action = "duplicate"
	ActionMove           Action = "move"
	ActionToggle         Action = "toggle"
	ActionDelete         Action = "delete"
	ActionPublish        Action = "publish"
	ActionPublishAll     Action = "publish_all"
	ActionDiscard        Action = "discard"
	ActionDiscardAll     Action = "discard_all"
	ActionReload         Action = "reload"
	ActionReloadAll      Action = "reload_all"
	ActionCopyName       Action = "copy_name"
	ActionCopyPath       Action = "copy_path"
	ActionOpenDetails    Action = "open_details"
	ActionExpand         Action = "expand"
	ActionCollapse       Action = "collapse"
	ActionExpandAll      Action = "expand_all"
	ActionCollapseAll    Action = "collapse_all"
	ActionExpandGroups   Action = "expand_groups"
	ActionCollapseGroups Action = "collapse_groups"
	ActionNextGroup      Action = "next_group"
	ActionPrevGroup      Action = "prev_group"
	ActionFirst          Action = "first"
	ActionLast           Action = "last"

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

var active = validate(DefaultKeyMap())

// DefaultKeyMap returns complete default TUI keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		BlockGlobal: {
			ActionQuit:            {"q"},
			ActionForceQuit:       {"ctrl+c"},
			ActionFocusProjects:   {"1"},
			ActionFocusParameters: {"2"},
			ActionFocusDetails:    {"3"},
			ActionFocusLogs:       {"0"},
			ActionFocusNext:       {"tab"},
		},
		BlockFilter: {
			ActionFilterFuzzy:      {"~"},
			ActionFilterStartsWith: {"^"},
			ActionFilterIncludes:   {"/"},
			ActionFilterExact:      {"="},
			ActionFilterApply:      {"enter"},
			ActionFilterCancel:     {"esc"},
			ActionFilterUp:         {"up"},
			ActionFilterDown:       {"down"},
		},
		BlockProjects: {
			ActionToggleMode: {"c"},
			ActionRefresh:    {"r"},
			ActionSelect:     {"enter"},
			ActionOpen:       {"o"},
			ActionMark:       {"space"},
			ActionUp:         {"up", "k"},
			ActionDown:       {"down", "j"},
			ActionPageUp:     {"pgup", "h"},
			ActionPageDown:   {"pgdown", "l"},
			ActionHome:       {"home"},
			ActionEnd:        {"end"},
		},
		BlockParameters: {
			ActionToggleMaximize: {"z"},
			ActionRename:         {"r"},
			ActionEdit:           {"e"},
			ActionNew:            {"a"},
			ActionDuplicate:      {"c"},
			ActionMove:           {"m"},
			ActionToggle:         {" ", "space"},
			ActionDelete:         {"x"},
			ActionPublish:        {"p"},
			ActionPublishAll:     {"P"},
			ActionDiscard:        {"d"},
			ActionDiscardAll:     {"D"},
			ActionReload:         {"u"},
			ActionReloadAll:      {"U"},
			ActionCopyName:       {"y"},
			ActionCopyPath:       {"Y"},
			ActionOpenDetails:    {"enter"},
			ActionExpand:         {"right", "l"},
			ActionCollapse:       {"left", "h"},
			ActionExpandAll:      {">"},
			ActionCollapseAll:    {"<"},
			ActionExpandGroups:   {")"},
			ActionCollapseGroups: {"("},
			ActionNextGroup:      {"pgdown"},
			ActionPrevGroup:      {"pgup"},
			ActionFirst:          {"home"},
			ActionLast:           {"end"},
			ActionUp:             {"up", "k"},
			ActionDown:           {"down", "j"},
		},
		BlockDetails: {
			ActionClose:     {"esc"},
			ActionSubmit:    {"ctrl+enter"},
			ActionEditValue: {"right", "e"},
			ActionMove:      {"m"},
			ActionRename:    {"r"},
			ActionCopyName:  {"y"},
			ActionCopyPath:  {"Y"},
			ActionCopyValue: {"ctrl+y"},
			ActionDelete:    {"x"},
		},
		BlockDetailsForm: {
			ActionClose:    {"esc"},
			ActionUp:       {"up", "k"},
			ActionDown:     {"down", "j"},
			ActionPageUp:   {"pgup", "h"},
			ActionPageDown: {"pgdown", "l"},
			ActionHome:     {"home"},
			ActionEnd:      {"end"},
			ActionRight:    {"right"},
			ActionSubmit:   {"enter"},
		},
		BlockLogs: {
			ActionToggleMode:   {"c"},
			ActionLevelDown:    {"["},
			ActionLevelUp:      {"]"},
			ActionResizeGrow:   {"=", "+"},
			ActionResizeShrink: {"-", "_"},
			ActionBlankLine:    {"enter"},
			ActionUp:           {"up", "k"},
			ActionDown:         {"down", "j"},
			ActionPageUp:       {"pgup", "h"},
			ActionPageDown:     {"pgdown", "l"},
			ActionHome:         {"home"},
			ActionEnd:          {"end"},
		},
		BlockDialog: {
			ActionCancel:   {"esc"},
			ActionPrev:     {"left", "h", "shift+tab"},
			ActionNext:     {"right", "l", "tab"},
			ActionUp:       {"up", "k"},
			ActionDown:     {"down", "j"},
			ActionPageUp:   {"pgup"},
			ActionPageDown: {"pgdown"},
			ActionHome:     {"home"},
			ActionEnd:      {"end"},
			ActionSubmit:   {"enter"},
		},
		BlockBoolInput: {
			ActionCancel:    {"esc"},
			ActionCopyValue: {"ctrl+y"},
			ActionSubmit:    {"enter"},
			ActionUp:        {"up", "k"},
			ActionDown:      {"down", "j"},
		},
		BlockJSONInput: {
			ActionCancel:    {"esc"},
			ActionCopyValue: {"ctrl+y"},
			ActionFormat:    {"ctrl+f"},
			ActionSave:      {"ctrl+s", "ctrl+enter"},
		},
		BlockNumberInput: {
			ActionCancel:    {"esc"},
			ActionCopyValue: {"ctrl+y"},
			ActionSubmit:    {"enter"},
		},
		BlockStringInput: {
			ActionCancel:         {"esc"},
			ActionCopyValue:      {"ctrl+y"},
			ActionToggleExpanded: {"ctrl+e"},
			ActionSave:           {"ctrl+s", "ctrl+enter"},
			ActionSubmit:         {"enter"},
		},
		BlockMoveInput: {
			ActionCancel: {"esc"},
			ActionSubmit: {"enter"},
			ActionUp:     {"up", "k"},
			ActionDown:   {"down", "j"},
		},
		BlockRenameInput: {
			ActionCancel: {"esc"},
			ActionSubmit: {"enter"},
		},
	}
}

// Load reads global config, merges missing keys, writes complete map if needed.
func Load() (State, error) {
	cfg, err := coreconfig.LoadAppConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return State{}, err
		}
		cfg = &coreconfig.AppConfig{}
	}
	merged := merge(DefaultKeyMap(), cfg.Keys)
	nextConfig := toConfigMap(merged)
	if !reflect.DeepEqual(cfg.Keys, nextConfig) {
		cfg.Keys = nextConfig
		if err := coreconfig.SaveAppConfig(cfg); err != nil {
			return State{}, err
		}
	}
	active = validate(merged)
	logConflicts(active)
	return Current(), nil
}

// Current returns active keymap state.
func Current() State {
	return State{
		keys:     Clone(active.keys),
		disabled: cloneDisabled(active.disabled),
	}
}

// Matches reports whether key triggers action in active keymap.
func Matches(block Block, action Action, k string) bool {
	return active.Matches(block, action, k)
}

// Matches reports whether key triggers action.
func (s State) Matches(block Block, action Action, k string) bool {
	if s.Disabled(block, action) {
		return false
	}
	return slices.Contains(s.keys[block][action], k)
}

// Disabled reports whether action is disabled due to conflict.
func Disabled(block Block, action Action) bool {
	return active.Disabled(block, action)
}

// Disabled reports whether action is disabled due to conflict.
func (s State) Disabled(block Block, action Action) bool {
	_, ok := s.disabled[actionRef{block: block, action: action}]
	return ok
}

// Keys returns keys for action.
func Keys(block Block, action Action) []string {
	if Disabled(block, action) {
		return nil
	}
	return append([]string(nil), active.keys[block][action]...)
}

// Binding returns help binding for action.
func Binding(block Block, action Action, desc string) key.Binding {
	keys := Keys(block, action)
	binding := key.NewBinding(key.WithKeys(keys...), key.WithHelp(Label(block, action), desc))
	binding.SetEnabled(len(keys) > 0)
	return binding
}

// Label returns compact help label for action.
func Label(block Block, action Action) string {
	return strings.Join(Keys(block, action), "/")
}

// FilterModeForKey returns filter mode configured for key.
func FilterModeForKey(k string) (corefilter.Mode, bool) {
	switch {
	case Matches(BlockFilter, ActionFilterFuzzy, k):
		return corefilter.ModeFuzzy, true
	case Matches(BlockFilter, ActionFilterStartsWith, k):
		return corefilter.ModeStartsWith, true
	case Matches(BlockFilter, ActionFilterIncludes, k):
		return corefilter.ModeIncludes, true
	case Matches(BlockFilter, ActionFilterExact, k):
		return corefilter.ModeExact, true
	default:
		return "", false
	}
}

// Clone returns deep copy of keymap.
func Clone(m KeyMap) KeyMap {
	out := make(KeyMap, len(m))
	for block, actions := range m {
		out[block] = make(map[Action][]string, len(actions))
		for action, keys := range actions {
			out[block][action] = append([]string(nil), keys...)
		}
	}
	return out
}

func merge(defaults KeyMap, configured map[string]map[string][]string) KeyMap {
	out := Clone(defaults)
	for blockName, actions := range configured {
		block := Block(blockName)
		defaultActions, ok := defaults[block]
		if !ok {
			continue
		}
		for actionName, keys := range actions {
			action := Action(actionName)
			if _, ok := defaultActions[action]; !ok {
				continue
			}
			clean := cleanKeys(keys)
			if len(clean) > 0 {
				out[block][action] = clean
			}
		}
	}
	return out
}

func validate(keys KeyMap) State {
	disabled := map[actionRef]struct{}{}
	for block, actions := range keys {
		byKey := map[string][]Action{}
		for action, actionKeys := range actions {
			for _, k := range actionKeys {
				byKey[k] = append(byKey[k], action)
			}
		}
		for _, conflictActions := range byKey {
			if len(conflictActions) < 2 {
				continue
			}
			for _, action := range conflictActions {
				disabled[actionRef{block: block, action: action}] = struct{}{}
			}
		}
	}
	return State{keys: Clone(keys), disabled: disabled}
}

func logConflicts(state State) {
	logger := corelog.For("tui.config")
	for _, conflict := range conflicts(state) {
		logger.Error("keybinding conflict", "block", conflict.block, "key", conflict.key, "actions", strings.Join(conflict.actions, ","))
	}
}

type conflict struct {
	block   Block
	key     string
	actions []string
}

func conflicts(state State) []conflict {
	var out []conflict
	for block, actions := range state.keys {
		byKey := map[string][]Action{}
		for action, keys := range actions {
			if !state.Disabled(block, action) {
				continue
			}
			for _, k := range keys {
				byKey[k] = append(byKey[k], action)
			}
		}
		for k, conflictActions := range byKey {
			conflictActions = slices.DeleteFunc(conflictActions, func(action Action) bool {
				count := 0
				for _, other := range actions {
					if slices.Contains(other, k) {
						count++
					}
				}
				return count < 2
			})
			if len(conflictActions) < 2 {
				continue
			}
			actionNames := make([]string, 0, len(conflictActions))
			for _, action := range conflictActions {
				actionNames = append(actionNames, string(action))
			}
			sort.Strings(actionNames)
			out = append(out, conflict{block: block, key: k, actions: actionNames})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].block != out[j].block {
			return out[i].block < out[j].block
		}
		return out[i].key < out[j].key
	})
	return out
}

func cleanKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := map[string]struct{}{}
	for _, k := range keys {
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

func toConfigMap(m KeyMap) map[string]map[string][]string {
	out := make(map[string]map[string][]string, len(m))
	for block, actions := range m {
		out[string(block)] = make(map[string][]string, len(actions))
		for action, keys := range actions {
			out[string(block)][string(action)] = append([]string(nil), keys...)
		}
	}
	return out
}

func cloneDisabled(disabled map[actionRef]struct{}) map[actionRef]struct{} {
	out := make(map[actionRef]struct{}, len(disabled))
	for ref := range disabled {
		out[ref] = struct{}{}
	}
	return out
}

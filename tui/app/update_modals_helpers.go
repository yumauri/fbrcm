package app

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func modalCancel(m Model, block tuiconfig.Block, k string, close func(Model) Model) (Model, tea.Cmd, bool) {
	if tuiconfig.Matches(block, tuiconfig.ActionCancel, k) {
		return close(m), nil, true
	}
	return m, nil, false
}

func modalCopy(m Model, block tuiconfig.Block, k, value string) (Model, tea.Cmd, bool) {
	if tuiconfig.Matches(block, tuiconfig.ActionCopyValue, k) {
		return m, copyToClipboardCmd(value), true
	}
	return m, nil, false
}

func modalSubmit(m Model, block tuiconfig.Block, k string, action tuiconfig.Action, allowed bool, submit func() tea.Cmd) (Model, tea.Cmd, bool) {
	if !tuiconfig.Matches(block, action, k) {
		return m, nil, false
	}
	if allowed {
		return m, submit(), true
	}
	return m, nil, true
}

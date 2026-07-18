package app

import (
	tea "charm.land/bubbletea/v2"

	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

func (m *Model) requestQuit() tea.Cmd {
	if !m.detailsVisible || !m.details.Dirty() {
		return tea.Quit
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Discard Unsaved Details?",
		Body: []string{
			"The open Details form has unsaved changes.",
			"",
			"Quit and discard those changes?",
		},
		Buttons: []dialogcmp.Button{
			{Label: "Keep Editing", Variant: dialogcmp.ButtonVariantAccent},
			{Label: "Quit", Variant: dialogcmp.ButtonVariantDanger, OnPress: tea.Quit},
		},
	})
	return nil
}

func (m *Model) openAccountsBlockedByDirtyDetailsDialog() {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Unsaved Details",
		Body: []string{
			"Save or discard the open Details changes before managing accounts or profiles.",
		},
		Buttons: []dialogcmp.Button{
			{Label: "Keep Editing", Variant: dialogcmp.ButtonVariantAccent},
		},
	})
}

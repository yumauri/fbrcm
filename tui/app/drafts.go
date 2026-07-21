package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

var dialogParameterNameStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteBlueBright)

type dialogBodyFunc func() ([]string, error)

func dialogProjectLine(project core.Project) string { return viewutil.ProjectLine(project) }

func (m *Model) closeDialog() {
	if !m.dialog.IsOpen() {
		return
	}
	m.dialog = m.dialog.Close()
	m.dialogQueue = nil
}

func dialogDiffLines(diffText string) []string {
	diffText = strings.Trim(diffText, "\n")
	if idx := strings.Index(diffText, "\n\nSummary:\n"); idx >= 0 {
		diffText = diffText[:idx]
	}
	if diffText == "" {
		return []string{"No changes."}
	}
	return strings.Split(diffText, "\n")
}

func dialogCanceledCmd() tea.Cmd {
	return func() tea.Msg {
		return messages.DialogCanceledMsg{}
	}
}

func detailsEditCanceledCmd(closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsEditCanceledMsg{CloseDetails: closeDetails}
	}
}

func detailsInvalidFixCmd() tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsInvalidFixMsg{}
	}
}

func detailsInvalidDiscardCmd(closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsInvalidDiscardMsg{CloseDetails: closeDetails}
	}
}

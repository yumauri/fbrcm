package mouseutil

import (
	"strings"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// SelectFilePickerRow moves a Bubble Tea file picker's existing selection to
// a visible rendered row. The upstream component does not handle mouse input.
func SelectFilePickerRow(picker filepicker.Model, row int) (filepicker.Model, bool) {
	lines := strings.Split(strings.TrimRight(picker.View(), "\n"), "\n")
	if row < 0 || row >= len(lines) {
		return picker, false
	}
	if strings.TrimSpace(ansi.Strip(lines[row])) == "" {
		return picker, false
	}
	selected := -1
	for index, line := range lines {
		if strings.HasPrefix(ansi.Strip(line), picker.Cursor) {
			selected = index
			break
		}
	}
	if selected < 0 {
		return picker, false
	}
	code := tea.KeyDown
	if row < selected {
		code = tea.KeyUp
	}
	for selected != row {
		picker, _ = picker.Update(tea.KeyPressMsg(tea.Key{Code: code}))
		if row < selected {
			selected--
		} else {
			selected++
		}
	}
	return picker, true
}

package diff

import (
	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func colorAdded(value string) string {
	if corestyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(corestyles.ColorAdded).Render(value)
}

func colorRemoved(value string) string {
	if corestyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(corestyles.ColorRemoved).Render(value)
}

func colorChanged(value string) string {
	if corestyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(corestyles.ColorChanged).Render(value)
}

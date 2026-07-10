package importpkg

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/selection"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

func styleImportStrategySelectedChoice(choice *selection.Choice[mergeChoice]) string {
	if clistyles.NoColorEnabled() {
		return selection.DefaultSelectedChoiceStyle(choice)
	}
	if choice.Value.value == string(importStrategyOverride) {
		return lipgloss.NewStyle().Foreground(clistyles.PaletteError).Bold(true).Render(choice.String)
	}
	return selection.DefaultSelectedChoiceStyle(choice)
}

func styleImportStrategyUnselectedChoice(choice *selection.Choice[mergeChoice]) string {
	if clistyles.NoColorEnabled() || choice.Value.value != string(importStrategyOverride) {
		return choice.String
	}
	return lipgloss.NewStyle().Foreground(clistyles.PaletteError).Render(choice.String)
}

func styleImportStrategyFinalChoice(choice *selection.Choice[mergeChoice]) string {
	base := selection.DefaultFinalChoiceStyle(choice)
	if clistyles.NoColorEnabled() || choice.Value.value != string(importStrategyOverride) {
		return base
	}
	return lipgloss.NewStyle().Foreground(clistyles.PaletteError).Render(base)
}

func styleImportStrategySelectedMarker(choice *selection.Choice[mergeChoice]) string {
	if clistyles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Render("▸ ")
	}
	if choice.Value.value == string(importStrategyOverride) {
		return lipgloss.NewStyle().Foreground(clistyles.PaletteError).Bold(true).Render("▸ ")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true).Render("▸ ")
}

func styleConflictSelectedChoice(choice *selection.Choice[mergeChoice]) string {
	return renderConflictChoiceLabel(choice.Value.value, choice.String, true)
}

func styleConflictUnselectedChoice(choice *selection.Choice[mergeChoice]) string {
	return renderConflictChoiceLabel(choice.Value.value, choice.String, false)
}

func styleConflictFinalChoice(choice *selection.Choice[mergeChoice]) string {
	return renderConflictChoiceLabel(choice.Value.value, choice.String, false)
}

func renderConflictChoiceLabel(choiceValue, label string, selected bool) string {
	if clistyles.NoColorEnabled() {
		if selected {
			return lipgloss.NewStyle().Bold(true).Render(label)
		}
		return label
	}

	start := strings.LastIndex(label, " (")
	end := strings.LastIndex(label, ")")
	if start < 0 || end <= start {
		if selected {
			return lipgloss.NewStyle().Bold(true).Render(label)
		}
		return label
	}

	prefix := label[:start+2]
	value := label[start+2 : end]
	suffix := label[end:]

	valueStyle := lipgloss.NewStyle().Foreground(clistyles.ColorAdded)
	if choiceValue == string(conflictResolutionCurrent) {
		valueStyle = valueStyle.Foreground(clistyles.PaletteError)
	}
	if selected {
		valueStyle = valueStyle.Bold(true)
	}

	textStyle := lipgloss.NewStyle()
	if selected {
		textStyle = textStyle.Bold(true)
	}

	return textStyle.Render(prefix) + valueStyle.Render(value) + textStyle.Render(suffix)
}

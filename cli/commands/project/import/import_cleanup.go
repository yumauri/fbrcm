package importpkg

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

func configHasContent(cfg *firebase.RemoteConfig) bool { return importer.ConfigHasContent(cfg) }
func pruneUnusedConditions(cfg *firebase.RemoteConfig) { importer.PruneUnusedConditions(cfg) }
func dropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	importer.DropUnknownConditionReferences(cfg)
}
func normalizeEmptyParameterMaps(cfg *firebase.RemoteConfig) {
	importer.NormalizeEmptyParameterMaps(cfg)
}

type groupSummary struct {
	Name       string
	Parameters int
}

func renderGroupsTable(groups []groupSummary) string {
	rows := make([][]string, 0, len(groups))
	groupWidth := lipgloss.Width("Group")
	parametersWidth := lipgloss.Width("Parameters")
	for _, group := range groups {
		count := fmt.Sprintf("%d", group.Parameters)
		rows = append(rows, []string{group.Name, count})
		groupWidth = max(groupWidth, lipgloss.Width(group.Name))
		parametersWidth = max(parametersWidth, lipgloss.Width(count))
	}
	styleFunc := func(row, col int) lipgloss.Style {
		_ = col
		style := lipgloss.NewStyle().Padding(0, 1)
		if clistyles.NoColorEnabled() {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateBright)
	}
	tbl := table.New().Headers("Group", "Parameters").Rows(rows...).Width(groupWidth + parametersWidth + 7).
		Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

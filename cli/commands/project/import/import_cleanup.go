package importpkg

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

func configHasContent(cfg *firebase.RemoteConfig) bool {
	return cfg != nil && (len(cfg.Conditions) > 0 || len(cfg.Parameters) > 0 || len(cfg.ParameterGroups) > 0)
}

func pruneUnusedConditions(cfg *firebase.RemoteConfig) {
	if cfg == nil || len(cfg.Conditions) == 0 {
		return
	}

	used := make(map[string]struct{})
	collectUsedConditions(used, cfg.Parameters)
	for _, group := range cfg.ParameterGroups {
		collectUsedConditions(used, group.Parameters)
	}

	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if _, ok := used[condition.Name]; ok {
			kept = append(kept, condition)
		}
	}
	cfg.Conditions = kept
}

func collectUsedConditions(used map[string]struct{}, params map[string]firebase.RemoteConfigParam) {
	for _, param := range params {
		for condition := range param.ConditionalValues {
			used[condition] = struct{}{}
		}
	}
}

func dropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	rcmutate.DropUnknownConditionReferences(cfg)
}

type groupSummary struct {
	Name       string
	Parameters int
}

func summarizeGroups(groups map[string]firebase.RemoteConfigGroup) []groupSummary {
	names := strfold.SortedKeys(groups)
	out := make([]groupSummary, 0, len(names))
	for _, name := range names {
		out = append(out, groupSummary{
			Name:       name,
			Parameters: len(groups[name].Parameters),
		})
	}
	return out
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

	tbl := table.New().
		Headers("Group", "Parameters").
		Rows(rows...).
		Width(groupWidth + parametersWidth + 7).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func removeEmptyGroups(cfg *firebase.RemoteConfig) {
	rcmutate.RemoveEmptyGroups(cfg)
}

package versions

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	clidiffview "github.com/yumauri/fbrcm/cli/shared/diffview"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/dictdiff"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdiffinput "github.com/yumauri/fbrcm/core/rc/diffinput"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func renderVersionSideBySide(
	result rcdiff.Result,
	width int,
) (string, error) {
	inputs := make([]dictdiff.Input, 0, len(result.Conditions)+len(result.GroupDescriptions)+len(result.Parameters))

	for _, change := range result.Conditions {
		if change.Kind == rcdiff.ChangeUnchanged {
			continue
		}
		inputs = append(inputs, dictdiff.Input{
			EntityName: "Condition: " + change.Name,
			Left: dictdiff.NamedDictionary{
				Properties: rcdiffinput.Condition(change.PreviousPosition, change.Current),
			},
			Right: dictdiff.NamedDictionary{
				Properties: rcdiffinput.Condition(change.FinalPosition, change.Final),
			},
		})
	}
	for _, change := range result.GroupDescriptions {
		if change.Kind == rcdiff.ChangeUnchanged {
			continue
		}
		inputs = append(inputs, dictdiff.Input{
			EntityName: "Group: " + change.Group,
			Left: dictdiff.NamedDictionary{
				Properties: rcdiffinput.Group(change.Current, change.Kind != rcdiff.ChangeAdded),
			},
			Right: dictdiff.NamedDictionary{
				Properties: rcdiffinput.Group(change.Final, change.Kind != rcdiff.ChangeRemoved),
			},
		})
	}
	for _, change := range result.Parameters {
		if change.Kind == rcdiff.ChangeUnchanged {
			continue
		}
		leftGroup := change.Group
		leftKey := change.Key
		if change.PreviousKey != "" {
			leftGroup = change.PreviousGroup
			leftKey = change.PreviousKey
		}
		displayGroup := change.Group
		if change.Final == nil {
			displayGroup = leftGroup
		}
		leftProperties := rcdiffinput.Parameter(leftGroup, change.Current)
		rightProperties := rcdiffinput.Parameter(change.Group, change.Final)
		if leftKey != change.Key {
			leftProperties["name"] = dictdiff.Enum(leftKey)
			rightProperties["name"] = dictdiff.Enum(change.Key)
		}
		inputs = append(inputs, dictdiff.Input{
			EntityName: rcdiffinput.ParameterEntityName(displayGroup, change.Key),
			Left: dictdiff.NamedDictionary{
				Properties: leftProperties,
			},
			Right: dictdiff.NamedDictionary{
				Properties: rightProperties,
			},
		})
	}

	results := make([]dictdiff.Result, 0, len(inputs))
	for _, input := range inputs {
		compared, err := dictdiff.Compare(input)
		if err != nil {
			return "", fmt.Errorf("prepare %s diff: %w", strings.ToLower(entityType(input.EntityName)), err)
		}
		if len(compared.Properties) > 0 {
			results = append(results, compared)
		}
	}
	return clidiffview.Render(results, width), nil
}

func entityType(name string) string {
	label, _, found := strings.Cut(name, ":")
	if !found {
		return "entity"
	}
	return label
}

func renderVersionDiffSummary(result rcdiff.Result) string {
	summary := result.TotalSummary()
	parts := []string{
		rcdisplay.FormatCount(summary.Added, "addition", "additions"),
		rcdisplay.FormatCount(summary.Changed, "change", "changes"),
		rcdisplay.FormatCount(summary.Removed, "removal", "removals"),
	}
	if clistyles.NoColorEnabled() {
		return "Summary: " + strings.Join(parts, " · ")
	}
	parts[0] = lipgloss.NewStyle().Foreground(clistyles.ColorAdded).Render(parts[0])
	parts[1] = lipgloss.NewStyle().Foreground(clistyles.ColorChanged).Render(parts[1])
	parts[2] = lipgloss.NewStyle().Foreground(clistyles.ColorRemoved).Render(parts[2])
	return clistyles.PanelMuted.Render("Summary: ") + strings.Join(parts, clistyles.PanelMuted.Render(" · "))
}

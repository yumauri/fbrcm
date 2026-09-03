package versions

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestRenderVersionSideBySideAdaptsEveryRemoteConfigEntity(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	from := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{
			Name:       "Audience",
			Expression: "country == 'US'",
			TagColor:   "BLUE",
		}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"WEB": {
				Description: "Web settings",
				Parameters: map[string]firebase.RemoteConfigParam{
					"banner": {
						ValueType:    "JSON",
						DefaultValue: &firebase.RemoteConfigValue{Value: `{"message":"old"}`},
					},
				},
			},
		},
	}
	to := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{
			Name:       "Audience",
			Expression: "country == 'CA'",
			TagColor:   "GREEN",
		}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"WEB": {
				Description: "Web application settings",
				Parameters: map[string]firebase.RemoteConfigParam{
					"banner": {
						ValueType:    "JSON",
						DefaultValue: &firebase.RemoteConfigValue{Value: `{"message":"new"}`},
					},
				},
			},
		},
	}

	output, err := renderVersionSideBySide(
		rcdiff.CompareRemoteConfigs(from, to),
		100,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Condition: Audience",
		"Group: WEB",
		"Property: WEB / banner",
		"expression",
		"description",
		"value · default",
		`"message": "old"`,
		`"message": "new"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("renderVersionSideBySide() = %q, want substring %q", output, want)
		}
	}
	for _, unwanted := range []string{"From:", "To:", "┌", "┐", "└", "┘"} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("renderVersionSideBySide() = %q, do not want substring %q", output, unwanted)
		}
	}
}

func TestRenderVersionSideBySideRespectsTerminalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	from := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"long": {ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: strings.Repeat("before ", 20)}},
	}}
	to := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"long": {ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: strings.Repeat("after ", 20)}},
	}}

	const width = 44
	output, err := renderVersionSideBySide(
		rcdiff.CompareRemoteConfigs(from, to),
		width,
	)
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(output, "\n") {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line width = %d, want <= %d: %q", got, width, line)
		}
	}
}

func TestRenderVersionDiffSummaryUsesGrammaticalCounts(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result := rcdiff.Result{
		Conditions: []rcdiff.ConditionChange{{Kind: rcdiff.ChangeAdded}},
		Parameters: []rcdiff.ParameterChange{
			{Kind: rcdiff.ChangeChanged},
			{Kind: rcdiff.ChangeRemoved},
			{Kind: rcdiff.ChangeRemoved},
		},
	}
	got := renderVersionDiffSummary(result)
	for _, want := range []string{"1 addition", "1 change", "2 removals"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderVersionDiffSummary() = %q, want substring %q", got, want)
		}
	}
}

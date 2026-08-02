package viewutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestRenderManagedValueSummaryUsesSharedTUIStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	rollout := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		RolloutValue: json.RawMessage(`{"value":"20","percent":10}`),
	}, "NUMBER")

	got, ok := RenderManagedValueSummary(rollout, "NUMBER")
	if !ok {
		t.Fatal("rollout was not recognized as a managed value")
	}
	for fragment, want := range map[string]string{
		"icon":       styles.PanelMuted.Render("◐ "),
		"percentage": styles.SecondaryTitleCount.Render("10%"),
		"arrow":      styles.PanelMuted.Render(" → "),
		"value":      corestyles.ValueTextStyle("20", "NUMBER").Render("20"),
		"separator":  styles.PanelMuted.Render(" | "),
		"control":    styles.PanelMuted.Render("(no change)"),
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rollout does not use the %s style: %q", fragment, got)
		}
	}
}

func TestRenderManagedExperimentSummaryUsesTypeStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	experiment := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		ExperimentValue: json.RawMessage(`{"variantValue":[{"value":""},{"noChange":true},{"value":"new value"}],"exposurePercent":15}`),
	}, "STRING")

	got, ok := RenderManagedValueSummary(experiment, "STRING")
	if !ok {
		t.Fatal("experiment was not recognized as a managed value")
	}
	want := styles.PanelMuted.Render("⚗ ") +
		styles.SecondaryTitleCount.Render("15%") +
		styles.PanelMuted.Render(" : ") +
		styles.PanelMuted.Render("(empty string)") +
		styles.PanelMuted.Render(" | ") +
		styles.PanelMuted.Render("(no change)") +
		styles.PanelMuted.Render(" | ") +
		corestyles.ValueTextStyle("new value", "STRING").Render("new value")
	if got != want {
		t.Fatalf("experiment render = %q, want %q", got, want)
	}
}

func TestRenderManagedValueSummaryMutesOpaqueLabels(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	values := []firebase.RemoteConfigValue{
		{PersonalizationValue: json.RawMessage(`{}`)},
		{ExperimentValue: json.RawMessage(`{}`)},
		{UnknownValueOption: "futureValue", UnknownValue: json.RawMessage(`{}`)},
	}
	for _, value := range values {
		summary := core.SummarizeRemoteConfigDisplayValue(value, "STRING")
		got, ok := RenderManagedValueSummary(summary, "STRING")
		if !ok || got != styles.PanelMuted.Render(summary.Text) {
			t.Fatalf("opaque summary = %q, %t; want muted %q", got, ok, summary.Text)
		}
	}
}

func TestRenderManagedValueSummaryRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	summary := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		RolloutValue: json.RawMessage(`{"value":"20","percent":10}`),
	}, "NUMBER")
	got, ok := RenderManagedValueSummary(summary, "NUMBER")
	if !ok || got != "◐ 10% → 20 | (no change)" {
		t.Fatalf("no-color rollout = %q, %t", got, ok)
	}
}

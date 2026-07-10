package diff

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func TestRenderRemoteConfigDiffNoChanges(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "on"},
				ValueType:    "STRING",
			},
		},
	}

	diff, changed := RenderRemoteConfigDiff(cfg, cfg)
	if changed {
		t.Fatalf("changed = true, want false")
	}
	if diff != "" {
		t.Fatalf("diff = %q, want empty", diff)
	}
}

func TestRenderRemoteConfigDiffAddedParameter(t *testing.T) {
	current := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}
	final := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "true"},
				ValueType:    "BOOLEAN",
			},
		},
	}

	diff, changed := RenderRemoteConfigDiff(current, final)
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	for _, want := range []string{"Parameters:", "flag", "Summary:", "1 parameter added"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q:\n%s", want, diff)
		}
	}
}

func TestRenderRemoteConfigDiffChangedConditionAndGroupedParameter(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	current := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "beta", Expression: "app.version == '1'", TagColor: "BLUE"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "old"},
				ValueType:    "STRING",
			},
		},
	}
	final := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "beta", Expression: "app.version == '2'", TagColor: "GREEN"},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": {
						DefaultValue: &firebase.RemoteConfigValue{Value: "new"},
						Description:  "New description",
						ValueType:    "BOOLEAN",
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"beta": {Value: "conditional"},
						},
					},
				},
			},
		},
	}

	diff, changed := RenderRemoteConfigDiff(current, final)
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	for _, want := range []string{
		"Conditions:",
		"app.version == '1'",
		"app.version == '2'",
		"Parameters:",
		"flag [group-a]",
		"group:",
		"(root)",
		"[group-a]",
		"type:",
		"STRING",
		"BOOLEAN",
		"default:",
		"old",
		"new",
		"cond beta:",
		"Summary:",
		"0 condition added, 0 removed, 1 changed, 0 unchanged",
		"0 parameter added, 0 removed, 1 changed, 0 unchanged",
	} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q:\n%s", want, diff)
		}
	}
}

func TestRenderRemoteConfigDiffRemovedParameter(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	current := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "on"},
				ValueType:    "STRING",
			},
			"keep": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "yes"},
				ValueType:    "STRING",
			},
		},
	}
	final := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"keep": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "yes"},
				ValueType:    "STRING",
			},
		},
	}

	diff, changed := RenderRemoteConfigDiff(current, final)
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	for _, want := range []string{
		"Parameters:",
		"flag",
		"Summary:",
		"0 parameter added, 1 removed, 0 changed, 1 unchanged",
	} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q:\n%s", want, diff)
		}
	}
}

func TestConflictPreviewAndChoiceValues(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	conditionPreview := RenderConflictPreview(
		"condition beta",
		firebase.RemoteConfigCondition{Expression: "old"},
		firebase.RemoteConfigCondition{Expression: "new"},
	)
	if !strings.Contains(conditionPreview, "beta") || !strings.Contains(conditionPreview, "old") || !strings.Contains(conditionPreview, "new") {
		t.Fatalf("condition conflict preview = %q", conditionPreview)
	}

	slot := ParamSlotPreview{
		Group: "group-a",
		Param: firebase.RemoteConfigParam{
			DefaultValue: &firebase.RemoteConfigValue{Value: "on"},
			ValueType:    "BOOLEAN",
		},
	}
	choice := RenderConflictChoiceValue(slot)
	for _, want := range []string{"group=[group-a]", "default=on", "type=BOOLEAN"} {
		if !strings.Contains(choice, want) {
			t.Fatalf("slot choice = %q, want substring %q", choice, want)
		}
	}

	if got := RenderConflictChoiceValue("  hello world  "); got != `"hello world"` {
		t.Fatalf("string choice = %q, want quoted trimmed value", got)
	}
}

func TestDiffFormattingHelpers(t *testing.T) {
	if !rcdisplay.IsSimpleToken("enabled") {
		t.Fatalf("IsSimpleToken(enabled) = false, want true")
	}
	for _, value := range []string{"hello world", "tab\tvalue", "line\nvalue", `say "hi"`} {
		if rcdisplay.IsSimpleToken(value) {
			t.Fatalf("IsSimpleToken(%q) = true, want false", value)
		}
	}
	if got := formatGroupValue(""); got != "(root)" {
		t.Fatalf("formatGroupValue(root) = %q, want (root)", got)
	}
	if got := formatGroupValue("group-a"); got != "[group-a]" {
		t.Fatalf("formatGroupValue(group-a) = %q, want [group-a]", got)
	}
	if got := emptyAsDash(" "); got != "(empty)" {
		t.Fatalf("emptyAsDash(blank) = %q, want (empty)", got)
	}
	if got := emptyAsDash("value"); got != "value" {
		t.Fatalf("emptyAsDash(value) = %q, want value", got)
	}
}

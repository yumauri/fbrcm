package filter

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

// exprTestConfig builds a representative RemoteConfig used to lock in the
// behavior of the expression engine before it is refactored/split.
func exprTestConfig() *firebase.RemoteConfig {
	return &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "ios", Expression: "device.os == 'ios'"},
			{Name: "android", Expression: "device.os == 'android'"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"feature_login": {
				ValueType:    "STRING",
				DefaultValue: &firebase.RemoteConfigValue{Value: "on"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"ios": {Value: "off"},
				},
				Description: "Login feature toggle",
			},
			"max_items": {
				ValueType:    "NUMBER",
				DefaultValue: &firebase.RemoteConfigValue{Value: "10"},
			},
			"enabled": {
				ValueType:    "BOOLEAN",
				DefaultValue: &firebase.RemoteConfigValue{Value: "true"},
			},
			"config_json": {
				ValueType:    "JSON",
				DefaultValue: &firebase.RemoteConfigValue{Value: `{"a":1,"b":"two"}`},
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"checkout": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"checkout_flow": {
						ValueType:    "STRING",
						DefaultValue: &firebase.RemoteConfigValue{Value: "v2"},
					},
				},
			},
		},
	}
}

func TestCompileExpressionEmpty(t *testing.T) {
	expr, err := CompileExpression("   ")
	if err != nil {
		t.Fatalf("CompileExpression empty: unexpected error %v", err)
	}
	if expr != nil {
		t.Fatalf("CompileExpression empty: expected nil expression, got %v", expr)
	}
	// A nil expression must always match.
	matched, err := expr.MatchProject("p", "P", nil)
	if err != nil {
		t.Fatalf("nil MatchProject: unexpected error %v", err)
	}
	if !matched {
		t.Fatalf("nil MatchProject: expected match")
	}
}

func TestCompileExpressionInvalid(t *testing.T) {
	if _, err := CompileExpression("project_id == "); err == nil {
		t.Fatalf("expected compile error for invalid expression")
	}
}

func TestMatchProject(t *testing.T) {
	cfg := exprTestConfig()
	cases := []struct {
		name string
		expr string
		want bool
	}{
		{"project id equal", `project_id == "demo-prod"`, true},
		{"project id not equal", `project_id == "other"`, false},
		{"project name equal", `project == "Demo Prod"`, true},
		{"condition membership", `"ios" in conditions`, true},
		{"condition absent", `"web" in conditions`, false},
		{"group membership", `"checkout" in groups`, true},
		{"parameter lookup default", `parameters["feature_login"].default == "on"`, true},
		{"parameter group field", `parameters["checkout_flow"].group == "checkout"`, true},
		{"boolean and", `project_id == "demo-prod" && "ios" in conditions`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := CompileExpression(tc.expr)
			if err != nil {
				t.Fatalf("compile %q: %v", tc.expr, err)
			}
			got, err := expr.MatchProject("demo-prod", "Demo Prod", cfg)
			if err != nil {
				t.Fatalf("match %q: %v", tc.expr, err)
			}
			if got != tc.want {
				t.Fatalf("expr %q = %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}

func TestMatchParameter(t *testing.T) {
	cfg := exprTestConfig()
	cases := []struct {
		name  string
		expr  string
		param string
		group string
		want  bool
	}{
		{"name equal", `name == "feature_login"`, "feature_login", "", true},
		{"name starts with", `name startsWith "feature"`, "feature_login", "", true},
		{"name matches regex", `name matches "^feature_"`, "feature_login", "", true},
		{"string default", `default == "on"`, "feature_login", "", true},
		{"value matches conditional", `value == "off"`, "feature_login", "", true},
		{"is_string default", `is_string(default)`, "feature_login", "", true},
		{"root group is nil", `group == nil`, "feature_login", "", true},
		{"number default coercion", `default == 10`, "max_items", "", true},
		{"is_number default", `is_number(default)`, "max_items", "", true},
		{"number greater", `default > 5`, "max_items", "", true},
		{"is_boolean default", `is_boolean(default)`, "enabled", "", true},
		{"is_json default", `is_json(default)`, "config_json", "", true},
		{"jq selects field", `default | jq(.a == 1)`, "config_json", "", true},
		{"group param group label", `group == "checkout"`, "checkout_flow", "checkout", true},
		{"group param default", `default == "v2"`, "checkout_flow", "checkout", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := CompileExpression(tc.expr)
			if err != nil {
				t.Fatalf("compile %q: %v", tc.expr, err)
			}
			got, err := expr.MatchParameter("demo-prod", "Demo Prod", cfg, tc.param, tc.group)
			if err != nil {
				t.Fatalf("match %q: %v", tc.expr, err)
			}
			if got != tc.want {
				t.Fatalf("expr %q (param %q group %q) = %v, want %v", tc.expr, tc.param, tc.group, got, tc.want)
			}
		})
	}
}

func TestFilterMatchModes(t *testing.T) {
	cases := []struct {
		name  string
		value string
		query string
		mode  Mode
		want  bool
	}{
		{"empty query always matches", "anything", "", ModeFuzzy, true},
		{"fuzzy subsequence", "feature_login", "flgn", ModeFuzzy, true},
		{"fuzzy no match", "feature_login", "xyz", ModeFuzzy, false},
		{"starts with hit", "feature_login", "feat", ModeStartsWith, true},
		{"starts with miss", "feature_login", "login", ModeStartsWith, false},
		{"includes hit", "feature_login", "ure_lo", ModeIncludes, true},
		{"includes miss", "feature_login", "zzz", ModeIncludes, false},
		{"exact case-insensitive", "Feature_Login", "feature_login", ModeExact, true},
		{"exact miss", "feature_login", "feature", ModeExact, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Match(tc.value, tc.query, tc.mode)
			if got != tc.want {
				t.Fatalf("Match(%q,%q,%v) = %v, want %v", tc.value, tc.query, tc.mode, got, tc.want)
			}
		})
	}
}

func TestModeLabelRoundTrip(t *testing.T) {
	for _, mode := range []Mode{ModeFuzzy, ModeStartsWith, ModeIncludes, ModeExact} {
		label := mode.Label()
		got, ok := ModeFromLabel(label)
		if !ok {
			t.Fatalf("ModeFromLabel(%q) not ok", label)
		}
		if got != mode {
			t.Fatalf("round trip %v -> %q -> %v", mode, label, got)
		}
	}
}

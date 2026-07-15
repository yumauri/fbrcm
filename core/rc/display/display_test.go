package display

import (
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestFormatLocalDateTime(t *testing.T) {
	value := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.Local)
	if got := FormatLocalDateTime(value); got != "2020-01-02 03:04:05" {
		t.Fatalf("FormatLocalDateTime() = %q, want 2020-01-02 03:04:05", got)
	}
	if got := FormatLocalDateTime(time.Time{}); got != "" {
		t.Fatalf("FormatLocalDateTime(zero) = %q, want empty", got)
	}
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name      string
		value     firebase.RemoteConfigValue
		valueType string
		want      string
	}{
		{
			name:      "in app default",
			value:     firebase.RemoteConfigValue{UseInAppDefault: true},
			valueType: "STRING",
			want:      "<in-app default>",
		},
		{
			name:      "personalization",
			value:     firebase.RemoteConfigValue{PersonalizationValue: []byte(`{}`)},
			valueType: "JSON",
			want:      "<personalization>",
		},
		{
			name:      "rollout",
			value:     firebase.RemoteConfigValue{RolloutValue: []byte(`{}`)},
			valueType: "STRING",
			want:      "<rollout>",
		},
		{
			name:      "empty string type",
			value:     firebase.RemoteConfigValue{Value: ""},
			valueType: "STRING",
			want:      "(empty string)",
		},
		{
			name:      "empty default type",
			value:     firebase.RemoteConfigValue{Value: ""},
			valueType: "",
			want:      "(empty string)",
		},
		{
			name:      "empty boolean type",
			value:     firebase.RemoteConfigValue{Value: ""},
			valueType: "BOOLEAN",
			want:      "(empty boolean)",
		},
		{
			name:      "plain value",
			value:     firebase.RemoteConfigValue{Value: "enabled"},
			valueType: "STRING",
			want:      "enabled",
		},
		{
			name:      "multiline escaped",
			value:     firebase.RemoteConfigValue{Value: "line1\nline2"},
			valueType: "STRING",
			want:      `line1\nline2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSummary(tt.value, tt.valueType); got != tt.want {
				t.Fatalf("FormatSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmptyValueType(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: "string"},
		{in: "  BOOLEAN  ", want: "boolean"},
		{in: "JSON", want: "json"},
	}

	for _, tt := range tests {
		if got := EmptyValueType(tt.in); got != tt.want {
			t.Fatalf("EmptyValueType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatRawValue(t *testing.T) {
	tests := []struct {
		value     string
		valueType string
		want      string
	}{
		{value: "", valueType: "", want: "(empty string)"},
		{value: "", valueType: " BOOLEAN ", want: "(empty boolean)"},
		{value: "enabled", valueType: "STRING", want: "enabled"},
		{value: "line1\nline2", valueType: "STRING", want: `line1\nline2`},
	}

	for _, tt := range tests {
		if got := FormatRawValue(tt.value, tt.valueType); got != tt.want {
			t.Fatalf("FormatRawValue(%q, %q) = %q, want %q", tt.value, tt.valueType, got, tt.want)
		}
	}
}

func TestFormatCount(t *testing.T) {
	for _, tc := range []struct {
		count int
		want  string
	}{
		{count: 0, want: "0 parameters"},
		{count: 1, want: "1 parameter"},
		{count: 5, want: "5 parameters"},
	} {
		if got := FormatCount(tc.count, "parameter", "parameters"); got != tc.want {
			t.Fatalf("FormatCount(%d) = %q, want %q", tc.count, got, tc.want)
		}
	}
}

func TestFormatDiff(t *testing.T) {
	tests := []struct {
		name  string
		value firebase.RemoteConfigValue
		want  string
	}{
		{name: "plain", value: firebase.RemoteConfigValue{Value: "enabled"}, want: "enabled"},
		{name: "default", value: firebase.RemoteConfigValue{UseInAppDefault: true}, want: "useInAppDefault"},
		{name: "personalization", value: firebase.RemoteConfigValue{PersonalizationValue: []byte(`{"a":"\u003c"}`)}, want: `{"a":"<"}`},
		{name: "rollout", value: firebase.RemoteConfigValue{RolloutValue: []byte(`{"b":"\u0026"}`)}, want: `{"b":"&"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDiff(tt.value); got != tt.want {
				t.Fatalf("FormatDiff() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPlainValue(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: "(empty)"},
		{in: "enabled", want: "enabled"},
		{in: "hello world", want: `"hello world"`},
		{in: `"quoted"`, want: `"\"quoted\""`},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := FormatPlainValue(tt.in); got != tt.want {
				t.Fatalf("FormatPlainValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsSimpleToken(t *testing.T) {
	if !IsSimpleToken("enabled") {
		t.Fatalf("IsSimpleToken(enabled) = false, want true")
	}
	for _, value := range []string{"hello world", "a\tb", "a\nb", `"quoted"`} {
		if IsSimpleToken(value) {
			t.Fatalf("IsSimpleToken(%q) = true, want false", value)
		}
	}
}

func TestFormatProject(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		want      string
	}{
		{name: "", projectID: "proj-1", want: "proj-1"},
		{name: "My App", projectID: "proj-1", want: "My App (proj-1)"},
		{name: "  ", projectID: "proj-1", want: "proj-1"},
	}

	for _, tt := range tests {
		if got := FormatProject(tt.name, tt.projectID); got != tt.want {
			t.Fatalf("FormatProject(%q, %q) = %q, want %q", tt.name, tt.projectID, got, tt.want)
		}
	}
}

func TestFormatConditionLabel(t *testing.T) {
	if got := FormatConditionLabel("default"); got != "Default value" {
		t.Fatalf("FormatConditionLabel(default) = %q, want Default value", got)
	}
	if got := FormatConditionLabel("ios"); got != "ios" {
		t.Fatalf("FormatConditionLabel(ios) = %q, want ios", got)
	}
}

func TestFormatParameterHeader(t *testing.T) {
	tests := []struct {
		key   string
		group string
		want  string
	}{
		{key: "flag", group: "", want: "flag"},
		{key: "flag", group: "group-a", want: "flag [group-a]"},
	}

	for _, tt := range tests {
		if got := FormatParameterHeader(tt.key, tt.group); got != tt.want {
			t.Fatalf("FormatParameterHeader(%q, %q) = %q, want %q", tt.key, tt.group, got, tt.want)
		}
	}
}

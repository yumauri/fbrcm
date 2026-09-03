package shared

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestParameterSearchMatch(t *testing.T) {
	search := NewParameterSearch("  Login  Feature ")
	if search.Empty() {
		t.Fatal("NewParameterSearch should not be empty")
	}

	param := firebase.RemoteConfigParam{
		Description:  "Login feature toggle",
		DefaultValue: &firebase.RemoteConfigValue{Value: "on"},
		ConditionalValues: map[string]firebase.RemoteConfigValue{
			"ios": {Value: "off"},
		},
	}
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "ios", Expression: "device.os == 'ios'"}},
	}

	if !MatchParameterSearch("feature_login", param, cfg, search) {
		t.Fatal("MatchParameterSearch should match description")
	}
	if MatchParameterSearch("other_flag", firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: "x"},
	}, nil, search) {
		t.Fatal("MatchParameterSearch should not match unrelated name without shared text")
	}
}

func TestParameterSearchEmptyMatchesAll(t *testing.T) {
	search := NewParameterSearch("   ")
	param := firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "x"}}
	if !MatchParameterSearch("any", param, nil, search) {
		t.Fatal("empty search should match all parameters")
	}
}

func TestCollapseAndNormalizeSearchText(t *testing.T) {
	if got := collapseSpaces("  a   b  "); got != "a b" {
		t.Fatalf("collapseSpaces = %q", got)
	}
	if got := normalizeSearchText("Feature_Login"); got != "feature login" {
		t.Fatalf("normalizeSearchText = %q", got)
	}
}

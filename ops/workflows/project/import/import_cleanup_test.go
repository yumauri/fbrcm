package importpkg

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestPruneUnusedConditions(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "used"},
			{Name: "unused"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"used": {Value: "on"},
				},
			},
		},
	}
	pruneUnusedConditions(cfg)
	if len(cfg.Conditions) != 1 || cfg.Conditions[0].Name != "used" {
		t.Fatalf("conditions = %+v, want used only", cfg.Conditions)
	}
}

func TestConfigHasContent(t *testing.T) {
	if configHasContent(nil) {
		t.Fatal("nil config should have no content")
	}
	if configHasContent(&firebase.RemoteConfig{}) {
		t.Fatal("empty config should have no content")
	}
	if !configHasContent(&firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {},
		},
	}) {
		t.Fatal("config with parameters should have content")
	}
}

func TestRenderGroupsTable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderGroupsTable([]groupSummary{{Name: "checkout", Parameters: 2}})
	if got == "" || !strings.Contains(got, "checkout") {
		t.Fatalf("renderGroupsTable = %q", got)
	}
}

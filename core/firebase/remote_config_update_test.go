package firebase

import (
	"strings"
	"testing"
)

func TestPrepareRemoteConfigUpdateRejectsUnsupportedConditionFields(t *testing.T) {
	_, err := PrepareRemoteConfigUpdate([]byte(`{"conditions":[{"name":"staff","expression":"true","description":"unsupported"}]}`))
	if err == nil || !strings.Contains(err.Error(), `unknown field "description"`) {
		t.Fatalf("PrepareRemoteConfigUpdate error = %v", err)
	}
}

func TestMarshalRemoteConfigForUpdateDoesNotMutateSource(t *testing.T) {
	cfg := &RemoteConfig{Conditions: []RemoteConfigCondition{{Name: "staff", Expression: "true", TagColor: "deep_orange"}}}
	if _, err := MarshalRemoteConfigForUpdate(cfg); err != nil {
		t.Fatalf("MarshalRemoteConfigForUpdate = %v", err)
	}
	if cfg.Conditions[0].TagColor != "deep_orange" {
		t.Fatalf("source config was mutated: %#v", cfg.Conditions[0])
	}
}

func TestMarshalRemoteConfigForUpdateRejectsUnsupportedConditionColor(t *testing.T) {
	_, err := MarshalRemoteConfigForUpdate(&RemoteConfig{Conditions: []RemoteConfigCondition{{Name: "staff", Expression: "true", TagColor: "RED"}}})
	if err == nil || !strings.Contains(err.Error(), `condition "staff"`) || !strings.Contains(err.Error(), `unsupported condition color "RED"`) {
		t.Fatalf("MarshalRemoteConfigForUpdate error = %v", err)
	}
}

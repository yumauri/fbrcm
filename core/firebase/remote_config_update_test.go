package firebase

import (
	"encoding/json"
	"reflect"
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

func TestPrepareRemoteConfigUpdatePreservesManagedValues(t *testing.T) {
	raw := []byte(`{
		"parameters": {
			"provider": {
				"conditionalValues": {
					"beta": {"personalizationValue": {"personalizationId": "personalization_1"}},
					"experiment": {"experimentValue": {"experimentId": "experiment_1", "exposurePercent": 12.5, "variantValue": [{"variantId": "0", "value": "false"}, {"variantId": "1", "value": "true"}]}},
					"staff": {"rolloutValue": {"rolloutId": "rollout_1", "value": "kyc2", "percent": 10}},
					"future": {"someFutureValue": {"id": "future_1", "settings": {"enabled": true}}}
				}
			}
		},
		"version": {"versionNumber": "12"}
	}`)
	update, err := PrepareRemoteConfigUpdate(raw)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(update, &got); err != nil {
		t.Fatal(err)
	}
	parameters := got["parameters"].(map[string]any)
	provider := parameters["provider"].(map[string]any)
	values := provider["conditionalValues"].(map[string]any)
	wantPersonalization := map[string]any{"personalizationId": "personalization_1"}
	if value := values["beta"].(map[string]any)["personalizationValue"]; !reflect.DeepEqual(value, wantPersonalization) {
		t.Fatalf("personalizationValue = %#v, want %#v", value, wantPersonalization)
	}
	wantExperiment := map[string]any{
		"experimentId":    "experiment_1",
		"exposurePercent": float64(12.5),
		"variantValue": []any{
			map[string]any{"variantId": "0", "value": "false"},
			map[string]any{"variantId": "1", "value": "true"},
		},
	}
	if value := values["experiment"].(map[string]any)["experimentValue"]; !reflect.DeepEqual(value, wantExperiment) {
		t.Fatalf("experimentValue = %#v, want %#v", value, wantExperiment)
	}
	wantRollout := map[string]any{"rolloutId": "rollout_1", "value": "kyc2", "percent": float64(10)}
	if value := values["staff"].(map[string]any)["rolloutValue"]; !reflect.DeepEqual(value, wantRollout) {
		t.Fatalf("rolloutValue = %#v, want %#v", value, wantRollout)
	}
	wantFuture := map[string]any{"id": "future_1", "settings": map[string]any{"enabled": true}}
	if value := values["future"].(map[string]any)["someFutureValue"]; !reflect.DeepEqual(value, wantFuture) {
		t.Fatalf("someFutureValue = %#v, want %#v", value, wantFuture)
	}
	if _, ok := got["version"]; ok {
		t.Fatalf("update still contains read-only version: %#v", got["version"])
	}
}

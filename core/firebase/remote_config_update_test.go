package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPrepareRemoteConfigUpdateWritesOnlyExplicitChangeNoteVersionMetadata(t *testing.T) {
	raw := []byte(`{"version":{"versionNumber":"12","description":"inherited"},"parameters":{}}`)
	withoutNote, err := PrepareRemoteConfigUpdate(raw)
	if err != nil {
		t.Fatal(err)
	}
	var without map[string]any
	if err := json.Unmarshal(withoutNote, &without); err != nil {
		t.Fatal(err)
	}
	if _, ok := without["version"]; ok {
		t.Fatalf("update inherited source version metadata: %#v", without["version"])
	}

	withNote, err := PrepareRemoteConfigUpdate(raw, " Enable checkout v2 ")
	if err != nil {
		t.Fatal(err)
	}
	var with map[string]any
	if err := json.Unmarshal(withNote, &with); err != nil {
		t.Fatal(err)
	}
	version, ok := with["version"].(map[string]any)
	if !ok || !reflect.DeepEqual(version, map[string]any{"description": "Enable checkout v2"}) {
		t.Fatalf("version = %#v, want description-only metadata", with["version"])
	}
}

func TestChangeNoteContextNormalizesAndPreservesExplicitEmpty(t *testing.T) {
	ctx, err := WithChangeNote(context.Background(), " release note ")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := ChangeNoteFromContext(ctx); !ok || got != "release note" {
		t.Fatalf("ChangeNoteFromContext = %q, %v", got, ok)
	}
	ctx, err = WithChangeNote(context.Background(), " ")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := ChangeNoteFromContext(ctx); !ok || got != "" {
		t.Fatalf("empty ChangeNoteFromContext = %q, %v", got, ok)
	}
	if _, err := WithChangeNote(context.Background(), "line one\nline two"); err == nil {
		t.Fatal("multiline change note returned nil error")
	} else {
		var invalid *InvalidChangeNoteError
		if !errors.As(err, &invalid) {
			t.Fatalf("multiline change note error = %T, want InvalidChangeNoteError", err)
		}
	}
}

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

func TestMarshalRemoteConfigForUpdateRejectsInvalidTemplateSemantics(t *testing.T) {
	tests := []struct {
		name string
		cfg  *RemoteConfig
		want string
	}{
		{"empty condition", &RemoteConfig{Conditions: []RemoteConfigCondition{{}}}, "empty name"},
		{"duplicate condition", &RemoteConfig{Conditions: []RemoteConfigCondition{{Name: "beta", Expression: "true"}, {Name: "beta", Expression: "false"}}}, "duplicated"},
		{"unsupported value type", &RemoteConfig{Parameters: map[string]RemoteConfigParam{"flag": {ValueType: "FLOAT"}}}, "unsupported valueType"},
		{"duplicate parameter", &RemoteConfig{Parameters: map[string]RemoteConfigParam{"flag": {}}, ParameterGroups: map[string]RemoteConfigGroup{"group": {Parameters: map[string]RemoteConfigParam{"flag": {}}}}}, "appears more than once"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := MarshalRemoteConfigForUpdate(test.cfg)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("MarshalRemoteConfigForUpdate error = %v, want %q", err, test.want)
			}
		})
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

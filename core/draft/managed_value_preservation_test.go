package draft

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestUnrelatedMutationPreservesManagedValues(t *testing.T) {
	current := json.RawMessage(`{
		"parameters": {
			"managed": {
				"defaultValue": {
					"experimentValue": {
						"experimentId": "experiment-1",
						"exposurePercent": 12.5,
						"variantValue": [
							{"variantId": "0", "noChange": true},
							{"variantId": "1", "value": ""}
						]
					}
				}
			},
			"future": {
				"defaultValue": {
					"someFutureValue": {
						"id": "future-1",
						"settings": {"enabled": true}
					}
				}
			},
			"plain": {
				"defaultValue": {"value": "before"}
			}
		},
		"version": {"versionNumber": "7"}
	}`)

	final, err := BuildMutatedRemoteConfig(current, "", func(cfg *firebase.RemoteConfig) error {
		plain := cfg.Parameters["plain"]
		plain.Description = "updated"
		cfg.Parameters["plain"] = plain
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	before := managedExperimentValue(t, current)
	after := managedExperimentValue(t, final)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("managed value changed\nbefore: %#v\nafter:  %#v", before, after)
	}
	beforeFuture := remoteConfigValueOption(t, current, "future", "someFutureValue")
	afterFuture := remoteConfigValueOption(t, final, "future", "someFutureValue")
	if !reflect.DeepEqual(afterFuture, beforeFuture) {
		t.Fatalf("future value changed\nbefore: %#v\nafter:  %#v", beforeFuture, afterFuture)
	}
}

func TestUnknownValueEditorsRejectOpaqueValue(t *testing.T) {
	current := json.RawMessage(`{
		"parameters": {
			"future": {
				"defaultValue": {
					"someFutureValue": {"id": "future-1"}
				},
				"valueType": "STRING"
			}
		}
	}`)
	mutations := []Mutation{
		SetStringParameterValue("", "future", "default", "replacement"),
		EditParameterDetails(ParameterDetailsEdit{
			ParamKey:      "future",
			NextParamKey:  "future",
			NextValueType: "STRING",
			ValueEdits:    []ParameterValueEdit{{Label: "default", NextValue: "replacement"}},
		}),
	}
	for _, mutation := range mutations {
		if _, err := BuildMutatedRemoteConfig(current, "", mutation); err == nil || !strings.Contains(err.Error(), "supports only plain values") {
			t.Fatalf("opaque value mutation error = %v", err)
		}
	}
}

func TestStructuralMutationsRejectManagedValues(t *testing.T) {
	current := json.RawMessage(`{
		"parameters": {
			"managed": {
				"defaultValue": {
					"rolloutValue": {"rolloutId":"rollout-1","value":"20","percent":10}
				}
			}
		}
	}`)
	mutations := []Mutation{
		DeleteParameter("", "managed"),
		RenameParameter("", "managed", "renamed"),
		MoveParameter("", "managed", "group"),
		DuplicateParameterNamed("", "managed", "copy"),
	}
	for _, mutation := range mutations {
		if _, err := BuildMutatedRemoteConfig(current, "", mutation); err == nil || !strings.Contains(err.Error(), "rollout value is read-only") {
			t.Fatalf("managed structural mutation error = %v", err)
		}
	}
}

func managedExperimentValue(t *testing.T, raw json.RawMessage) any {
	return remoteConfigValueOption(t, raw, "managed", "experimentValue")
}

func remoteConfigValueOption(t *testing.T, raw json.RawMessage, parameter, option string) any {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	parameters := document["parameters"].(map[string]any)
	entry := parameters[parameter].(map[string]any)
	defaultValue := entry["defaultValue"].(map[string]any)
	return defaultValue[option]
}

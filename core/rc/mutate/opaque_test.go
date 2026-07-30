package mutate

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestEnsureOpaqueValuesUnchangedAllowsMetadataAndPlainValueChanges(t *testing.T) {
	current := opaqueMutationConfig()
	final, err := firebase.CloneRemoteConfig(current)
	if err != nil {
		t.Fatal(err)
	}
	param := final.Parameters["managed"]
	param.Description = "updated"
	param.ConditionalValues["plain"] = firebase.RemoteConfigValue{Value: "updated"}
	final.Parameters["managed"] = param

	if err := EnsureOpaqueValuesUnchanged(current, final); err != nil {
		t.Fatalf("metadata/plain mutation rejected: %v", err)
	}
}

func TestEnsureOpaqueValuesUnchangedRejectsEveryOpaqueMutation(t *testing.T) {
	tests := []struct {
		name string
		edit func(*firebase.RemoteConfig)
		want string
	}{
		{
			name: "replace rollout",
			edit: func(cfg *firebase.RemoteConfig) {
				param := cfg.Parameters["managed"]
				param.DefaultValue = &firebase.RemoteConfigValue{Value: "20"}
				cfg.Parameters["managed"] = param
			},
			want: "rollout value",
		},
		{
			name: "remove personalization",
			edit: func(cfg *firebase.RemoteConfig) {
				param := cfg.Parameters["managed"]
				delete(param.ConditionalValues, "personalized")
				cfg.Parameters["managed"] = param
			},
			want: "personalization value",
		},
		{
			name: "relocate experiment",
			edit: func(cfg *firebase.RemoteConfig) {
				param := cfg.Parameters["managed"]
				value := param.ConditionalValues["experiment"]
				delete(param.ConditionalValues, "experiment")
				param.ConditionalValues["renamed"] = value
				cfg.Parameters["managed"] = param
			},
			want: "A/B test value",
		},
		{
			name: "change managed parameter type",
			edit: func(cfg *firebase.RemoteConfig) {
				param := cfg.Parameters["managed"]
				param.ValueType = "NUMBER"
				cfg.Parameters["managed"] = param
			},
			want: "value type is read-only",
		},
		{
			name: "add unknown",
			edit: func(cfg *firebase.RemoteConfig) {
				cfg.Parameters["future"] = firebase.RemoteConfigParam{
					DefaultValue: &firebase.RemoteConfigValue{
						UnknownValueOption: "futureValue",
						UnknownValue:       json.RawMessage(`{"id":"future-1"}`),
					},
				}
			},
			want: `unknown value option "futureValue"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := opaqueMutationConfig()
			final, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				t.Fatal(err)
			}
			tt.edit(final)
			err = EnsureOpaqueValuesUnchanged(current, final)
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), "read-only") {
				t.Fatalf("error = %v, want read-only %q error", err, tt.want)
			}
		})
	}
}

func opaqueMutationConfig() *firebase.RemoteConfig {
	return &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"managed": {
			DefaultValue: &firebase.RemoteConfigValue{
				RolloutValue: json.RawMessage(`{"rolloutId":"rollout-1","value":"20","percent":10}`),
			},
			ConditionalValues: map[string]firebase.RemoteConfigValue{
				"plain":        {Value: "before"},
				"personalized": {PersonalizationValue: json.RawMessage(`{"personalizationId":"p-1"}`)},
				"experiment":   {ExperimentValue: json.RawMessage(`{"experimentId":"e-1"}`)},
			},
		},
	}}
}

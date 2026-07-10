package firebase

import (
	"strings"
	"testing"
)

func TestParseCloneRemoteConfigDeepCopiesParsedConfig(t *testing.T) {
	raw := []byte(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`)

	cloned, err := ParseCloneRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseCloneRemoteConfig returned error: %v", err)
	}
	if cloned.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("flag value = %q, want on", cloned.Parameters["flag"].DefaultValue.Value)
	}

	cloned.Parameters["flag"] = RemoteConfigParam{}
	if !strings.Contains(string(raw), `"on"`) {
		t.Fatalf("raw config was mutated")
	}
}

func TestParseCloneRemoteConfigRejectsInvalidJSON(t *testing.T) {
	_, err := ParseCloneRemoteConfig([]byte(`{`))
	if err == nil {
		t.Fatalf("ParseCloneRemoteConfig accepted invalid JSON")
	}
}

func TestCloneRemoteConfigDeepCopiesMaps(t *testing.T) {
	original := &RemoteConfig{
		Parameters: map[string]RemoteConfigParam{
			"flag": {
				DefaultValue: &RemoteConfigValue{Value: "on"},
			},
		},
	}

	cloned, err := CloneRemoteConfig(original)
	if err != nil {
		t.Fatalf("CloneRemoteConfig returned error: %v", err)
	}
	if cloned == original {
		t.Fatalf("CloneRemoteConfig returned same pointer")
	}
	original.Parameters["flag"] = RemoteConfigParam{}
	if cloned.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("CloneRemoteConfig did not deep-copy parameters")
	}
}

func TestCloneRemoteConfigNil(t *testing.T) {
	cloned, err := CloneRemoteConfig(nil)
	if err != nil {
		t.Fatalf("CloneRemoteConfig(nil) returned error: %v", err)
	}
	if cloned == nil {
		t.Fatalf("CloneRemoteConfig(nil) = nil, want empty config")
	}
}

func TestMarshalRemoteConfig(t *testing.T) {
	got, err := MarshalRemoteConfig(&RemoteConfig{
		Parameters: map[string]RemoteConfigParam{
			"flag": {DefaultValue: &RemoteConfigValue{Value: "on"}},
		},
	})
	if err != nil {
		t.Fatalf("MarshalRemoteConfig returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("MarshalRemoteConfig returned empty payload")
	}
}

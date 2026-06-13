package shared

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestCloneRemoteConfigDeepCopiesMaps(t *testing.T) {
	original := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "old"},
			},
		},
	}

	cloned := CloneRemoteConfig(original)
	cloned.Parameters["flag"] = firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: "new"},
	}

	if original.Parameters["flag"].DefaultValue.Value != "old" {
		t.Fatalf("CloneRemoteConfig did not deep-copy parameters")
	}
}

func TestCloneRemoteConfigNil(t *testing.T) {
	cloned := CloneRemoteConfig(nil)
	if cloned == nil {
		t.Fatalf("CloneRemoteConfig(nil) = nil, want empty config")
	}
}

func TestMarshalRemoteConfig(t *testing.T) {
	got, err := MarshalRemoteConfig(&firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "on"}},
		},
	})
	if err != nil {
		t.Fatalf("MarshalRemoteConfig returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("MarshalRemoteConfig returned empty payload")
	}
}

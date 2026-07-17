package groups

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestNamedGroupMutationSkipsProjectsWithoutTarget(t *testing.T) {
	called := false
	mutation := namedGroupMutation("shared", func(*firebase.RemoteConfig, string) (bool, error) {
		called = true
		return true, nil
	})
	changed, err := mutation(&firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"other": {}}})
	if err != nil || changed || called {
		t.Fatalf("mutation = changed %v, called %v, err %v; want skipped", changed, called, err)
	}
}

func TestNamedGroupMutationResolvesCaseInsensitiveNamePerProject(t *testing.T) {
	var resolved string
	mutation := namedGroupMutation("SHARED", func(_ *firebase.RemoteConfig, name string) (bool, error) {
		resolved = name
		return true, nil
	})
	changed, err := mutation(&firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"Shared": {}}})
	if err != nil || !changed || resolved != "Shared" {
		t.Fatalf("mutation = changed %v, resolved %q, err %v", changed, resolved, err)
	}
}

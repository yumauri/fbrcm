package shared

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestParseRemoteConfigOrder(t *testing.T) {
	raw := []byte(`{"version":{"versionNumber":"7"},"parameters":{"b":{"conditionalValues":{"second":{"value":"2"},"first":{"value":"1"}},"defaultValue":{"value":"b"}},"a":{"defaultValue":{"value":"a"}}},"parameterGroups":{"group-b":{"parameters":{"g2":{"defaultValue":{"value":"g2"}}, "g1":{"defaultValue":{"value":"g1"}}}}}}`)

	order, err := ParseRemoteConfigOrder(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfigOrder returned error: %v", err)
	}
	assertStrings(t, "top level", order.TopLevel, []string{"version", "parameters", "parameterGroups"})
	assertStrings(t, "parameters", order.Parameters, []string{"b", "a"})
	assertStrings(t, "groups", order.Groups, []string{"group-b"})
	assertStrings(t, "group parameters", order.GroupParameters["group-b"], []string{"g2", "g1"})
	assertStrings(t, "conditional values", order.ConditionalValues["b"], []string{"second", "first"})
	if got := string(order.VersionRaw); got != `{"versionNumber":"7"}` {
		t.Fatalf("VersionRaw = %s", got)
	}
}

func TestMarshalPrettyRemoteConfigWithOrder(t *testing.T) {
	raw := []byte(`{"version":{"versionNumber":"7"},"parameters":{"b":{"defaultValue":{"value":"b"}},"a":{"defaultValue":{"value":"a"}}}}`)
	order, err := ParseRemoteConfigOrder(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfigOrder returned error: %v", err)
	}
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}

	out, err := MarshalPrettyRemoteConfigWithOrder(cfg, order)
	if err != nil {
		t.Fatalf("MarshalPrettyRemoteConfigWithOrder returned error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid json:\n%s", out)
	}
	text := string(out)
	if strings.Index(text, `"version"`) > strings.Index(text, `"parameters"`) {
		t.Fatalf("top-level order was not preserved:\n%s", text)
	}
	if strings.Index(text, `"b"`) > strings.Index(text, `"a"`) {
		t.Fatalf("parameter order was not preserved:\n%s", text)
	}
}

func TestMarshalPrettyRemoteConfigWithOrderNilConfig(t *testing.T) {
	out, err := MarshalPrettyRemoteConfigWithOrder(nil, RemoteConfigOrder{})
	if err != nil {
		t.Fatalf("MarshalPrettyRemoteConfigWithOrder returned error: %v", err)
	}
	if got := string(out); got != "{}\n" {
		t.Fatalf("nil config output = %q, want empty object with newline", got)
	}
}

func TestParseRemoteConfigOrderRejectsInvalidJSON(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`[`),
		[]byte(`[]`),
		[]byte(`{"parameters":{}} trailing`),
	} {
		if _, err := ParseRemoteConfigOrder(raw); err == nil {
			t.Fatalf("ParseRemoteConfigOrder(%q) returned nil error", string(raw))
		}
	}
}

func TestOrderedKeysPreservesPreferredThenSortsRest(t *testing.T) {
	got := orderedKeys(map[string]int{
		"zeta":  1,
		"alpha": 2,
		"beta":  3,
	}, []string{"missing", "beta"})
	assertStrings(t, "ordered keys", got, []string{"beta", "alpha", "zeta"})
}

func TestRemoteConfigFieldPresent(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{"flag": {}},
		Version:    firebase.RemoteConfigVersion{VersionNumber: "42"},
	}

	if !remoteConfigFieldPresent(cfg, "parameters") {
		t.Fatalf("parameters field reported absent")
	}
	if !remoteConfigFieldPresent(cfg, "version") {
		t.Fatalf("version field reported absent")
	}
	if remoteConfigFieldPresent(cfg, "conditions") {
		t.Fatalf("empty conditions field reported present")
	}
	if remoteConfigFieldPresent(cfg, "unknown") {
		t.Fatalf("unknown field reported present")
	}
}

func assertStrings(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %#v, want %#v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %#v, want %#v", label, got, want)
		}
	}
}

package publication

import (
	"encoding/json"
	"strings"
	"testing"
)

func testRemoteConfig(value string) json.RawMessage {
	return json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"` + value + `"},"valueType":"STRING"}},"version":{"versionNumber":"1"}}`)
}

func testPlan(t *testing.T) *Plan {
	t.Helper()
	plan := New("test", "update", "stateless", nil)
	plan.Targets = append(plan.Targets, Target{
		Target: "demo", ProjectID: "demo", Template: "client", Action: ActionPublish,
		Base:       Snapshot{Version: "1", ETag: `"etag"`, RemoteConfig: testRemoteConfig("off")},
		Candidate:  Snapshot{RemoteConfig: testRemoteConfig("on")},
		Validation: Validation{Source: "firebase", ValidatedAt: plan.CreatedAt},
		Source:     Source{Kind: "direct"},
	})
	if err := Seal(plan); err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestPlanRoundTripAndIntegrity(t *testing.T) {
	plan := testPlan(t)
	raw, err := Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.PlanID != plan.PlanID || !strings.HasPrefix(parsed.PlanID, IDPrefix) {
		t.Fatalf("plan id = %q", parsed.PlanID)
	}

	tampered := strings.Replace(string(raw), `"off"`, `"maybe"`, 1)
	_, err = Parse([]byte(tampered))
	if err == nil || !strings.Contains(err.Error(), "digest") && !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("tampered plan error = %v", err)
	}
}

func TestPlanRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	raw, err := Marshal(testPlan(t))
	if err != nil {
		t.Fatal(err)
	}
	withUnknown := strings.Replace(string(raw), `"kind":`, `"unknown":true,"kind":`, 1)
	if _, err := Parse([]byte(withUnknown)); err == nil {
		t.Fatal("unknown plan field accepted")
	}
	if _, err := Parse(append(raw, []byte(`{}`)...)); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}

func TestRemoteConfigDigestIgnoresVersionMetadata(t *testing.T) {
	a := testRemoteConfig("on")
	b := json.RawMessage(strings.Replace(string(a), `"1"`, `"99"`, 1))
	aDigest, err := RemoteConfigDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	bDigest, err := RemoteConfigDigest(b)
	if err != nil {
		t.Fatal(err)
	}
	if aDigest != bDigest {
		t.Fatalf("version metadata affected digest: %s != %s", aDigest, bDigest)
	}
}

func TestPlanRejectsInvalidExecutionAndValidationProvenance(t *testing.T) {
	plan := testPlan(t)
	plan.Execution.HooksEnabled = true
	plan.Execution.HookDefinitionSHA256 = "not-a-digest"
	if err := Seal(plan); err == nil || !strings.Contains(err.Error(), "hook_definition_sha256") {
		t.Fatalf("invalid hook digest error = %v", err)
	}

	plan = testPlan(t)
	plan.Targets[0].Validation.Source = "guessed"
	if err := Seal(plan); err == nil || !strings.Contains(err.Error(), "validation source") {
		t.Fatalf("invalid validation source error = %v", err)
	}
}

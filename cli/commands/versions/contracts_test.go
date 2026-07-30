package versions

import "testing"

func TestVersionPublishJSONRepresentsNoOp(t *testing.T) {
	payload := versionPublishJSON("demo", false, "7", "7", "", nil, true, false)
	if payload["operation"] != "rollback" || payload["changed"] != false || payload["dry_run"] != true {
		t.Fatalf("no-op payload = %#v", payload)
	}
	if payload["published_version"] != nil {
		t.Fatalf("published_version = %#v, want nil", payload["published_version"])
	}
	if _, ok := payload["change_note"]; ok {
		t.Fatalf("rollback payload unexpectedly contains change_note: %#v", payload)
	}

	payload = versionPublishJSON("demo", true, "7", "3", "8", nil, false, true)
	if payload["operation"] != "restore" || payload["changed"] != true || payload["published_version"] != "8" {
		t.Fatalf("changed payload = %#v", payload)
	}
	if value, ok := payload["change_note"]; !ok || value != (*string)(nil) {
		t.Fatalf("restore change_note = %#v, present=%v", value, ok)
	}
}

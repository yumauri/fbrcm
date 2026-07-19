package versions

import "testing"

func TestVersionPublishJSONRepresentsNoOp(t *testing.T) {
	payload := versionPublishJSON("demo", false, "7", "7", "", true, false)
	if payload["operation"] != "rollback" || payload["changed"] != false || payload["dry_run"] != true {
		t.Fatalf("no-op payload = %#v", payload)
	}
	if payload["published_version"] != nil {
		t.Fatalf("published_version = %#v, want nil", payload["published_version"])
	}

	payload = versionPublishJSON("demo", true, "7", "3", "8", false, true)
	if payload["operation"] != "restore" || payload["changed"] != true || payload["published_version"] != "8" {
		t.Fatalf("changed payload = %#v", payload)
	}
}

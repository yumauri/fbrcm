package versions

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestVersionPublishJSONRepresentsNoOp(t *testing.T) {
	payload := versionPublishJSON("demo", false, "7", "7", "", nil, true, false, true, core.ValidationSourceLocal)
	if payload.Operation != "rollback" || payload.Status != "unchanged" || payload.Changed || !payload.DryRun {
		t.Fatalf("no-op payload = %#v", payload)
	}
	if payload.PublishedVersion != nil {
		t.Fatalf("published_version = %#v, want nil", payload.PublishedVersion)
	}
	if payload.ChangeNote != nil {
		t.Fatalf("rollback change_note = %#v, want nil", payload.ChangeNote)
	}
	payload = versionPublishJSON("demo", true, "7", "7", "", nil, false, false, true, core.ValidationSourceLocal)
	if payload.Status != "unchanged" || payload.Changed || payload.DryRun || payload.PublishedVersion != nil {
		t.Fatalf("live no-op payload = %#v", payload)
	}

	payload = versionPublishJSON("demo", true, "7", "3", "8", nil, false, true, true, core.ValidationSourceFirebase)
	if payload.Operation != "restore" || payload.Status != "published" || !payload.Changed || payload.PublishedVersion == nil || *payload.PublishedVersion != "8" {
		t.Fatalf("changed payload = %#v", payload)
	}
	if payload.ChangeNote != nil {
		t.Fatalf("restore change_note = %#v", payload.ChangeNote)
	}
	if !payload.Validated || payload.ValidationSource != core.ValidationSourceFirebase {
		t.Fatalf("validation metadata = %#v/%#v", payload.Validated, payload.ValidationSource)
	}
}

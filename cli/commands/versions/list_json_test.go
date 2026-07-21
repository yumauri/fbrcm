package versions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestVersionListJSONIsPlainArray(t *testing.T) {
	result := core.RemoteConfigVersionList{
		Versions: []core.RemoteConfigVersionEntry{{
			RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "42"},
			Current:             true,
		}},
		NextPageToken: "next-page",
	}
	items := versionListJSON(result)
	raw, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "[") || strings.Contains(string(raw), `"versions"`) || strings.Contains(string(raw), `"project"`) || strings.Contains(string(raw), `"next_page_token"`) {
		t.Fatalf("version list JSON = %s", raw)
	}
	if len(items) != 1 || items[0].VersionNumber != "42" {
		t.Fatalf("version list items = %#v", items)
	}
	empty := versionListJSON(core.RemoteConfigVersionList{})
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty version list = %#v, want non-nil empty slice", empty)
	}
	emptyRaw, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	if string(emptyRaw) != "[]" {
		t.Fatalf("empty version list JSON = %s, want []", emptyRaw)
	}
}

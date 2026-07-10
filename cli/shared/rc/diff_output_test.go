package rc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRenderRemoteConfigDiffDetectsChanges(t *testing.T) {
	current := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "old"}},
		},
	}
	final := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "new"}},
		},
	}

	diff, changed := RenderRemoteConfigDiff(current, final)
	if !changed {
		t.Fatal("RenderRemoteConfigDiff changed = false, want true")
	}
	if !strings.Contains(diff, "flag") {
		t.Fatalf("diff = %q, want flag mention", diff)
	}
}

func TestRenderConflictPreviewAndChoiceValue(t *testing.T) {
	preview := RenderConflictPreview("default", "on", "off")
	if preview == "" || !strings.Contains(preview, "default") {
		t.Fatalf("RenderConflictPreview = %q", preview)
	}
	if got := RenderConflictChoiceValue("on"); got == "" {
		t.Fatal("RenderConflictChoiceValue returned empty string")
	}
}

func TestRenderRemovedParameterDetail(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	param := firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: "gone"},
		Description:  "remove me",
	}
	got := RenderRemovedParameterDetail("flag", "", param)
	if !strings.Contains(got, "flag") || !strings.Contains(got, "gone") {
		t.Fatalf("RenderRemovedParameterDetail = %q", got)
	}
}

func TestWriteOrderPreservingRemoteConfigStdout(t *testing.T) {
	raw := []byte(`{"version":{"versionNumber":"1"},"parameters":{"b":{"defaultValue":{"value":"b"}},"a":{"defaultValue":{"value":"a"}}}}`)
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig = %v", err)
	}
	cfg.Parameters["a"] = firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "a2"}}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := WriteOrderPreservingRemoteConfigStdout(cmd, cfg, raw); err != nil {
		t.Fatalf("WriteOrderPreservingRemoteConfigStdout = %v", err)
	}
	text := out.String()
	if strings.Index(text, `"b"`) > strings.Index(text, `"a"`) {
		t.Fatalf("parameter order not preserved:\n%s", text)
	}
	if !strings.Contains(text, `"a2"`) {
		t.Fatalf("updated value missing:\n%s", text)
	}
}

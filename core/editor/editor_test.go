package editor

import (
	"runtime"
	"slices"
	"testing"

	"github.com/yumauri/fbrcm/core/env"
)

func TestResolvePrecedence(t *testing.T) {
	t.Setenv(env.Editor, "fbrcm-editor")
	t.Setenv("VISUAL", "visual-editor")
	t.Setenv("EDITOR", "editor")
	if got := Resolve(""); got != "fbrcm-editor" {
		t.Fatalf("Resolve = %q, want fbrcm-editor", got)
	}
	if got := Resolve(" explicit "); got != "explicit" {
		t.Fatalf("Resolve explicit = %q, want explicit", got)
	}
}

func TestCommandPassesPathAsSeparateShellArgument(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix shell command assertion")
	}
	t.Setenv("SHELL", "/bin/example-shell")
	cmd := Command("code --wait", "/tmp/value with spaces.json")
	want := []string{"/bin/example-shell", "-c", `exec code --wait "$1"`, "fbrcm-editor", "/tmp/value with spaces.json"}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("Command args = %#v, want %#v", cmd.Args, want)
	}
}

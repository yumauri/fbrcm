package editor

import (
	"os/exec"
	"runtime"
	"slices"
	"strings"
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

func TestCommandPassesPathThroughEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix shell command assertion")
	}
	t.Setenv("SHELL", "/bin/example-shell")
	cmd := Command("code --wait", "/tmp/value with spaces.json")
	want := []string{"/bin/example-shell", "-c", `exec code --wait "$FBRCM_EDITOR_FILE"`}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("Command args = %#v, want %#v", cmd.Args, want)
	}
	if got := environmentValue(cmd.Env, editorFileEnv); got != "/tmp/value with spaces.json" {
		t.Fatalf("%s = %q, want staged path", editorFileEnv, got)
	}
}

func TestCommandPassesPathToPOSIXAndFishShells(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix shell execution assertion")
	}
	shells := []string{"/bin/sh"}
	if fish, err := exec.LookPath("fish"); err == nil {
		shells = append(shells, fish)
	}
	const path = "/tmp/value with spaces;$(not-a-command).json"
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			t.Setenv("SHELL", shell)
			output, err := Command("printf '%s'", path).Output()
			if err != nil {
				t.Fatalf("run editor command: %v", err)
			}
			if got := string(output); got != path {
				t.Fatalf("editor path = %q, want %q", got, path)
			}
		})
	}
}

func environmentValue(environment []string, name string) string {
	prefix := name + "="
	for _, entry := range slices.Backward(environment) {
		if value, ok := strings.CutPrefix(entry, prefix); ok {
			return value
		}
	}
	return ""
}

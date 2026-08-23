package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/env"
)

const expectedPlainVersion = `▄─┐▄
█- █─▄ ▄─▄ ▄── ▄─▄─▄
▀  ▀─▀ ▀   ▀── ▀ ▀ ▀
fbrcm 1.2.3 (commit abc123, built 2026-06-14)
Victor Didenko <yumaa.verdin@gmail.com> (https://yumaa.name)
`

func TestRootVersionUsesFirebaseGradient(t *testing.T) {
	t.Setenv(env.NoColor, "")
	output := executeVersion(t, "-v")

	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("colored version output has no ANSI styling: %q", output)
	}
	if plain := ansi.Strip(output); plain != expectedPlainVersion {
		t.Fatalf("plain version output = %q, want %q", plain, expectedPlainVersion)
	}
}

func TestRootVersionHonorsNoColor(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	if output := executeVersion(t, "--version"); output != expectedPlainVersion {
		t.Fatalf("NO_COLOR version output = %q, want %q", output, expectedPlainVersion)
	}
}

func TestRootJSONVersionIsUnstyled(t *testing.T) {
	t.Setenv(env.NoColor, "")
	if output := executeVersion(t, "--version", "--json"); output != expectedPlainVersion {
		t.Fatalf("JSON version text = %q, want unstyled %q", output, expectedPlainVersion)
	}
}

func executeVersion(t *testing.T, args ...string) string {
	t.Helper()
	cmd := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return output.String()
}

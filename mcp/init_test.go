package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/ops/contract"
)

func TestStandaloneJSONStartupAndHelp(t *testing.T) {
	state := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(state, "config"))
	t.Setenv(env.CacheDir, filepath.Join(state, "cache"))
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	for _, test := range []struct {
		args    []string
		code    int
		command string
	}{
		{[]string{"mcp", "--json"}, 2, "mcp"},
		{[]string{"mcp", "--stateless", "--json"}, 2, "mcp"},
		{[]string{"mcp", "--json", "--help"}, 0, "help"},
	} {
		var out, stderr bytes.Buffer
		code := Run(t.Context(), nil, "test", "", "", test.args, bytes.NewReader(nil), &out, &stderr)
		if code != test.code {
			t.Fatalf("%v: code=%d stdout=%s stderr=%s", test.args, code, out.String(), stderr.String())
		}
		var envelope contract.Envelope
		if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
			t.Fatalf("stdout is not one envelope: %s: %v", out.String(), err)
		}
		if envelope.Command != test.command || envelope.ExitCode != test.code {
			t.Fatalf("unexpected envelope: %#v", envelope)
		}
		if test.command == "mcp" {
			entries, err := os.ReadDir(state)
			if err != nil || len(entries) != 0 {
				t.Fatalf("JSON rejection touched local state: %v %v", entries, err)
			}
		}
	}
}

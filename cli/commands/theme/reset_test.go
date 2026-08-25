package theme

import (
	"bytes"
	"strings"
	"testing"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func TestSwitchBuiltInAndResetClearSelections(t *testing.T) {
	setupThemeCommandTest(t)
	if _, err := coreconfig.ImportTheme("firebase", []byte("[colors]\nprimary = \"#FFC400\"\n")); err != nil {
		t.Fatal(err)
	}
	if err := coreconfig.SetConfiguredTheme("firebase", coreconfig.ThemeScopeGlobal); err != nil {
		t.Fatal(err)
	}

	switchCmd := newSwitchCommand()
	var output bytes.Buffer
	switchCmd.SetOut(&output)
	switchCmd.SetArgs([]string{coreconfig.BuiltInThemeName})
	if err := switchCmd.Execute(); err != nil {
		t.Fatalf("switch built-in = %v", err)
	}
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil || resolved.Global.Config.Theme != "" {
		t.Fatalf("global selection after switch = %#v, %v", resolved.Global.Config, err)
	}
	if !strings.Contains(output.String(), "switched: built-in") {
		t.Fatalf("switch output = %q", output.String())
	}

	resetCmd := newResetCommand()
	output.Reset()
	resetCmd.SetOut(&output)
	if err := resetCmd.Execute(); err != nil {
		t.Fatalf("unchanged reset = %v", err)
	}
	if !strings.Contains(output.String(), "already uses its default") {
		t.Fatalf("reset output = %q", output.String())
	}
}

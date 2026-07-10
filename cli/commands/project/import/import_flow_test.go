package importpkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestReadRemoteConfigFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "import.json")
	if err := os.WriteFile(path, []byte(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("from", "", "")
	if err := cmd.Flags().Set("from", path); err != nil {
		t.Fatal(err)
	}

	raw, err := readRemoteConfig(cmd)
	if err != nil {
		t.Fatalf("readRemoteConfig = %v", err)
	}
	if !strings.Contains(string(raw), "flag") {
		t.Fatalf("raw = %s", raw)
	}
}

func TestReadImportOptionsMergeResolveValidation(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringArray("group", nil, "")
	cmd.Flags().StringArray("filter", nil, "")
	cmd.Flags().String("expr", "", "")
	cmd.Flags().String("search", "", "")
	cmd.Flags().Bool("remove-all-conditions", false, "")
	cmd.Flags().Bool("remove-project-specific-conditions", false, "")
	cmd.Flags().Bool("merge", false, "")
	cmd.Flags().Bool("override", false, "")
	cmd.Flags().String("merge-resolve", "", "")

	if err := cmd.Flags().Set("merge-resolve", "current"); err != nil {
		t.Fatal(err)
	}
	if _, err := readImportOptions(cmd); err == nil || !strings.Contains(err.Error(), "requires --merge") {
		t.Fatalf("readImportOptions = %v, want merge required error", err)
	}

	if err := cmd.Flags().Set("merge", "true"); err != nil {
		t.Fatal(err)
	}
	opts, err := readImportOptions(cmd)
	if err != nil {
		t.Fatalf("readImportOptions merge = %v", err)
	}
	if opts.mergeResolve != "current" {
		t.Fatalf("mergeResolve = %q, want current", opts.mergeResolve)
	}
}

func TestChooseImportStrategyFlags(t *testing.T) {
	override, err := chooseImportStrategy(importOptions{override: true})
	if err != nil || override != importStrategyOverride {
		t.Fatalf("override strategy = %q err=%v", override, err)
	}
	merge, err := chooseImportStrategy(importOptions{merge: true})
	if err != nil || merge != importStrategyMerge {
		t.Fatalf("merge strategy = %q err=%v", merge, err)
	}
}

func TestBuildFinalImportConfigEmptyCurrentUsesImport(t *testing.T) {
	importCfg := mergeImportFixture()
	final, err := buildFinalImportConfig(&cobra.Command{}, &firebase.RemoteConfig{}, importCfg, importOptions{override: true})
	if err != nil {
		t.Fatalf("buildFinalImportConfig = %v", err)
	}
	if _, ok := final.Parameters["root_new"]; !ok {
		t.Fatalf("final config = %+v, want imported params", final.Parameters)
	}
}

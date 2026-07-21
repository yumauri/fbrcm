package importpkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

type importRoundTripFunc func(*http.Request) (*http.Response, error)

func (f importRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

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
	cmd.Flags().Bool("keep-portable-conditions-only", false, "")
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
	override, err := chooseImportStrategy(&cobra.Command{}, importOptions{override: true})
	if err != nil || override != importStrategyOverride {
		t.Fatalf("override strategy = %q err=%v", override, err)
	}
	merge, err := chooseImportStrategy(&cobra.Command{}, importOptions{merge: true})
	if err != nil || merge != importStrategyMerge {
		t.Fatalf("merge strategy = %q err=%v", merge, err)
	}
}

func TestImportConditionCountLine(t *testing.T) {
	if got := importConditionCountLine(5, 2); got != "Import conditions: 2 kept · 3 removed" {
		t.Fatalf("condition count line = %q", got)
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

func TestRunYesJSONDraftProducesStructuredResultWithoutPrompt(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	project := core.Project{Name: "Demo", ProjectID: "demo", AuthID: "main"}
	if err := config.SaveProjects([]config.Project{project}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current := `{"parameters":{"old":{"defaultValue":{"value":"old"}}},"version":{"versionNumber":"1"}}`
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: importRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := ""
		switch {
		case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, ":listVersions"):
			body = `{"versions":[{"versionNumber":"1"}]}`
		case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/remoteConfig"):
			body = current
		default:
			return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{"etag-1"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})})
	svc.InjectFirebaseService("main", client)

	sourcePath := filepath.Join(root, "import.json")
	if err := os.WriteFile(sourcePath, []byte(`{"parameters":{"new":{"defaultValue":{"value":"new"}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newImportFlowTestCommand()
	for name, value := range map[string]string{
		"from": sourcePath, "draft": "true", "override": "true", "yes": "true", "json": "true",
	} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(""))
	if err := Run(cmd, svc, project); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("JSON stderr = %q, want empty", stderr.String())
	}
	var result importResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode JSON result %q: %v", stdout.String(), err)
	}
	if result.ProjectID != "demo" || result.Status != "drafted" || !result.Changed || !result.Draft || result.DryRun {
		t.Fatalf("result = %+v", result)
	}
	if _, ok, err := svc.LoadDraft("demo"); err != nil || !ok {
		t.Fatalf("LoadDraft = ok %v, err %v", ok, err)
	}
}

func newImportFlowTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("from", "", "")
	cmd.Flags().StringArray("group", nil, "")
	cmd.Flags().StringArray("filter", nil, "")
	cmd.Flags().String("expr", "", "")
	cmd.Flags().String("search", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("draft", false, "")
	cmd.Flags().Bool("remove-all-conditions", false, "")
	cmd.Flags().Bool("keep-portable-conditions-only", false, "")
	cmd.Flags().Bool("merge", false, "")
	cmd.Flags().Bool("override", false, "")
	cmd.Flags().String("merge-resolve", "", "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("json", false, "")
	return cmd
}

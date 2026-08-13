package versions

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func TestCachedVersionListUsesOnlyLocalProjectResolution(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := coreconfig.SwitchProfile(coreconfig.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := coreconfig.SaveProjects([]coreconfig.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := coreconfig.SaveParametersCache("demo", &coreconfig.ParametersCache{
		ETag: "etag-7", CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"7"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	registryPath := coreconfig.GetProjectsFilePath()
	before, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}

	cmd := New(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "demo", "--cached", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cached list = %v\n%s", err, out.String())
	}
	after, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("cached version list rewrote the projects registry")
	}
}

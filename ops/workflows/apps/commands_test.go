package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

type fakeAppReader struct {
	apps    []core.FirebaseApp
	details core.FirebaseAppDetails
	config  core.FirebaseAppConfig
}

func (f fakeAppReader) ListFirebaseApps(context.Context, string, core.ListFirebaseAppsOptions) ([]core.FirebaseApp, error) {
	return f.apps, nil
}

func (f fakeAppReader) GetFirebaseApp(context.Context, string, string) (core.FirebaseAppDetails, error) {
	return f.details, nil
}

func (f fakeAppReader) GetFirebaseAppConfig(context.Context, string, string) (core.FirebaseAppConfig, error) {
	return f.config, nil
}

func TestAppsListCommandSuccess(t *testing.T) {
	svc := appsCommandTestCore(t)
	reader := fakeAppReader{apps: []core.FirebaseApp{{DisplayName: "Demo Android", Platform: core.AppPlatformAndroid, Namespace: "com.example", AppID: "android-id", State: "ACTIVE"}}}
	cmd := cliadapter.Command(newListDefinition(svc, reader))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"demo", "--platform", "android", "--filter", "=com.example"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Project: Demo (demo)") || !strings.Contains(out.String(), "Demo Android") {
		t.Fatalf("output =\n%s", out.String())
	}
}

func TestAppsShowCommandSuccess(t *testing.T) {
	svc := appsCommandTestCore(t)
	reader := fakeAppReader{details: core.FirebaseAppDetails{FirebaseApp: core.FirebaseApp{DisplayName: "Demo Web", Platform: core.AppPlatformWeb, AppID: "web-id", ResourceName: "projects/demo/webApps/w", State: "ACTIVE"}, ProjectID: "demo"}}
	cmd := cliadapter.Command(newShowDefinition(svc, reader))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"demo", "web-id"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Name: Demo Web") || !strings.Contains(out.String(), "Platform: web") {
		t.Fatalf("output =\n%s", out.String())
	}
}

func TestAppsConfigCommandSuccess(t *testing.T) {
	svc := appsCommandTestCore(t)
	reader := fakeAppReader{config: core.FirebaseAppConfig{App: core.FirebaseApp{ResourceName: "projects/demo/androidApps/a", AppID: "android-id", Platform: core.AppPlatformAndroid}, SuggestedFilename: "google-services.json", MediaType: "application/json", Contents: []byte("{\"project\":\"demo\"}\n")}}
	cmd := cliadapter.Command(newConfigDefinition(svc, reader))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"demo", "android-id"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "{\"project\":\"demo\"}\n" {
		t.Fatalf("config = %q", out.String())
	}
}

func TestAppsConfigCommandJSONArtifactAndPrivateFile(t *testing.T) {
	svc := appsCommandTestCore(t)
	reader := fakeAppReader{config: core.FirebaseAppConfig{App: core.FirebaseApp{ResourceName: "projects/demo/iosApps/i", AppID: "ios-id", Platform: core.AppPlatformIOS, State: "ACTIVE"}, SuggestedFilename: "GoogleService-Info.plist", MediaType: "application/x-plist", Contents: []byte("<plist/>\n")}}

	jsonCommand := cliadapter.Command(newConfigDefinition(svc, reader))
	var output bytes.Buffer
	jsonCommand.SetOut(&output)
	jsonCommand.SetArgs([]string{"demo", "ios-id", "--json"})
	if err := jsonCommand.Execute(); err != nil {
		t.Fatal(err)
	}
	var result appConfigResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode output: %v\n%s", err, output.String())
	}
	if result.SuggestedFilename != "GoogleService-Info.plist" || result.Artifact.Encoding != "utf-8" || result.Artifact.MediaType != "application/x-plist" {
		t.Fatalf("artifact = %#v", result)
	}

	destination := filepath.Join(t.TempDir(), "nested", "GoogleService-Info.plist")
	fileCommand := cliadapter.Command(newConfigDefinition(svc, reader))
	fileCommand.SetOut(&bytes.Buffer{})
	fileCommand.SetArgs([]string{"demo", "ios-id", "--to", destination})
	if err := fileCommand.Execute(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != config.PrivateFileMode {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func appsCommandTestCore(t *testing.T) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

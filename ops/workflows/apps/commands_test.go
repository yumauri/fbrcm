package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/ops/shared"
)

type fakeAppReader struct {
	apps             []core.FirebaseApp
	details          core.FirebaseAppDetails
	config           core.FirebaseAppConfig
	selectedProject  *string
	selectedAppQuery *string
	selectedOptions  *core.ListFirebaseAppsOptions
	source           core.AppCacheSource
	cachedAt         time.Time
	refreshError     error
	cacheError       error
}

func (f fakeAppReader) ReadFirebaseApps(_ context.Context, _ string, opts core.ListFirebaseAppsOptions) (core.FirebaseAppsResult, error) {
	if f.selectedOptions != nil {
		*f.selectedOptions = opts
	}
	return core.FirebaseAppsResult{Apps: f.apps, Source: fakeAppSource(f.source), CachedAt: f.cachedAt, RefreshError: f.refreshError, CacheError: f.cacheError}, nil
}

func (f fakeAppReader) ReadFirebaseApp(_ context.Context, projectID, appQuery string, opts core.ListFirebaseAppsOptions) (core.FirebaseAppDetailsResult, error) {
	if f.selectedProject != nil {
		*f.selectedProject = projectID
	}
	if f.selectedAppQuery != nil {
		*f.selectedAppQuery = appQuery
	}
	if f.selectedOptions != nil {
		*f.selectedOptions = opts
	}
	return core.FirebaseAppDetailsResult{App: f.details, Source: fakeAppSource(f.source), CachedAt: f.cachedAt, RefreshError: f.refreshError, CacheError: f.cacheError}, nil
}

func (f fakeAppReader) ReadFirebaseAppConfig(_ context.Context, projectID, appQuery string, opts core.ListFirebaseAppsOptions) (core.FirebaseAppConfigResult, error) {
	if f.selectedProject != nil {
		*f.selectedProject = projectID
	}
	if f.selectedAppQuery != nil {
		*f.selectedAppQuery = appQuery
	}
	if f.selectedOptions != nil {
		*f.selectedOptions = opts
	}
	return core.FirebaseAppConfigResult{Config: f.config, Source: fakeAppSource(f.source), CachedAt: f.cachedAt, RefreshError: f.refreshError, CacheError: f.cacheError}, nil
}

func fakeAppSource(source core.AppCacheSource) core.AppCacheSource {
	if source == "" {
		return core.AppCacheSourceFirebase
	}
	return source
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
	reader := fakeAppReader{details: core.FirebaseAppDetails{FirebaseApp: core.FirebaseApp{DisplayName: "Demo Web", Platform: core.AppPlatformWeb, Namespace: "web.example", AppID: "web-id", ResourceName: "projects/demo/webApps/w", State: "ACTIVE"}, ProjectID: "demo"}}
	cmd := cliadapter.Command(newShowDefinition(svc, reader))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"web.example", "--project", "dmo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Name: Demo Web") || !strings.Contains(out.String(), "Platform: web") {
		t.Fatalf("output =\n%s", out.String())
	}
}

func TestAppsHumanCommandsUseConfiguredNerdFontGlyphs(t *testing.T) {
	svc := appsCommandTestCore(t)
	enabled := true
	if err := config.SaveAppConfig(&config.AppConfig{NerdFontGlyphs: &enabled}); err != nil {
		t.Fatal(err)
	}

	list := cliadapter.Command(newListDefinition(svc, fakeAppReader{apps: []core.FirebaseApp{{DisplayName: "Demo Android", Platform: core.AppPlatformAndroid, AppID: "android-id", State: "ACTIVE"}}}))
	var listOutput bytes.Buffer
	list.SetOut(&listOutput)
	list.SetArgs([]string{"demo"})
	if err := list.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listOutput.String(), nerdFontAndroidGlyph+" android") {
		t.Fatalf("list output omits configured Nerd Font glyph:\n%s", listOutput.String())
	}

	show := cliadapter.Command(newShowDefinition(svc, fakeAppReader{details: core.FirebaseAppDetails{FirebaseApp: core.FirebaseApp{DisplayName: "Demo Web", Platform: core.AppPlatformWeb, AppID: "web-id", ResourceName: "projects/demo/webApps/w", State: "ACTIVE"}, ProjectID: "demo"}}))
	var showOutput bytes.Buffer
	show.SetOut(&showOutput)
	show.SetArgs([]string{"web-id", "--project", "demo"})
	if err := show.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(showOutput.String(), "Platform: "+nerdFontWebGlyph+" web") {
		t.Fatalf("show output omits configured Nerd Font glyph:\n%s", showOutput.String())
	}
}

func TestAppsConfigCommandSuccess(t *testing.T) {
	svc := appsCommandTestCore(t)
	reader := fakeAppReader{config: core.FirebaseAppConfig{App: core.FirebaseApp{ResourceName: "projects/demo/androidApps/a", AppID: "android-id", Platform: core.AppPlatformAndroid}, SuggestedFilename: "google-services.json", MediaType: "application/json", Contents: []byte("{\"project\":\"demo\"}\n")}}
	cmd := cliadapter.Command(newConfigDefinition(svc, reader))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"android-id", "--project", "demo"})
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
	jsonCommand.SetArgs([]string{"ios-id", "--project", "demo", "--json"})
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
	fileCommand.SetArgs([]string{"ios-id", "--project", "demo", "--to", destination})
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

func TestAppsShowInfersConfiguredProjectFromAppID(t *testing.T) {
	svc := appsCommandTestCore(t)
	var selectedProject, selectedAppQuery string
	reader := fakeAppReader{
		details:          core.FirebaseAppDetails{FirebaseApp: core.FirebaseApp{AppID: "1:123:web:c3", Platform: core.AppPlatformWeb}},
		selectedProject:  &selectedProject,
		selectedAppQuery: &selectedAppQuery,
	}
	cmd := cliadapter.Command(newShowDefinition(svc, reader))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"1:123:web:c3"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if selectedProject != "demo" || selectedAppQuery != "1:123:web:c3" {
		t.Fatalf("reader selection = project %q, app %q", selectedProject, selectedAppQuery)
	}
}

func TestAppsConfigInfersConfiguredProjectFromAppID(t *testing.T) {
	svc := appsCommandTestCore(t)
	var selectedProject string
	reader := fakeAppReader{
		config:          core.FirebaseAppConfig{App: core.FirebaseApp{AppID: "1:123:android:a1"}, Contents: []byte("{}\n")},
		selectedProject: &selectedProject,
	}
	cmd := cliadapter.Command(newConfigDefinition(svc, reader))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"1:123:android:a1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if selectedProject != "demo" {
		t.Fatalf("reader project = %q, want demo", selectedProject)
	}
}

func TestAppsShowRequiresProjectForNonAppID(t *testing.T) {
	svc := appsCommandTestCore(t)
	cmd := cliadapter.Command(newShowDefinition(svc, fakeAppReader{}))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"com.example.demo"})
	err := cmd.Execute()
	var argumentErr *shared.ArgumentError
	if !errors.As(err, &argumentErr) || !strings.Contains(err.Error(), "--project is required") {
		t.Fatalf("error = %#v", err)
	}
}

func TestAppsCacheFlagsReachReader(t *testing.T) {
	svc := appsCommandTestCore(t)
	var options core.ListFirebaseAppsOptions
	cmd := cliadapter.Command(newListDefinition(svc, fakeAppReader{selectedOptions: &options}))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"demo", "--cached", "--show-deleted"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !options.CachedOnly || !options.ShowDeleted || options.Update {
		t.Fatalf("options = %#v", options)
	}
}

func TestAppsCacheFlagsAreMutuallyExclusive(t *testing.T) {
	svc := appsCommandTestCore(t)
	cmd := cliadapter.Command(newShowDefinition(svc, fakeAppReader{}))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"app", "--project", "demo", "--cached", "--update"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected mutually exclusive cache flag error")
	}
}

func TestAppsCacheFlagsRejectStatelessExecution(t *testing.T) {
	svc := appsCommandTestCore(t)
	cmd := cliadapter.Command(newListDefinition(svc, fakeAppReader{}))
	cmd.SetContext(core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy()))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"demo", "--cached"})
	err := cmd.Execute()
	var argumentErr *shared.ArgumentError
	if !errors.As(err, &argumentErr) || !strings.Contains(err.Error(), "--stateless") {
		t.Fatalf("error = %#v", err)
	}
}

func TestAppsListPublishesCacheProvenanceAndWarnings(t *testing.T) {
	svc := appsCommandTestCore(t)
	cachedAt := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)
	cmd := cliadapter.Command(newListDefinition(svc, fakeAppReader{apps: []core.FirebaseApp{{AppID: "app"}}, source: core.AppCacheSourceCacheStale, cachedAt: cachedAt, refreshError: errors.New("offline")}))
	cmd.SetContext(shared.WithMachineState(context.Background()))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"demo", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result appListResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].Source != core.AppCacheSourceCacheStale || result.Items[0].CachedAt == nil || !result.Items[0].CachedAt.Equal(cachedAt) {
		t.Fatalf("result = %#v", result)
	}
	warnings := shared.MachineWarnings(cmd)
	if len(warnings) != 1 || warnings[0].Code != "cache.stale" {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func appsCommandTestCore(t *testing.T) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoColor, "1")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

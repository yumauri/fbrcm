package apps

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
)

func TestRenderAppsTableUsesNaturalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	apps := []core.FirebaseApp{{DisplayName: "Android", Platform: core.AppPlatformAndroid, Namespace: "com.example", AppID: "1:2:android:3", State: "ACTIVE"}}
	got := renderAppsTableAtWidth(apps, 200)
	if !strings.Contains(got, "com.example") || !strings.Contains(got, "1:2:android:3") || strings.Contains(got, "…") {
		t.Fatalf("table =\n%s", got)
	}
	if width := lipgloss.Width(got); width >= 200 {
		t.Fatalf("natural table width = %d", width)
	}
}

func TestRenderAppsTableEllipsizesFlexibleColumns(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	apps := []core.FirebaseApp{{DisplayName: "A very long application name", Platform: core.AppPlatformAndroid, Namespace: "com.example.application", AppID: "1:123456789:android:abcdef", State: "ACTIVE"}}
	got := renderAppsTableAtWidth(apps, 54)
	if !strings.Contains(got, "…") {
		t.Fatalf("narrow table did not ellipsize:\n%s", got)
	}
	if width := lipgloss.Width(got); width > 54 {
		t.Fatalf("narrow table width = %d, want <= 54\n%s", width, got)
	}
}

func TestRenderAppDetailsIncludesPlatformFields(t *testing.T) {
	packageName := "com.example"
	base := core.FirebaseApp{DisplayName: "Demo", Platform: core.AppPlatformAndroid, AppID: "app-id", ResourceName: "projects/demo/androidApps/a", Namespace: packageName, State: "ACTIVE"}
	app := core.FirebaseAppDetails{
		FirebaseApp: base,
		ProjectID:   "demo", PackageName: &packageName, SHA1Hashes: []string{"AA:BB"},
	}
	got := renderAppDetails(app)
	for _, expected := range []string{"Name: Demo", "Platform: android", "Package name: com.example", "SHA-1 hashes: AA:BB"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("details missing %q:\n%s", expected, got)
		}
	}
}

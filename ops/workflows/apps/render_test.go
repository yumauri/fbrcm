package apps

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

func TestRenderAppsTableUsesNaturalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	apps := []core.FirebaseApp{{DisplayName: "Android", Platform: core.AppPlatformAndroid, Namespace: "com.example", AppID: "1:2:android:3", State: "ACTIVE"}}
	got := renderAppsTableAtWidth(apps, 200, false)
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
	got := renderAppsTableAtWidth(apps, 54, false)
	if !strings.Contains(got, "…") {
		t.Fatalf("narrow table did not ellipsize:\n%s", got)
	}
	if width := lipgloss.Width(got); width > 54 {
		t.Fatalf("narrow table width = %d, want <= 54\n%s", width, got)
	}
}

func TestRenderAppDetailsIncludesPlatformFields(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	packageName := "com.example"
	base := core.FirebaseApp{DisplayName: "Demo", Platform: core.AppPlatformAndroid, AppID: "app-id", ResourceName: "projects/demo/androidApps/a", Namespace: packageName, State: "ACTIVE"}
	app := core.FirebaseAppDetails{
		FirebaseApp: base,
		ProjectID:   "demo", PackageName: &packageName, SHA1Hashes: []string{"AA:BB"},
	}
	got := renderAppDetails(app, false)
	for _, expected := range []string{"Name: Demo", "Platform: android", "Package name: com.example", "SHA-1 hashes: AA:BB"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("details missing %q:\n%s", expected, got)
		}
	}
}

func TestRenderAppsColorsPlatformsAndAppIDPlatformSegments(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	tests := []struct {
		platform core.AppPlatform
		appID    string
		color    color.Color
	}{
		{core.AppPlatformAndroid, "1:123:android:aaa", clistyles.FirebaseAndroid},
		{core.AppPlatformIOS, "1:123:ios:bbb", clistyles.FirebaseIOS},
		{core.AppPlatformWeb, "1:123:web:ccc", clistyles.FirebaseWeb},
	}
	for _, test := range tests {
		label := platformLabel(test.platform, true)
		styledLabel := renderPlatformLabel(label, test.platform)
		wantLabel := lipgloss.NewStyle().Foreground(test.color).Render(label)
		if styledLabel != wantLabel {
			t.Errorf("%s platform label = %q, want %q", test.platform, styledLabel, wantLabel)
		}

		styledID := renderAppID(test.appID, lipgloss.NewStyle())
		name := string(test.platform)
		wantID := strings.Replace(test.appID, name, lipgloss.NewStyle().Foreground(test.color).Render(name), 1)
		if styledID != wantID {
			t.Errorf("%s App ID = %q, want %q", test.platform, styledID, wantID)
		}

		table := renderAppsTableAtWidth([]core.FirebaseApp{{DisplayName: "Demo", Platform: test.platform, AppID: test.appID, State: "ACTIVE"}}, 200, true)
		if prefix := appStylePrefix(lipgloss.NewStyle().Foreground(test.color)); prefix == "" || !strings.Contains(table, prefix) {
			t.Errorf("%s table does not contain platform color:\n%s", test.platform, table)
		}
		plain := ansi.Strip(table)
		if !strings.Contains(plain, label) || !strings.Contains(plain, test.appID) {
			t.Errorf("%s colored table changed content:\n%s", test.platform, plain)
		}
	}
}

func TestRenderAppsColorsRespectNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	appID := "1:123:web:ccc"
	label := platformLabel(core.AppPlatformWeb, true)
	if got := renderPlatformLabel(label, core.AppPlatformWeb); got != label {
		t.Fatalf("NO_COLOR platform label = %q, want %q", got, label)
	}
	if got := renderAppID(appID, clistyles.PanelMuted.Background(clistyles.ColorRowStripe)); got != appID {
		t.Fatalf("NO_COLOR App ID = %q, want %q", got, appID)
	}
	table := renderAppsTableAtWidth([]core.FirebaseApp{{DisplayName: "Web", Platform: core.AppPlatformWeb, AppID: appID, State: "ACTIVE"}}, 200, true)
	if strings.Contains(table, "\x1b[") {
		t.Fatalf("NO_COLOR table contains ANSI styling:\n%q", table)
	}
}

func TestRenderAppsColorsFitNarrowTable(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	apps := []core.FirebaseApp{{DisplayName: "A very long application name", Platform: core.AppPlatformAndroid, Namespace: "com.example.application", AppID: "1:123456789:android:abcdef", State: "ACTIVE"}}
	got := renderAppsTableAtWidth(apps, 56, true)
	if width := lipgloss.Width(got); width > 56 {
		t.Fatalf("colored narrow table width = %d, want <= 56\n%s", width, got)
	}
	plain := ansi.Strip(got)
	if !strings.Contains(plain, nerdFontAndroidGlyph+" android") || !strings.Contains(plain, "…") {
		t.Fatalf("colored narrow table lost platform or ellipsis:\n%s", plain)
	}
}

func TestRenderAppsUsesNerdFontPlatformGlyphsWhenEnabled(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	apps := []core.FirebaseApp{
		{DisplayName: "Android", Platform: core.AppPlatformAndroid, AppID: "android-id", State: "ACTIVE"},
		{DisplayName: "iOS", Platform: core.AppPlatformIOS, AppID: "ios-id", State: "ACTIVE"},
		{DisplayName: "Web", Platform: core.AppPlatformWeb, AppID: "web-id", State: "ACTIVE"},
	}
	got := renderAppsTableAtWidth(apps, 200, true)
	for _, expected := range []string{
		nerdFontAndroidGlyph + " android",
		nerdFontIOSGlyph + " ios",
		nerdFontWebGlyph + " web",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("table missing %q:\n%s", expected, got)
		}
	}

	details := renderAppDetails(core.FirebaseAppDetails{FirebaseApp: apps[1]}, true)
	if !strings.Contains(details, "Platform: "+nerdFontIOSGlyph+" ios") {
		t.Fatalf("details missing Nerd Font platform label:\n%s", details)
	}
}

func TestRenderAppsNerdFontGlyphsFitNarrowTable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	apps := []core.FirebaseApp{{DisplayName: "A very long application name", Platform: core.AppPlatformAndroid, Namespace: "com.example.application", AppID: "1:123456789:android:abcdef", State: "ACTIVE"}}
	got := renderAppsTableAtWidth(apps, 56, true)
	if !strings.Contains(got, nerdFontAndroidGlyph+" android") {
		t.Fatalf("narrow table omitted platform glyph:\n%s", got)
	}
	if width := lipgloss.Width(got); width > 56 {
		t.Fatalf("narrow table width = %d, want <= 56\n%s", width, got)
	}
}

func appStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}

func TestRenderAppListProjectColumnWidths(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	items := []appListItem{{Project: "A very long project display name", DisplayName: "Application", Platform: core.AppPlatformAndroid, Namespace: "com.example.app", AppID: "1:123:android:abcdef", State: "ACTIVE"}}
	natural := renderAppListTable(items, true, 250, false)
	if !strings.Contains(natural, "│ Project ") || !strings.Contains(natural, items[0].Project) || strings.Contains(natural, "…") || lipgloss.Width(natural) >= 250 {
		t.Fatalf("natural=%s", natural)
	}
	narrow := renderAppListTable(items, true, 60, false)
	if lipgloss.Width(narrow) > 60 || !strings.Contains(narrow, "…") || strings.Contains(narrow, "\x1b[") {
		t.Fatalf("narrow=%s", narrow)
	}
}

func TestRenderAppListColorsProjectColumnLikeGet(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	project := "Demo"
	got := renderAppListTable([]appListItem{{Project: project, DisplayName: "Android", Platform: core.AppPlatformAndroid, AppID: "1:2:android:3", State: "ACTIVE"}}, true, 200, false)
	want := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateBright).Render(project)
	if !strings.Contains(got, want) {
		t.Fatalf("project cell does not use get project color:\n%s", got)
	}
}

package styles

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
)

func TestRenderFirebaseAppPlatformsColorsMultipleAppIDs(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	value := "app.id in ['1:123:android:a1', '1:123:ios:b2', '1:123:web:c3']"
	base := lipgloss.NewStyle().Foreground(PaletteSlateDim).Background(ColorRowStripe)
	got := RenderFirebaseAppPlatforms(value, base)

	if plain := ansi.Strip(got); plain != value {
		t.Fatalf("rendered text = %q, want %q", plain, value)
	}
	for _, platform := range []core.AppPlatform{core.AppPlatformAndroid, core.AppPlatformIOS, core.AppPlatformWeb} {
		want := base.Foreground(FirebaseAppPlatformColor(platform)).Render(string(platform))
		if !strings.Contains(got, want) {
			t.Errorf("rendered expression missing styled %q: %q", platform, got)
		}
	}
}

func TestRenderFirebaseAppPlatformsRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	value := "app.id == '1:123:web:abc'"
	got := RenderFirebaseAppPlatforms(value, lipgloss.NewStyle().Foreground(FirebaseWeb))
	if got != value {
		t.Fatalf("NO_COLOR rendering = %q, want %q", got, value)
	}
}

func TestRenderFirebaseAppPlatformsUsesBaseStyleWithoutAppID(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	base := lipgloss.NewStyle().Foreground(PaletteSlateDim)
	if got, want := RenderFirebaseAppPlatforms("true", base), base.Render("true"); got != want {
		t.Fatalf("rendering without App ID = %q, want %q", got, want)
	}
}

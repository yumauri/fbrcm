package versions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/oauth2"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

func TestVersionsBlameCommandJSONSuccess(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoColor, "1")
	t.Setenv(env.GoogleAccessToken, "test-token")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	transport := blameRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "listVersions") {
			return blameJSONResponse(`{"versions":[{"versionNumber":"3","updateTime":"2026-09-06T14:32:10Z","updateUser":{"email":"alice@example.com"},"description":"Enable flag"},{"versionNumber":"2"},{"versionNumber":"1"}]}`, ""), nil
		}
		version := req.URL.Query().Get("versionNumber")
		configs := map[string]string{
			"3": `{"version":{"versionNumber":"3"},"parameters":{"flag":{"defaultValue":{"value":"new"},"valueType":"STRING"}}}`,
			"2": `{"version":{"versionNumber":"2"},"parameters":{"flag":{"defaultValue":{"value":"old"},"valueType":"STRING"}}}`,
			"1": `{"version":{"versionNumber":"1"}}`,
		}
		return blameJSONResponse(configs[version], `"etag-`+version+`"`), nil
	})
	base := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: transport})
	ctx := core.WithExecutionPolicy(base, core.StatelessExecutionPolicy())

	cmd := cliadapter.Command(newVersionsBlameCommandDefinition(svc))
	cmd.SetContext(ctx)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"demo", "flag", "--all", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result versionBlameResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Parameter != "flag" || result.AtVersion != "3" || !result.HistoryExhausted || len(result.Changes) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if result.Changes[0].PreviousVersion != "2" || result.Changes[0].Version != "3" || result.Changes[0].UpdateUser.Email != "alice@example.com" {
		t.Fatalf("newest change = %#v", result.Changes[0])
	}
	if result.Boundary == nil || result.Boundary.Version != "1" || result.Boundary.State != "absent" {
		t.Fatalf("boundary = %#v", result.Boundary)
	}
}

type blameRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn blameRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func blameJSONResponse(body, etag string) *http.Response {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	if etag != "" {
		header.Set("ETag", etag)
	}
	return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}

func TestRenderVersionBlameUsesVerticalLogAndSideBySideDiff(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	group := "checkout"
	history := core.ParameterHistory{
		Parameter: "checkout_enabled", AtVersion: "142", AtGroup: &group,
		ScannedVersionCount: 3,
		Changes: []core.ParameterHistoryChange{{
			PreviousVersion: "141",
			Version: firebase.RemoteConfigVersion{
				VersionNumber: "142", UpdateTime: "2026-09-06T14:32:10Z",
				UpdateUser: firebase.RemoteConfigUser{Name: "Alice Smith", Email: "alice@example.com"},
				ChangeNote: "Enable beta checkout",
			},
			Change: rcdiff.ParameterChange{
				Key: "checkout_enabled", Group: "checkout", Kind: rcdiff.ChangeChanged,
				Current: &firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
				Final:   &firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "true"}},
			},
		}},
	}

	output, err := renderVersionBlame(core.Project{Name: "Acme Production", ProjectID: "acme-prod"}, history, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Project: Acme Production (acme-prod)",
		"Parameter: checkout / checkout_enabled",
		"┆",
		"┝━ version 141 → 142",
		"│  Alice Smith <alice@example.com> on 2026-09-06",
		"│      Enable beta checkout",
		"│  Property: checkout / checkout_enabled",
		"│  value · default",
		"false",
		"true",
		"Showing 1 change after scanning 3 retained versions.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("renderVersionBlame() misses %q:\n%s", want, output)
		}
	}
}

func TestVersionBlameRailIsDarkerThanTransition(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	if got, want := blameRail("│"), lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDark).Render("│"); got != want {
		t.Fatalf("blameRail() = %q, want %q", got, want)
	}
	if got, want := blameMuted("version 141 → "), clistyles.PanelMuted.Render("version 141 → "); got != want {
		t.Fatalf("blameMuted() = %q, want %q", got, want)
	}
}

func TestVersionBlameHeadingUsesFullWidthGrayBackground(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	item := core.ParameterHistoryChange{
		PreviousVersion: "3776",
		Version:         firebase.RemoteConfigVersion{VersionNumber: "3777"},
	}

	const width = 48
	got := renderVersionBlameHeading(item, width)
	if gotWidth := lipgloss.Width(got); gotWidth != width {
		t.Fatalf("heading width = %d, want %d: %q", gotWidth, width, got)
	}

	background := clistyles.ColorInactiveSelection
	headingWidth := lipgloss.Width("┝━ version 3776 → 3777")
	want := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDark).Background(background).Render("┝━ ") +
		clistyles.PanelMuted.Background(background).Render("version 3776 → ") +
		lipgloss.NewStyle().Bold(true).Foreground(clistyles.PaletteSlateBright).Background(background).Render("3777") +
		lipgloss.NewStyle().Background(background).Render(strings.Repeat(" ", width-headingWidth))
	if got != want {
		t.Fatalf("heading = %q, want %q", got, want)
	}
}

func TestVersionBlameSeparatesEntriesWithTwoEmptyRailLines(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	parameter := &firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "true"}}
	history := core.ParameterHistory{
		Parameter: "flag",
		AtVersion: "3",
		Changes: []core.ParameterHistoryChange{
			{
				PreviousVersion: "2",
				Version:         firebase.RemoteConfigVersion{VersionNumber: "3"},
				Change:          rcdiff.ParameterChange{Key: "flag", Kind: rcdiff.ChangeChanged, Final: parameter},
			},
			{
				PreviousVersion: "1",
				Version:         firebase.RemoteConfigVersion{VersionNumber: "2"},
				Change:          rcdiff.ParameterChange{Key: "flag", Kind: rcdiff.ChangeAdded, Final: parameter},
			},
		},
	}

	got, err := renderVersionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, history, 80)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "\n┆\n┆\n┝━ version 1 → 2") {
		t.Fatalf("renderVersionBlame() does not contain two rail lines between entries:\n%s", got)
	}
}

func TestRenderVersionBlameReportsPresentHistoryBoundary(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	group := ""
	output, err := renderVersionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, core.ParameterHistory{
		Parameter: "flag", AtVersion: "2", AtGroup: &group, ScannedVersionCount: 2,
		HistoryExhausted: true, Boundary: &core.ParameterHistoryBoundary{Version: "1", Present: true},
	}, 80)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"No attributable change found.", `Parameter "flag" already exists in the oldest retained version, 1.`, "original author and creation time are unavailable"} {
		if !strings.Contains(output, want) {
			t.Fatalf("renderVersionBlame() misses %q:\n%s", want, output)
		}
	}
}

func TestRenderVersionBlameFitsNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	long := strings.Repeat("before ", 30)
	longer := strings.Repeat("after ", 30)
	output, err := renderVersionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, core.ParameterHistory{
		Parameter: "flag", AtVersion: "2", ScannedVersionCount: 2, HistoryExhausted: true,
		Boundary: &core.ParameterHistoryBoundary{Version: "1"},
		Changes: []core.ParameterHistoryChange{{
			PreviousVersion: "1",
			Version:         firebase.RemoteConfigVersion{VersionNumber: "2", UpdateTime: "2026-09-06T14:32:10Z", UpdateUser: firebase.RemoteConfigUser{Email: "a-very-long-address@example.com"}, ChangeNote: strings.Repeat("long note ", 20)},
			Change: rcdiff.ParameterChange{Key: "flag", Kind: rcdiff.ChangeChanged,
				Current: &firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: long}},
				Final:   &firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: longer}}},
		}},
	}, 44)
	if err != nil {
		t.Fatal(err)
	}
	for index, line := range strings.Split(output, "\n") {
		if got := lipgloss.Width(ansi.Strip(line)); got > 44 {
			t.Fatalf("line %d width = %d, want <= 44: %q", index, got, line)
		}
	}
}

func TestVersionBlameJSONPreservesMetadataAndBoundary(t *testing.T) {
	root := ""
	result := versionBlameJSON(core.Project{Name: "Demo", ProjectID: "demo"}, core.ParameterHistory{
		Parameter: "flag", AtVersion: "3", AtGroup: &root, ScannedVersionCount: 3, HistoryExhausted: true,
		Boundary: &core.ParameterHistoryBoundary{Version: "1", Present: true},
		Changes: []core.ParameterHistoryChange{{PreviousVersion: "2", Version: firebase.RemoteConfigVersion{
			VersionNumber: "3", UpdateTime: "2026-09-06T14:32:10Z", UpdateUser: firebase.RemoteConfigUser{Email: "alice@example.com"},
			ChangeNote: "note", RollbackSource: "1",
		}, Change: rcdiff.ParameterChange{Key: "flag", Kind: rcdiff.ChangeRemoved}}},
	})
	if result.AtGroup == nil || *result.AtGroup != "" || result.Boundary == nil || result.Boundary.State != "present" || result.Boundary.Group == nil {
		t.Fatalf("result = %#v", result)
	}
	if len(result.Changes) != 1 || result.Changes[0].ChangeNote == nil || result.Changes[0].RollbackSource == nil || result.Changes[0].Change.Kind != rcdiff.ChangeRemoved {
		t.Fatalf("changes = %#v", result.Changes)
	}
}

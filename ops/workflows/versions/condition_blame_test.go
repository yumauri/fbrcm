package versions

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestVersionsBlameRejectsInvalidEntitySelection(t *testing.T) {
	for _, args := range [][]string{
		{"demo"},
		{"demo", "flag", "--condition", "beta"},
		{"demo", "--condition="},
		{"demo", "--condition", " \t\n"},
		{"demo", "--condition", "beta", "--limit", "0"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := New(nil)
			cmd.SetArgs(append([]string{"blame"}, args...))
			if _, ok := errors.AsType[*shared.ArgumentError](cmd.Execute()); !ok {
				t.Fatal("expected a typed argument error before project resolution")
			}
		})
	}
}

func TestRenderConditionBlameShowsPublicationAndDefinitionDiff(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	history := core.ConditionHistory{
		Condition: "Beta Users", AtVersion: "3", ScannedVersionCount: 3, HistoryExhausted: true,
		Boundary: &core.ConditionHistoryBoundary{Version: "1"},
		Changes: []core.ConditionHistoryChange{
			{PreviousVersion: "2", Version: firebase.RemoteConfigVersion{
				VersionNumber: "3", UpdateTime: "2026-09-06T14:32:10Z",
				UpdateUser: firebase.RemoteConfigUser{Email: "alice@example.com"}, ChangeNote: "Expand beta",
			}, Change: rcdiff.ConditionChange{
				Name: "Beta Users", Kind: rcdiff.ChangeChanged, PreviousPosition: 1, FinalPosition: 2,
				Current: &firebase.RemoteConfigCondition{Name: "Beta Users", Expression: "true", TagColor: "BLUE"},
				Final:   &firebase.RemoteConfigCondition{Name: "Beta Users", Expression: "false", TagColor: "GREEN"},
			}},
			{PreviousVersion: "1", Version: firebase.RemoteConfigVersion{VersionNumber: "2"}, Change: rcdiff.ConditionChange{
				Name: "Beta Users", Kind: rcdiff.ChangeAdded, FinalPosition: 1,
				Final: &firebase.RemoteConfigCondition{Name: "Beta Users", Expression: "true", TagColor: "BLUE"},
			}},
		},
	}
	output, err := renderConditionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, history, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Condition: Beta Users", "version 2 → 3", "alice@example.com", "2026-09-06", "Expand beta", "expression", "position", "BLUE", "GREEN", "\n┆\n┆\n┝━ version 1 → 2"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "\x1b") {
		t.Fatalf("NO_COLOR output includes styling: %q", output)
	}
}

func TestRenderConditionBlameReportsUnknownOrigin(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	output, err := renderConditionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, core.ConditionHistory{
		Condition: "beta", AtVersion: "2", ScannedVersionCount: 2, HistoryExhausted: true,
		Boundary: &core.ConditionHistoryBoundary{Version: "1", Present: true},
	}, 80)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"No attributable change found.", `Condition "beta" already exists in the oldest retained version, 1.`, "original author and creation time are unavailable"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q:\n%s", want, output)
		}
	}
}

func TestRenderConditionBlameFitsNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	output, err := renderConditionBlame(core.Project{Name: "Demo", ProjectID: "demo"}, core.ConditionHistory{
		Condition: "beta", AtVersion: "2", ScannedVersionCount: 2, HistoryExhausted: true,
		Boundary: &core.ConditionHistoryBoundary{Version: "1"},
		Changes: []core.ConditionHistoryChange{{PreviousVersion: "1", Version: firebase.RemoteConfigVersion{
			VersionNumber: "2", UpdateUser: firebase.RemoteConfigUser{Email: "long-address@example.com"},
			UpdateTime: "2026-09-06T14:32:10Z", ChangeNote: strings.Repeat("long note ", 20),
		}, Change: rcdiff.ConditionChange{
			Name: "beta", Kind: rcdiff.ChangeChanged, PreviousPosition: 1, FinalPosition: 2,
			Current: &firebase.RemoteConfigCondition{Name: "beta", Expression: strings.Repeat("true && ", 30) + "true"},
			Final:   &firebase.RemoteConfigCondition{Name: "beta", Expression: strings.Repeat("false || ", 30) + "false"},
		}}},
	}, 44)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(output, "\n") {
		if ansi.StringWidth(line) > 44 {
			t.Fatalf("line %d exceeds width: %q", i, line)
		}
	}
}

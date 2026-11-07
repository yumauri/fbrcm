package explain

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

func TestConditionFlowConnectsFailuresUntilSelectionOrUncertainty(t *testing.T) {
	for _, tc := range []struct {
		name       string
		states     []evaluate.State
		winner     int
		unresolved bool
		arrows     int
	}{
		{"selected condition", []evaluate.State{evaluate.False, evaluate.False, evaluate.True, evaluate.False, evaluate.True}, 2, false, 2},
		{"first condition selected", []evaluate.State{evaluate.True, evaluate.False}, 0, false, 0},
		{"default selected", []evaluate.State{evaluate.False, evaluate.False}, -1, false, 2},
		{"no conditions", nil, -1, false, 0},
		{"first condition unknown", []evaluate.State{evaluate.Unknown, evaluate.False, evaluate.True}, 2, true, 0},
		{"failure then unknown", []evaluate.State{evaluate.False, evaluate.Unknown, evaluate.True}, 2, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := report{Project: "demo", Parameter: "flag", Winner: tc.winner, Unresolved: tc.unresolved,
				Param: firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "fallback"}}}
			for i, state := range tc.states {
				facts := map[string]evaluate.Fact{}
				if state != evaluate.Unknown {
					value := "different"
					if state == evaluate.True {
						value = "expected"
					}
					facts["app.customSignal['表示名']"] = evaluate.Fact{Known: true, Value: value, Source: "request"}
				}
				result, err := evaluate.Evaluate(`app.customSignal['表示名'].exactlyMatches(['expected'])`, &evaluate.Context{Facts: facts})
				if err != nil || result.State != state {
					t.Fatalf("fixture evaluation: %v, %v", result.State, err)
				}
				r.Conditions = append(r.Conditions, conditionResult{
					Entry: core.ConditionEntry{Priority: i + 1, Name: fmt.Sprintf("condition_%d_条件", i)},
					Value: firebase.RemoteConfigValue{Value: "value"}, Result: result,
				})
			}
			for _, noColor := range []string{"", "1"} {
				t.Setenv("NO_COLOR", noColor)
				for _, width := range []int{200, 48, 12, 4} {
					got := render(r, width)
					plain := ansi.Strip(got)
					if noColor != "" && plain != got {
						t.Fatal("NO_COLOR flow contains styling")
					}
					lines := strings.Split(plain, "\n")
					for _, line := range lines {
						if lipgloss.Width(line) > width {
							t.Fatalf("flow exceeds %d columns: %q", width, line)
						}
					}
					if width != 200 {
						continue
					}
					wantArrows := tc.arrows + 1 // Entry arrow before the first condition or the default.
					if arrows := strings.Count(plain, "\n↓ "); arrows != wantArrows {
						t.Fatalf("got %d flow arrows, want %d:\n%s", arrows, wantArrows, plain)
					}
					for i, condition := range r.Conditions {
						header := -1
						for j, line := range lines {
							if strings.Contains(line, "] "+condition.Entry.Name+" → ") {
								header = j
								break
							}
						}
						if header < 1 {
							t.Fatalf("missing condition header:\n%s", plain)
						}
						wantDivider := strings.Repeat("─", width)
						if i <= tc.arrows {
							wantDivider = "↓ " + strings.Repeat("─", width-2)
						}
						if lines[header-1] != wantDivider {
							t.Fatalf("flow does not cross divider before %s:\n%s", condition.Entry.Name, plain)
						}
						if i == 0 && (header < 2 || lines[header-2] != "") {
							t.Fatalf("entry arrow has a dangling flow line before it:\n%s", plain)
						}
						for j := header + 1; j < len(lines) && lines[j] != strings.Repeat("─", width) && lines[j] != "↓ "+strings.Repeat("─", width-2); j++ {
							if strings.HasPrefix(lines[j], "┆") != (i < tc.arrows) {
								t.Fatalf("flow gutter interrupted or continued past selection: %q\n%s", lines[j], plain)
							}
						}
					}
					if tc.arrows > 0 && noColor == "" {
						if !strings.Contains(got, clistyles.PanelMuted.Render("┆")) || !strings.Contains(got, clistyles.PanelMuted.Bold(true).Render("↓")+" "+clistyles.PanelMuted.Render(strings.Repeat("─", width-2))) {
							t.Fatal("flow and divider do not share the muted gray")
						}
					}
				}
			}
		})
	}
}

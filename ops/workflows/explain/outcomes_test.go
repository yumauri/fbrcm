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

func TestPossibleOutcomesUseRemainingRulesAndPriority(t *testing.T) {
	for _, tc := range []struct {
		name        string
		expressions []string
		values      []string
		winner      int
		want, omit  []string
	}{
		{
			name: "platform choices with a known prefix", winner: -1,
			expressions: []string{`app.customSignal['environment'].exactlyMatches(['sandbox']) && device.os == 'web'`, `device.os == 'ios'`},
			want:        []string{"├─ If device.os == 'web', then [10] candidate_0 would win → true", "├─ Otherwise, if device.os == 'ios', then [17] candidate_1 would win → true", "╰─ Otherwise: default → false"},
			omit:        []string{"environment", "unknown &&"},
		},
		{
			name: "all unknown conjuncts are required", winner: -1,
			expressions: []string{`device.os == 'web' && app.version >= '4.0'`},
			want:        []string{"If device.os == 'web' && app.version >= '4.0', then [10] candidate_0 would win → true"},
		},
		{
			name: "same remaining requirements cannot win twice", winner: -1,
			expressions: []string{`app.customSignal['environment'].exactlyMatches(['sandbox']) && device.os == 'web'`, `device.os == "web"`, `device.os == 'ios'`},
			values:      []string{"false", "true", "true"},
			want:        []string{"If device.os == 'web', then [10] candidate_0 would win → false", "Otherwise, if device.os == 'ios', then [24] candidate_2 would win → true", "Otherwise: default → false"},
			omit:        []string{"candidate_1", "[17]"},
		},
		{
			name: "stricter lower priority requirements are shadowed", winner: -1,
			expressions: []string{`device.os == 'web'`, `app.version >= '4.0' && device.os == 'web'`},
			want:        []string{"If device.os == 'web', then [10] candidate_0 would win", "Otherwise: default → false"},
			omit:        []string{"candidate_1", "app.version"},
		},
		{
			name: "weaker lower priority requirements remain reachable", winner: -1,
			expressions: []string{`device.os == 'web' && app.version >= '4.0'`, `device.os == 'web'`},
			want:        []string{"If device.os == 'web' && app.version >= '4.0', then [10] candidate_0 would win", "Otherwise, if device.os == 'web', then [17] candidate_1 would win"},
		},
		{
			name: "overlapping candidates preserve precedence", winner: -1,
			expressions: []string{`device.country in ['US']`, `device.language in ['en']`},
			want:        []string{"If device.country in ['US'], then [10] candidate_0 would win", "Otherwise, if device.language in ['en'], then [17] candidate_1 would win"},
		},
		{
			name: "false condition after an unknown rule cannot win", winner: -1,
			expressions: []string{`device.os == 'web' && false`, `app.version >= '4.0'`},
			want:        []string{"├─ If app.version >= '4.0', then [17] candidate_1 would win"},
			omit:        []string{"candidate_0", "device.os"},
		},
		{
			name: "known match blocks default and later candidates", winner: 1,
			expressions: []string{`device.os == 'web'`, "true", `app.version >= '4.0'`},
			want:        []string{"If device.os == 'web', then [10] candidate_0 would win", "╰─ Otherwise: [17] candidate_1 would win → true"},
			omit:        []string{"default", "candidate_2", "app.version"},
		},
		{
			name: "unsupported syntax stays opaque", winner: -1,
			expressions: []string{`future.deviceCategory == 'tablet'`},
			want:        []string{"If future.deviceCategory == 'tablet', then [10] candidate_0 would win → true (evaluation unavailable)", "Otherwise: default → false"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := report{Project: "demo", Parameter: "flag", Winner: tc.winner, Unresolved: true,
				Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}}}
			ctx := &evaluate.Context{Facts: map[string]evaluate.Fact{
				"app.customSignal['environment']": {Known: true, Value: "sandbox", Source: "request"},
			}}
			for i, expression := range tc.expressions {
				result, err := evaluate.Evaluate(expression, ctx)
				if err != nil {
					t.Fatal(err)
				}
				value := "true"
				if len(tc.values) > 0 {
					value = tc.values[i]
				}
				r.Conditions = append(r.Conditions, conditionResult{
					Entry: core.ConditionEntry{Priority: 10 + i*7, Name: fmt.Sprintf("candidate_%d", i), Expression: expression, TagColor: "PURPLE"},
					Value: firebase.RemoteConfigValue{Value: value}, Result: result,
				})
			}
			for _, noColor := range []string{"", "1"} {
				t.Setenv("NO_COLOR", noColor)
				for _, width := range []int{240, 48, 12} {
					got := render(r, width)
					if noColor != "" && ansi.Strip(got) != got {
						t.Fatal("NO_COLOR outcomes contain terminal styling")
					}
					for line := range strings.SplitSeq(got, "\n") {
						if lipgloss.Width(line) > width {
							t.Fatalf("outcomes exceed %d columns: %q", width, line)
						}
					}
					if width != 240 {
						continue
					}
					_, summary, ok := strings.Cut(ansi.Strip(got), "? flag = unknown\n")
					if !ok || !strings.HasPrefix(summary, "  ╰─ Possible outcomes (first matching branch wins)\n") {
						t.Fatalf("missing possible outcomes:\n%s", got)
					}
					for _, want := range tc.want {
						if !strings.Contains(summary, want) {
							t.Fatalf("missing %q:\n%s", want, summary)
						}
					}
					for _, omit := range tc.omit {
						if strings.Contains(summary, omit) {
							t.Fatalf("unreachable or already evaluated requirement %q:\n%s", omit, summary)
						}
					}
					if noColor == "" {
						_, coloredSummary, _ := strings.Cut(got, " = unknown\n")
						for _, candidate := range r.Conditions {
							if candidate.Result.State == evaluate.False {
								continue
							}
							if strings.Contains(summary, candidate.Entry.Name+" would win") && (!strings.Contains(coloredSummary, priorityText(candidate.Entry.Priority)) || !strings.Contains(coloredSummary, conditionName(candidate.Entry))) {
								t.Fatal("outcome priority or condition name lost its shared style")
							}
							if candidate.Result.State == evaluate.True {
								break
							}
						}
						if !strings.Contains(coloredSummary, lipgloss.NewStyle().Bold(true).Render("→")) {
							t.Fatal("outcome arrow is not bold")
						}
						if !strings.Contains(coloredSummary, clistyles.RemoteConfigValueStyle("true", "BOOLEAN").Render("true")) {
							t.Fatal("outcome value lost its shared type color")
						}
					}
				}
			}
		})
	}
}

func TestPossibleManagedOutcomesDoNotInventValues(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result, err := evaluate.Evaluate(`device.os == 'web'`, nil)
	if err != nil {
		t.Fatal(err)
	}
	managed := firebase.RemoteConfigValue{PersonalizationValue: []byte(`{"personalizationId":"p1"}`)}
	r := report{Project: "demo", Parameter: "flag", Winner: -1, Unresolved: true,
		Param: firebase.RemoteConfigParam{DefaultValue: &managed},
		Conditions: []conditionResult{{
			Entry: core.ConditionEntry{Priority: 10, Name: "web", Expression: `device.os == 'web'`},
			Value: managed, Result: result,
		}}}
	_, summary, ok := strings.Cut(render(r, 240), "? flag = unknown\n")
	if !ok || strings.Count(summary, "(requires Firebase-managed evaluation)") != 2 || strings.Contains(summary, "→ true") || strings.Contains(summary, "→ false") {
		t.Fatalf("managed outcomes imply an ordinary value:\n%s", summary)
	}
}

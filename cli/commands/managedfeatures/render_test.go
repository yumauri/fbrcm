package managedfeatures

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRenderExperimentsTableShowsBindingsAndListMetadataAtNaturalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	now := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)
	got := renderExperimentsTableAt([]core.ExperimentEntry{{
		Experiment: firebase.Experiment{
			Name:  "projects/123/namespaces/firebase/experiments/7",
			State: "RUNNING",
			Definition: firebase.ExperimentDefinition{
				DisplayName: "Passkey signup",
				Description: "Offer passkeys during signup",
				Variants: []firebase.ExperimentVariant{
					{Name: "Baseline", Weight: 1},
					{Name: "Variant A", Weight: 1},
				},
			},
			StartTime:      "2026-07-27T10:00:00Z",
			LastUpdateTime: "2026-07-28T23:19:00Z",
		},
		References: []core.ManagedValueReference{{
			Parameter: "signup_message", Condition: "android_beta", Percentage: new(0.0),
		}},
	}}, 300, now)
	for _, want := range []string{
		"ID", "Name", "Parameter", "Condition", "Exposure", "Last update", "State",
		"Passkey signup", "signup_message", "android_beta", "0%", "RUNNING", "41 minutes ago",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("experiments table missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"Description", "Offer passkeys during signup", "Variants", "Baseline", "Variant A"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("experiments table contains get-only field %q:\n%s", unwanted, got)
		}
	}
	assertTableHeaderOrder(t, got, "ID", "Name", "Parameter", "Condition", "Exposure", "Last update", "State")
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("NO_COLOR table contains ANSI escapes: %q", got)
	}
}

func TestRenderRolloutsTableUsesRequestedColumnsAndRelativeUpdate(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	now := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)
	got := renderRolloutsTableAt([]core.RolloutEntry{{
		Rollout: firebase.Rollout{
			Name:           "projects/123/namespaces/firebase/rollouts/funding",
			State:          "RUNNING",
			LastUpdateTime: "2026-07-28T10:00:00Z",
			Definition: firebase.RolloutDefinition{
				DisplayName: "Funding rollout",
				Description: "Do not show this description",
			},
		},
		References: []core.ManagedValueReference{{
			Parameter:  "funding_minimum_amount",
			Condition:  "android_beta",
			Value:      new("secret-value"),
			Percentage: new(10.0),
		}},
	}}, 300, now)

	for _, want := range []string{
		"funding", "Funding rollout", "funding_minimum_amount", "android_beta",
		"10%", "14 hours ago", "RUNNING",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rollouts table missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"Description", "Do not show this description", "Value", "secret-value"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("rollouts table contains unwanted field %q:\n%s", unwanted, got)
		}
	}
	assertTableHeaderOrder(t, got, "ID", "Name", "Parameter", "Condition", "Percentage", "Last update", "State")
}

func TestManagedFeatureTablesColorParametersAndConditionsLikeGet(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	rollouts := renderRolloutsTableAt([]core.RolloutEntry{{
		Rollout: firebase.Rollout{Name: "rollouts/funding"},
		References: []core.ManagedValueReference{{
			Parameter: "funding_minimum_amount", Condition: "beta",
			ConditionColor: "GREEN", Percentage: new(10.0),
		}},
	}}, 300, time.Now())
	personalizations := renderPersonalizationsTableAtWidth([]core.PersonalizationEntry{{
		ID: "provider",
		References: []core.ManagedValueReference{{
			Parameter: "kyc_provider", Condition: "staff", ConditionColor: "PURPLE",
		}},
	}}, 300)

	parameterPrefix := managedFeatureStylePrefix(lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright))
	greenPrefix := managedFeatureStylePrefix(lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor("GREEN")))
	purplePrefix := managedFeatureStylePrefix(lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor("PURPLE")))
	for name, table := range map[string]string{"rollouts": rollouts, "personalizations": personalizations} {
		if parameterPrefix == "" || !strings.Contains(table, parameterPrefix) {
			t.Fatalf("%s table does not use get parameter color:\n%s", name, table)
		}
	}
	if greenPrefix == "" || !strings.Contains(rollouts, greenPrefix) {
		t.Fatalf("rollouts table does not use condition tag color:\n%s", rollouts)
	}
	if purplePrefix == "" || !strings.Contains(personalizations, purplePrefix) {
		t.Fatalf("personalizations table does not use condition tag color:\n%s", personalizations)
	}
}

func TestManagedFeatureReferenceColorsRespectNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	reference := core.ManagedValueReference{
		Parameter: "parameter", Condition: "condition", ConditionColor: "GREEN", Percentage: new(10.0),
	}
	values := []string{
		renderRolloutsTableAt([]core.RolloutEntry{{
			Rollout:    firebase.Rollout{Name: "rollouts/example"},
			References: []core.ManagedValueReference{reference},
		}}, 300, time.Now()),
		renderPersonalizationsTableAtWidth([]core.PersonalizationEntry{{
			ID: "example", References: []core.ManagedValueReference{reference},
		}}, 300),
	}
	for index, value := range values {
		if strings.Contains(value, "\x1b[") {
			t.Fatalf("NO_COLOR table %d contains ANSI escapes: %q", index, value)
		}
	}
}

func TestFormatManagedFeatureRelativeTime(t *testing.T) {
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		value string
		want  string
	}{
		"empty":      {value: "", want: "—"},
		"sentinel":   {value: "1970-01-01T00:00:00Z", want: "—"},
		"invalid":    {value: "not-a-time", want: "not-a-time"},
		"seconds":    {value: "2026-07-29T11:59:31Z", want: "less than a minute ago"},
		"one minute": {value: "2026-07-29T11:59:00Z", want: "1 minute ago"},
		"minutes":    {value: "2026-07-29T11:19:00Z", want: "41 minutes ago"},
		"hours":      {value: "2026-07-28T22:00:00Z", want: "14 hours ago"},
		"days":       {value: "2026-07-27T12:00:00Z", want: "2 days ago"},
		"future":     {value: "2026-07-29T14:00:00Z", want: "in 2 hours"},
		"fractional": {value: "2026-07-29T11:59:00.123456789Z", want: "less than a minute ago"},
		"offset":     {value: "2026-07-29T13:00:00+02:00", want: "1 hour ago"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := formatManagedFeatureRelativeTime(test.value, now); got != test.want {
				t.Fatalf("formatManagedFeatureRelativeTime(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func assertTableHeaderOrder(t *testing.T, table string, headers ...string) {
	t.Helper()
	headerLine := strings.Split(ansi.Strip(table), "\n")[1]
	previous := -1
	for _, header := range headers {
		index := strings.Index(headerLine, header)
		if index < 0 || index <= previous {
			t.Fatalf("header %q is missing or out of order in %q", header, headerLine)
		}
		previous = index
	}
}

func managedFeatureStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, _ := strings.Cut(rendered, "x")
	return prefix
}

func TestManagedFeatureTablesRespectNarrowTerminalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	tests := map[string]struct {
		width int
		value string
	}{
		"experiments": {
			width: 80,
			value: renderExperimentsTableAtWidth([]core.ExperimentEntry{{
				Experiment: firebase.Experiment{
					Name: "projects/123/namespaces/firebase/experiments/123456",
					Definition: firebase.ExperimentDefinition{
						DisplayName: "An exceptionally long experiment display name that must be cropped",
					},
					State:          "RUNNING",
					StartTime:      "2026-07-27T09:10:11Z",
					LastUpdateTime: "2026-07-28T10:11:12Z",
				},
				References: []core.ManagedValueReference{{
					Parameter: "a_parameter_with_a_long_name", Condition: "a_condition_with_a_long_name",
				}},
			}}, 80),
		},
		"rollouts": {
			width: 100,
			value: renderRolloutsTableAtWidth([]core.RolloutEntry{{
				Rollout: firebase.Rollout{
					Name:       "projects/123/namespaces/firebase/rollouts/rollout_123",
					State:      "RUNNING",
					Definition: firebase.RolloutDefinition{DisplayName: "A long gradual funding rollout"},
				},
				References: []core.ManagedValueReference{{
					Parameter: "funding_minimum_amount_with_a_long_name", Condition: "android_beta_with_a_long_name",
					Value: new("20"), Percentage: new(10.0),
				}},
			}}, 100),
		},
		"personalizations": {
			width: 60,
			value: renderPersonalizationsTableAtWidth([]core.PersonalizationEntry{{
				ID: "personalization_with_a_very_long_identifier",
				References: []core.ManagedValueReference{{
					Group: "onboarding", Parameter: "identity_verification_provider", Condition: "android_beta_10",
				}},
			}}, 60),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(test.value, "…") {
				t.Fatalf("table was not cropped:\n%s", test.value)
			}
			for index, line := range strings.Split(ansi.Strip(test.value), "\n") {
				if width := lipgloss.Width(line); width > test.width {
					t.Fatalf("line %d width = %d, exceeds %d:\n%s", index, width, test.width, test.value)
				}
			}
		})
	}
}

func TestRenderManagedFeatureDetailsShowsKnownDataAndAPILimits(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	experiment := renderExperimentDetailsAtWidth(core.ExperimentEntry{
		Experiment: firebase.Experiment{
			Name:    "projects/123/namespaces/firebase/experiments/7",
			EndTime: "1970-01-01T00:00:00Z",
			Definition: firebase.ExperimentDefinition{
				DisplayName: "Signup",
				Variants:    []firebase.ExperimentVariant{{Name: "Baseline", Weight: 1}},
				Objectives: firebase.ExperimentObjectives{EventObjectives: []firebase.ExperimentEventObjective{{
					IsPrimary: true,
					CustomObjectiveDetails: &firebase.ExperimentCustomObjective{
						Event: "completed_signup", CountType: "ONCE",
					},
				}}},
			},
			State: "RUNNING",
		},
		References: []core.ManagedValueReference{{
			Group: "(root)", Parameter: "signup_message", Default: true, ValueType: "STRING",
			Percentage: new(0.0),
			Variants: []core.ManagedVariantValue{
				{ID: "control", Value: new("")},
				{ID: "enabled", NoChange: new(true)},
			},
		}},
	}, 300)
	for _, want := range []string{
		"ID: 7", "Ended: —", "Variants: 1 variant", "Baseline", "Objectives: 1 objective",
		"completed_signup · ONCE", "Bindings: 1 parameter value", "signup_message", "0%",
		`control=""`, "enabled=<no change>",
	} {
		if !strings.Contains(experiment, want) {
			t.Fatalf("experiment details missing %q:\n%s", want, experiment)
		}
	}

	personalization := renderPersonalizationDetails(core.PersonalizationEntry{
		ID: "personalization_1",
		References: []core.ManagedValueReference{{
			Group: "onboarding", Parameter: "kyc_provider", Condition: "android_beta_10", ValueType: "STRING",
		}},
	})
	for _, want := range []string{"Bindings: 1 parameter value", "kyc_provider", "does not expose personalization value candidates"} {
		if !strings.Contains(personalization, want) {
			t.Fatalf("personalization details missing %q:\n%s", want, personalization)
		}
	}
}

func TestManagedFeatureDetailTablesRespectNarrowTerminalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	values := map[string]struct {
		width int
		value string
	}{
		"experiment": {
			width: 64,
			value: renderExperimentDetailsAtWidth(core.ExperimentEntry{
				Experiment: firebase.Experiment{Definition: firebase.ExperimentDefinition{
					Variants: []firebase.ExperimentVariant{{Name: "A variant with a very long display name", Weight: 1}},
					Objectives: firebase.ExperimentObjectives{EventObjectives: []firebase.ExperimentEventObjective{{
						CustomObjectiveDetails: &firebase.ExperimentCustomObjective{Event: "an_extremely_long_analytics_event_name"},
					}}},
				}},
				References: []core.ManagedValueReference{{
					Group: "a_long_group", Parameter: "a_very_long_parameter_name", Condition: "a_very_long_condition",
					Variants: []core.ManagedVariantValue{{ID: "variant", Value: new("a very long value that must be cropped")}},
				}},
			}, 64),
		},
		"rollout": {
			width: 56,
			value: renderRolloutDetailsAtWidth(core.RolloutEntry{References: []core.ManagedValueReference{{
				Group: "a_long_group", Parameter: "a_very_long_parameter_name", Condition: "a_very_long_condition",
				Value: new("a very long rollout value"),
			}}}, 56),
		},
		"personalization": {
			width: 48,
			value: renderPersonalizationDetailsAtWidth(core.PersonalizationEntry{References: []core.ManagedValueReference{{
				Group: "a_long_group", Parameter: "a_very_long_parameter_name", Condition: "a_very_long_condition",
			}}}, 48),
		},
	}
	for name, test := range values {
		t.Run(name, func(t *testing.T) {
			for index, line := range strings.Split(ansi.Strip(test.value), "\n") {
				if !strings.ContainsAny(line, "┌│├└┬┼┴┐┤┘") {
					continue
				}
				if lipgloss.Width(line) > test.width {
					t.Fatalf("line %d exceeds %d columns:\n%s", index, test.width, test.value)
				}
			}
		})
	}
}

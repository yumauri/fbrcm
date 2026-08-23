package details

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestManagedFeatureDetailsRenderKnownAPIData(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tests := []struct {
		name string
		data *messages.ManagedFeatureViewData
		want []string
	}{
		{
			name: "experiment",
			data: &messages.ManagedFeatureViewData{
				Kind: messages.ManagedFeatureExperiment, Project: project,
				Experiment: &core.ExperimentEntry{
					Name:    "projects/123/namespaces/firebase/experiments/exp-1",
					State:   "RUNNING",
					EndTime: "1970-01-01T00:00:00Z",
					Definition: firebase.ExperimentDefinition{
						DisplayName: "Passkey signup",
						Variants:    []firebase.ExperimentVariant{{Name: "Baseline", Weight: 1}},
						Objectives: firebase.ExperimentObjectives{
							ActivationEvent: firebase.ExperimentActivationEvent{Event: "app_open"},
						},
					},
					References: []core.ManagedValueReference{{
						Group: "(root)", Parameter: "signup_message", Default: true, ValueType: "STRING",
						Percentage: new(0.0),
						Variants:   []core.ManagedVariantValue{{ID: "control", Value: new("")}},
					}},
				},
			},
			want: []string{
				"A/B Test", "exp-1", "Passkey signup", "RUNNING", "app_open", "Baseline (1)",
				"Ended", "—", "signup_message", "exposure 0%", `control = ""`,
			},
		},
		{
			name: "personalization",
			data: &messages.ManagedFeatureViewData{
				Kind: messages.ManagedFeaturePersonalization, Project: project,
				Template: core.ManagedFeatureTemplate{Version: "12", Source: "cache"},
				Personalization: &core.PersonalizationEntry{
					ID: "kyc-provider",
					References: []core.ManagedValueReference{{
						Group: "onboarding", Parameter: "kyc_provider", Condition: "android_beta", ValueType: "STRING",
					}},
				},
			},
			want: []string{"Personalization", "kyc-provider", "Remote Config version", "12", "onboarding / kyc_provider", "does not expose personalization"},
		},
		{
			name: "rollout",
			data: &messages.ManagedFeatureViewData{
				Kind: messages.ManagedFeatureRollout, Project: project,
				Template: core.ManagedFeatureTemplate{Version: "13", Source: "remote"},
				Rollout: &core.RolloutEntry{
					Name:       "projects/123/namespaces/firebase/rollouts/rollout-1",
					State:      "RUNNING",
					Definition: firebase.RolloutDefinition{DisplayName: "Funding rollout"},
					References: []core.ManagedValueReference{{
						Group: "onboarding", Parameter: "funding_minimum", Default: true,
						ValueType: "NUMBER", Value: new("20"), Percentage: new(10.0),
					}},
				},
			},
			want: []string{"Rollout", "rollout-1", "Funding rollout", "Remote Config version", "13", "Default value", "20", "10%"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := ansi.Strip(New().SetBounds(0, 0, 90, 80).SetActive(true).SetManagedFeatureData(test.data).View())
			for _, want := range test.want {
				if !strings.Contains(view, want) {
					t.Fatalf("%s details missing %q:\n%s", test.name, want, view)
				}
			}
		})
	}
}

func TestManagedFeatureDetailsColorBindingIdentity(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	reference := core.ManagedValueReference{
		Group:          "onboarding",
		Parameter:      "kyc_provider",
		Condition:      "android_beta",
		ConditionColor: "GREEN",
		ValueType:      "STRING",
	}

	for name, lines := range map[string][]string{
		"experiment":      appendExperimentBindings(nil, []core.ManagedValueReference{reference}),
		"personalization": appendBindings(nil, []core.ManagedValueReference{reference}, false),
	} {
		rendered := strings.Join(lines, "\n")
		for field, want := range map[string]string{
			"group":     styles.ParameterGroup.Render("onboarding"),
			"parameter": styles.ParameterName.Render("kyc_provider"),
			"condition": styles.DetailsConditionValueStyle("GREEN").Render("android_beta"),
		} {
			if !strings.Contains(rendered, want) {
				t.Fatalf("%s binding does not use the %s style:\n%s", name, field, rendered)
			}
		}
	}
}

func TestManagedFeatureDetailsColorState(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	tests := []struct {
		name  string
		state string
		lines []string
	}{
		{
			name:  "experiment",
			state: "RUNNING",
			lines: appendExperimentDetails(nil, 80, core.ManagedFeatureTemplate{}, &core.ExperimentEntry{
				State: "RUNNING",
			}),
		},
		{
			name:  "rollout",
			state: "PENDING",
			lines: appendRolloutDetails(nil, 80, core.ManagedFeatureTemplate{}, &core.RolloutEntry{
				State: "PENDING",
			}),
		},
	}

	for _, test := range tests {
		rendered := strings.Join(test.lines, "\n")
		want := renderedStylePrefix(styles.ManagedFeatureStatusStyle(test.state)) + test.state
		if !strings.Contains(rendered, want) {
			t.Fatalf("%s Details does not color state %q:\n%s", test.name, test.state, rendered)
		}
	}
}

func TestManagedFeatureDetailsFormatAllDatesInLocalTime(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.FixedZone("UTC+2", 2*60*60)
	t.Cleanup(func() {
		time.Local = previousLocal
	})

	experimentLines := appendExperimentDetails(nil, 80, core.ManagedFeatureTemplate{}, &core.ExperimentEntry{
		StartTime:      "2026-07-28T07:25:18.123Z",
		EndTime:        "2026-07-28T08:26:19Z",
		LastUpdateTime: "2026-07-28T09:27:20Z",
	})
	rolloutLines := appendRolloutDetails(nil, 80, core.ManagedFeatureTemplate{}, &core.RolloutEntry{
		CreateTime:     "2026-07-29T01:02:03Z",
		StartTime:      "2026-07-29T02:03:04Z",
		EndTime:        "2026-07-29T03:04:05Z",
		LastUpdateTime: "2026-07-29T04:05:06Z",
	})

	rendered := ansi.Strip(strings.Join(append(experimentLines, rolloutLines...), "\n"))
	for _, want := range []string{
		"2026-07-28 09:25:18",
		"2026-07-28 10:26:19",
		"2026-07-28 11:27:20",
		"2026-07-29 03:02:03",
		"2026-07-29 04:03:04",
		"2026-07-29 05:04:05",
		"2026-07-29 06:05:06",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("managed-feature Details missing local timestamp %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "T07:25:18") || strings.Contains(rendered, ".123Z") {
		t.Fatalf("managed-feature Details still contains raw RFC3339 timestamps:\n%s", rendered)
	}
}

func TestManagedFeatureTimeKeepsUnavailableAndUnexpectedValuesReadable(t *testing.T) {
	for name, test := range map[string]struct {
		value string
		want  string
	}{
		"empty":      {want: "—"},
		"epoch":      {value: "1970-01-01T00:00:00Z", want: "—"},
		"unexpected": {value: "not-a-time", want: "not-a-time"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := managedFeatureTime(test.value); got != test.want {
				t.Fatalf("managedFeatureTime(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

package managedfeatures

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestManagedFeatureCommandStructures(t *testing.T) {
	experiments := NewExperiments(nil)
	cmdtest.AssertSubcommands(t, experiments, "delete", "list", "show")
	cmdtest.AssertNestedFlag(t, experiments, []string{"list"}, "json")
	cmdtest.AssertNestedFlag(t, experiments, []string{"list"}, "filter")
	cmdtest.AssertNestedFlag(t, experiments, []string{"list"}, "update")
	listCmd := cmdtest.FindCommand(t, experiments, "list")
	if got := listCmd.Flags().Lookup("filter").Shorthand; got != "f" {
		t.Fatalf("experiments list --filter shorthand = %q, want f", got)
	}
	cmdtest.AssertNestedFlag(t, experiments, []string{"show"}, "json")
	cmdtest.AssertNestedFlag(t, experiments, []string{"show"}, "update")
	assertDeleteYesFlag(t, experiments)

	rollouts := NewRollouts(nil)
	cmdtest.AssertSubcommands(t, rollouts, "delete", "list", "show")
	assertDeleteYesFlag(t, rollouts)
	for _, child := range []string{"list", "show"} {
		cmdtest.AssertNestedFlag(t, rollouts, []string{child}, "update")
		cmdtest.AssertNestedFlag(t, rollouts, []string{child}, "json")
	}

	personalizations := NewPersonalizations(nil)
	cmdtest.AssertSubcommands(t, personalizations, "list", "show")
	for _, child := range []string{"list", "show"} {
		cmdtest.AssertNestedFlag(t, personalizations, []string{child}, "update")
		cmdtest.AssertNestedFlag(t, personalizations, []string{child}, "json")
	}
}

func TestFilterExperimentsByDisplayNameOnly(t *testing.T) {
	experiments := []core.ExperimentEntry{
		{
			Name: "projects/123/namespaces/firebase/experiments/passkey",
			Definition: firebase.ExperimentDefinition{
				DisplayName: "Passkey signup",
				Description: "Funds onboarding",
			},
		},
		{
			Name: "projects/123/namespaces/firebase/experiments/funding",
			Definition: firebase.ExperimentDefinition{
				DisplayName: "Funding amount",
				Description: "Passkey description must not match",
			},
		},
	}

	filtered := filterExperimentsByName(experiments, []string{"^passkey"})
	if len(filtered) != 1 || firebase.ManagedFeatureID(filtered[0].Name) != "passkey" {
		t.Fatalf("display-name filter = %#v, want passkey experiment", filtered)
	}
	filtered = filterExperimentsByName(experiments, []string{"=Passkey description must not match"})
	if len(filtered) != 0 {
		t.Fatalf("description-only filter returned %#v", filtered)
	}
}

func TestManagedFeatureDeleteConfirmation(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("n\n"))
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	confirmed, err := confirmManagedFeatureDelete(cmd, false, `Delete experiment "Signup" (7) from demo?`)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed {
		t.Fatal("confirmation accepted No input")
	}
	confirmed, err = confirmManagedFeatureDelete(cmd, true, "must not prompt")
	if err != nil || !confirmed {
		t.Fatalf("--yes confirmation = %t, %v", confirmed, err)
	}
}

func TestManagedFeatureIdentity(t *testing.T) {
	if got := managedFeatureIdentity("Signup", "7"); got != `"Signup" (7)` {
		t.Fatalf("identity = %q", got)
	}
	if got := managedFeatureIdentity("", "7"); got != "7" {
		t.Fatalf("ID-only identity = %q", got)
	}
}

func TestResolveManagedFeatureProjectRejectsTemplatePrefixes(t *testing.T) {
	cmd := &cobra.Command{}
	tests := []struct {
		query string
		want  string
	}{
		{
			query: "server@northstar",
			want:  "managed features support only the client Remote Config namespace; omit the server@ prefix",
		},
		{
			query: "client@northstar",
			want:  "managed-feature commands are project-scoped; omit the client@ prefix",
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			_, err := resolveProject(cmd, nil, test.query)
			if err == nil || err.Error() != test.want {
				t.Fatalf("resolveProject(%q) error = %v, want %q", test.query, err, test.want)
			}
		})
	}
}

func assertDeleteYesFlag(t *testing.T, root *cobra.Command) {
	t.Helper()
	cmdtest.AssertNestedFlag(t, root, []string{"delete"}, "yes")
	deleteCmd := cmdtest.FindCommand(t, root, "delete")
	if got := deleteCmd.Flags().Lookup("yes").Shorthand; got != "y" {
		t.Fatalf("%s delete --yes shorthand = %q, want y", root.Name(), got)
	}
}

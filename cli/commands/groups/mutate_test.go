package groups

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestReadDescriptionEditTreatsExplicitFalseAsAbsent(t *testing.T) {
	cmd := &cobra.Command{Use: "edit"}
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("no-description", false, "")
	if err := cmd.ParseFlags([]string{"--no-description=false"}); err != nil {
		t.Fatal(err)
	}
	_, err := readDescriptionEdit(cmd)
	var argument *shared.ArgumentError
	if !errors.As(err, &argument) {
		t.Fatalf("readDescriptionEdit error = %v, want ArgumentError", err)
	}
}

func TestReadDescriptionEditUsesOnlyTruthyRemovalFlag(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want string
	}{
		{name: "description with explicit false", argv: []string{"--description", "kept", "--no-description=false"}, want: "kept"},
		{name: "remove description", argv: []string{"--no-description"}, want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "edit"}
			cmd.Flags().String("description", "", "")
			cmd.Flags().Bool("no-description", false, "")
			if err := cmd.ParseFlags(test.argv); err != nil {
				t.Fatal(err)
			}
			got, err := readDescriptionEdit(cmd)
			if err != nil || got != test.want {
				t.Fatalf("readDescriptionEdit = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestReadDescriptionEditRejectsTruthyRemovalWithDescription(t *testing.T) {
	cmd := &cobra.Command{Use: "edit"}
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("no-description", false, "")
	if err := cmd.ParseFlags([]string{"--description", "kept", "--no-description"}); err != nil {
		t.Fatal(err)
	}
	_, err := readDescriptionEdit(cmd)
	var argument *shared.ArgumentError
	if !errors.As(err, &argument) {
		t.Fatalf("readDescriptionEdit error = %v, want ArgumentError", err)
	}
}

func TestNamedGroupMutationSkipsProjectsWithoutTarget(t *testing.T) {
	called := false
	mutation := namedGroupMutation("shared", func(*firebase.RemoteConfig, string) (bool, error) {
		called = true
		return true, nil
	})
	result, err := mutation(&firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"other": {}}})
	if err != nil || result.matched || result.applicable || called {
		t.Fatalf("mutation = result %+v, called %v, err %v; want skipped", result, called, err)
	}
}

func TestNamedGroupMutationRequiresExactCaseSensitiveName(t *testing.T) {
	var resolved string
	mutation := namedGroupMutation("SHARED", func(_ *firebase.RemoteConfig, name string) (bool, error) {
		resolved = name
		return true, nil
	})
	result, err := mutation(&firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"Shared": {}}})
	if err != nil || result.matched || result.applicable || resolved != "" {
		t.Fatalf("mutation = result %+v, resolved %q, err %v; want skipped", result, resolved, err)
	}
}

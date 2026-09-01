package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/cli/shared/planflag"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

const (
	projectFilterFlagHelp   = "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated"
	targetFilterFlagHelp    = "Filter template targets by optional client@ or server@ project query; may be repeated"
	parameterFilterFlagHelp = "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated"
	parameterSearchFlagHelp = "Search parameters by name, description, values, and conditions"
	dryRunFlagHelp          = "Preview the requested mutation without applying it"
	planOutFlagHelp         = "Write an immutable validated publication plan instead of applying the mutation"
)

func AddProjectFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("project", "p", nil, projectFilterFlagHelp)
}

func AddProjectTargetFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("project", "p", nil, targetFilterFlagHelp)
}

func AddProjectListFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("filter", "f", nil, projectFilterFlagHelp)
}

func AddParameterFilterFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("filter", "f", nil, parameterFilterFlagHelp)
	cmd.Flags().String("search", "", parameterSearchFlagHelp)
}

func AddDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, dryRunFlagHelp)
}

// AddPlanOutFlag enables immutable publication-plan output for one mutation
// command. The plan mode is intentionally distinct from draft and dry-run.
func AddPlanOutFlag(cmd *cobra.Command) {
	cmd.Flags().String("plan-out", "", planOutFlagHelp)
	for _, name := range []string{"dry-run", "draft", "yes"} {
		if cmd.Flags().Lookup(name) != nil {
			cmd.MarkFlagsMutuallyExclusive("plan-out", name)
		}
	}
}

func PlanOutputPath(cmd *cobra.Command) (string, bool, error) {
	return planflag.OutputPath(cmd)
}

func AddChangeNoteFlag(cmd *cobra.Command) {
	cmd.Flags().String("change-note", "", "Change note for publication or the stored draft")
}

func ReadChangeNoteFlag(cmd *cobra.Command) (*string, error) {
	if cmd.Flags().Lookup("change-note") == nil || !cmd.Flags().Changed("change-note") {
		return nil, nil
	}
	value, err := cmd.Flags().GetString("change-note")
	if err != nil {
		return nil, err
	}
	value, err = firebase.NormalizeChangeNote(value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func WithChangeNote(ctx context.Context, changeNote *string) (context.Context, error) {
	if changeNote == nil {
		return ctx, nil
	}
	return firebase.WithChangeNote(ctx, *changeNote)
}

func AddYesFlag(cmd *cobra.Command, help string) {
	cmd.Flags().BoolP("yes", "y", false, help)
}

// ValidateNonBlankInputs rejects explicitly supplied whitespace-only argv
// values before command-specific parsing can reinterpret them as an omitted
// selector, default, or collection resource. A small set of content flags
// intentionally accepts the exact empty string, but even those reject a value
// containing only whitespace.
func ValidateNonBlankInputs(cmd *cobra.Command, args []string) error {
	for index, value := range args {
		if strings.TrimSpace(value) == "" {
			return InvalidArgument(fmt.Errorf("argument %d requires a non-empty value", index+1))
		}
	}

	allowExactEmpty := map[string]bool{
		"change-note": true,
		"description": true,
		"group":       true,
		"label":       true,
		"value":       true,
	}
	var validationErr error
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if validationErr != nil {
			return
		}
		values, ok, err := stringFlagValues(cmd, flag)
		if err != nil {
			validationErr = InvalidArgument(err)
			return
		}
		if !ok {
			return
		}
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				continue
			}
			if value == "" && allowExactEmpty[flag.Name] && flag.Value.Type() == "string" {
				continue
			}
			validationErr = InvalidArgument(fmt.Errorf("--%s requires a non-empty value", flag.Name))
			return
		}
	})
	return validationErr
}

func stringFlagValues(cmd *cobra.Command, flag *pflag.Flag) ([]string, bool, error) {
	switch flag.Value.Type() {
	case "string":
		value, err := cmd.Flags().GetString(flag.Name)
		return []string{value}, true, err
	case "stringArray":
		values, err := cmd.Flags().GetStringArray(flag.Name)
		return values, true, err
	case "stringSlice":
		values, err := cmd.Flags().GetStringSlice(flag.Name)
		return values, true, err
	default:
		return nil, false, nil
	}
}

// ValidateQueryFlags rejects explicitly supplied empty filter queries. Empty
// mode-prefixed queries would otherwise be dropped and could broaden a
// mutation to every item.
func ValidateQueryFlags(cmd *cobra.Command) error {
	for _, name := range []string{"filter", "project"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil || !cmd.Flags().Changed(name) {
			continue
		}
		values, err := cmd.Flags().GetStringArray(name)
		if err != nil {
			return InvalidArgument(err)
		}
		for _, raw := range values {
			querySource := raw
			if name == "project" {
				target, _, err := rctarget.ParseSelector(raw)
				if err != nil {
					return InvalidArgument(fmt.Errorf("--project: %w", err))
				}
				querySource = target.ProjectID
			}
			_, query := filter.ParseModePrefixedQuery(querySource)
			if strings.TrimSpace(query) == "" {
				return InvalidArgument(fmt.Errorf("--%s requires a non-empty query", name))
			}
		}
	}
	return nil
}

// RejectTemplateProjectFilters enforces project-scoped commands that do not
// accept client@ or server@ template selectors.
func RejectTemplateProjectFilters(values []string) error {
	for _, raw := range values {
		_, explicit, err := rctarget.ParseSelector(raw)
		if err != nil {
			return InvalidArgument(fmt.Errorf("--project: %w", err))
		}
		if explicit {
			return InvalidArgument(fmt.Errorf("--project %q uses template syntax, but this command is project-scoped", raw))
		}
	}
	return nil
}

// RejectChangedFlags rejects flags that are unavailable in a particular
// invocation mode, including explicitly supplied false or empty values.
func RejectChangedFlags(cmd *cobra.Command, mode string, names ...string) error {
	for _, name := range names {
		if cmd.Flags().Changed(name) {
			return InvalidArgument(fmt.Errorf("--%s is unavailable in %s", name, mode))
		}
	}
	return nil
}

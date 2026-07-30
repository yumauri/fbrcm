package shared

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/firebase"
)

const (
	projectFilterFlagHelp   = "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated"
	targetFilterFlagHelp    = "Filter template targets by optional client@ or server@ project query; may be repeated"
	parameterFilterFlagHelp = "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated"
	parameterSearchFlagHelp = "Search parameters by name, description, values, and conditions"
	dryRunFlagHelp          = "Preview changes without writing local or Firebase state"
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

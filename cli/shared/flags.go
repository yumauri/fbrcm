package shared

import "github.com/spf13/cobra"

const (
	projectFilterFlagHelp   = "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated"
	parameterFilterFlagHelp = "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated"
	parameterSearchFlagHelp = "Search parameters by name, description, values, and conditions"
	dryRunFlagHelp          = "Log Firebase write requests without sending them"
)

func AddProjectFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("project", "p", nil, projectFilterFlagHelp)
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

func AddYesFlag(cmd *cobra.Command, help string) {
	cmd.Flags().BoolP("yes", "y", false, help)
}

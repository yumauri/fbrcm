package updatecmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type valueSpec struct {
	value     string
	valueType string
}

type updateSpec struct {
	value                      *valueSpec
	name                       string
	group                      string
	description                string
	removeConditionalValues    []string
	nameChanged                bool
	groupChanged               bool
	descriptionChanged         bool
	removeAllConditionalValues bool
}

type updateTotals struct {
	modifiedProjects int
	updatedParams    int
}

type updateOptions struct {
	shared.ParameterMutationOpts
	spec updateSpec
}

// New constructs the update command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [parameter]",
		Short: "Update Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateCommand(cmd, svc, args)
		},
	}

	addUpdateFlags(cmd)
	return cmd
}

func addUpdateFlags(cmd *cobra.Command) {
	shared.AddProjectFilterFlag(cmd)
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and update without confirmation")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().Bool("no-group", false, "Move parameter out of its group")
	cmd.Flags().String("name", "", "New parameter name")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.Flags().Bool("remove-all-conditional-values", false, "Remove all conditional values from matched parameters")
	cmd.Flags().StringArray("remove-conditional-value", nil, "Remove a conditional value from matched parameters; may be repeated")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
	cmd.MarkFlagsMutuallyExclusive("group", "no-group")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditional-values", "remove-conditional-value")
}

func runUpdateCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readUpdateOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		if opts.Draft {
			return fmt.Errorf("--draft is unavailable in stdin mode")
		}
		corelog.For("update").Info("stdin mode enabled; using remote config from stdin")
		return runUpdateStdin(cmd, opts.ParamFilters, opts.ParamExpr, opts.Search, opts.spec)
	}
	return runUpdateRemote(cmd, svc, opts)
}

func readUpdateOptions(cmd *cobra.Command, args []string) (updateOptions, error) {
	mutationOpts, err := shared.ReadParameterMutationOpts(cmd, args)
	if err != nil {
		return updateOptions{}, err
	}
	spec, err := readUpdateSpec(cmd)
	if err != nil {
		return updateOptions{}, err
	}
	return updateOptions{
		ParameterMutationOpts: mutationOpts,
		spec:                  spec,
	}, nil
}

func readUpdateSpec(cmd *cobra.Command) (updateSpec, error) {
	groupName, err := cmd.Flags().GetString("group")
	if err != nil {
		return updateSpec{}, err
	}
	noGroup, err := cmd.Flags().GetBool("no-group")
	if err != nil {
		return updateSpec{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return updateSpec{}, err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return updateSpec{}, err
	}
	removeAllConditionalValues, err := cmd.Flags().GetBool("remove-all-conditional-values")
	if err != nil {
		return updateSpec{}, err
	}
	removeConditionalValues, err := readRemoveConditionalValues(cmd)
	if err != nil {
		return updateSpec{}, err
	}
	value, err := readValueSpec(cmd)
	if err != nil {
		return updateSpec{}, err
	}

	groupChanged := cmd.Flags().Changed("group")
	if noGroup {
		groupChanged = true
		groupName = ""
	}
	descriptionChanged := cmd.Flags().Changed("description")
	nameChanged := cmd.Flags().Changed("name")
	groupName = strings.TrimSpace(groupName)
	name = strings.TrimSpace(name)
	if nameChanged && name == "" {
		return updateSpec{}, fmt.Errorf("--name cannot be empty")
	}

	return updateSpec{
		value:                      value,
		name:                       name,
		group:                      groupName,
		description:                description,
		removeConditionalValues:    removeConditionalValues,
		nameChanged:                nameChanged,
		groupChanged:               groupChanged,
		descriptionChanged:         descriptionChanged,
		removeAllConditionalValues: removeAllConditionalValues,
	}, nil
}

func readRemoveConditionalValues(cmd *cobra.Command) ([]string, error) {
	values, err := cmd.Flags().GetStringArray("remove-conditional-value")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("--remove-conditional-value cannot be empty")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func readValueSpec(cmd *cobra.Command) (*valueSpec, error) {
	value, err := shared.ReadValueFlag(cmd, false)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return &valueSpec{value: value.Value, valueType: value.Type}, nil
}

package updatecmd

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

type valueSpec struct {
	value           string
	valueType       string
	useInAppDefault bool
}

type updateSpec struct {
	value                      *valueSpec
	condition                  string
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
func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "update [parameter]",
		Short: "Update Remote Config parameters",
		Args:  invocation.MaximumNArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			return runUpdateCommand(cmd, svc, args)
		},
	}

	addUpdateFlags(cmd)
	invocation.RegisterResponse(cmd, []sharedrc.RemoteMutationJSONResult{}, sharedrc.PlanCreatedResult{}, contract.ArtifactData{})
	return cmd
}

func addUpdateFlags(cmd invocation.FlagGroups) {
	shared.AddProjectTargetFilterFlag(cmd)
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and update without confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().Bool("no-group", false, "Move parameter out of its group")
	cmd.Flags().String("name", "", "New parameter name")
	cmd.Flags().String("type", "", "Parameter type: string, boolean, number, or json")
	cmd.Flags().String("value", "", "Parameter value interpreted according to --type")
	cmd.Flags().Bool("use-in-app-default", false, "Use the application's default value")
	cmd.Flags().String("condition", "", "Set the value for this condition instead of the default value")
	cmd.Flags().Bool("remove-all-conditional-values", false, "Remove all conditional values from matched parameters")
	cmd.Flags().StringArray("remove-conditional-value", nil, "Remove a conditional value from matched parameters; may be repeated")
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
	cmd.MarkFlagsMutuallyExclusive("value", "use-in-app-default")
	cmd.MarkFlagsMutuallyExclusive("group", "no-group")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditional-values", "remove-conditional-value")
	cmd.MarkFlagsMutuallyExclusive("condition", "remove-all-conditional-values", "remove-conditional-value")
}

func runUpdateCommand(cmd invocation.Call, svc *core.Core, args []string) error {
	opts, err := readUpdateOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		if err := shared.RejectChangedFlags(cmd, "stdin mode", "project", "dry-run", "draft", "change-note", "plan-out"); err != nil {
			return err
		}
		corelog.For("update").Info("stdin mode enabled; using remote config from stdin")
		return runUpdateStdin(cmd, opts.ParamFilters, opts.ParamArgument, opts.ParamExpr, opts.Search, opts.spec)
	}
	return runUpdateRemote(cmd, svc, opts)
}

func readUpdateOptions(cmd invocation.Call, args []string) (updateOptions, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) == "" {
		return updateOptions{}, shared.InvalidArgument(fmt.Errorf("parameter argument cannot be empty"))
	}
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

func readUpdateSpec(cmd invocation.Call) (updateSpec, error) {
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
	condition, err := cmd.Flags().GetString("condition")
	if err != nil {
		return updateSpec{}, err
	}
	conditionChanged := cmd.Flags().Changed("condition")
	condition = strings.TrimSpace(condition)
	if conditionChanged && condition == "" {
		return updateSpec{}, shared.InvalidArgument(fmt.Errorf("--condition cannot be empty"))
	}
	if conditionChanged && value == nil {
		return updateSpec{}, shared.InvalidArgument(fmt.Errorf("--condition requires --value or --use-in-app-default"))
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
		return updateSpec{}, shared.InvalidArgument(fmt.Errorf("--name cannot be empty"))
	}

	return updateSpec{
		value:                      value,
		condition:                  condition,
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

func readRemoveConditionalValues(cmd invocation.Call) ([]string, error) {
	values, err := cmd.Flags().GetStringArray("remove-conditional-value")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, shared.InvalidArgument(fmt.Errorf("--remove-conditional-value cannot be empty"))
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func readValueSpec(cmd invocation.Call) (*valueSpec, error) {
	value, err := shared.ReadValueFlag(cmd, false)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return &valueSpec{value: value.Value, valueType: value.Type, useInAppDefault: value.UseInAppDefault}, nil
}

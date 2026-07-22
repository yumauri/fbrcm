package duplicatecmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

type duplicateOptions struct {
	projectFilters []string
	projectExpr    string
	dryRun         bool
	draft          bool
	yes            bool
	source         string
	target         string
}

// New constructs the duplicate command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duplicate <source> <target>",
		Short: "Duplicate a Remote Config parameter",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := readDuplicateOptions(cmd, args)
			if err != nil {
				return err
			}
			return runDuplicateRemote(cmd, svc, opts)
		},
	}
	shared.AddProjectFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and duplicate without confirmation")
	return cmd
}

func readDuplicateOptions(cmd *cobra.Command, args []string) (duplicateOptions, error) {
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return duplicateOptions{}, err
	}
	projectExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return duplicateOptions{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return duplicateOptions{}, err
	}
	draft, err := cmd.Flags().GetBool("draft")
	if err != nil {
		return duplicateOptions{}, err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return duplicateOptions{}, err
	}
	source := strings.TrimSpace(args[0])
	target := strings.TrimSpace(args[1])
	if source == "" {
		return duplicateOptions{}, fmt.Errorf("source parameter key cannot be empty")
	}
	if target == "" {
		return duplicateOptions{}, fmt.Errorf("target parameter key cannot be empty")
	}
	if strings.EqualFold(source, target) {
		return duplicateOptions{}, fmt.Errorf("source and target parameter keys must differ")
	}
	return duplicateOptions{
		projectFilters: projectFilters,
		projectExpr:    projectExpr,
		dryRun:         dryRun,
		draft:          draft,
		yes:            yes,
		source:         source,
		target:         target,
	}, nil
}

func runDuplicateRemote(cmd *cobra.Command, svc *core.Core, opts duplicateOptions) error {
	ctx := context.Background()
	if opts.dryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, opts.projectFilters)
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, opts.projectExpr)
	if err != nil {
		return err
	}
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })

	plan := func(project core.Project, _ *sharedrc.ProjectConfig) (sharedrc.RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			changed, finalCfg, err := duplicateProject(cmd, project, current, opts.source, opts.target, opts.yes)
			if err != nil {
				return 0, nil, err
			}
			if !changed {
				return 0, finalCfg, nil
			}
			return 1, finalCfg, nil
		}, nil
	}
	var totals sharedrc.RemoteMutationTotals
	if opts.draft {
		totals, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, "duplicate", plan)
	} else {
		totals, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, "duplicate", "📋", plan)
	}
	corelog.For("duplicate").Info("total", "projects", totals.ModifiedProjects, "parameters", totals.ChangedParams)
	sharedrc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.draft], "📋")
	return err
}

func duplicateProject(cmd *cobra.Command, project core.Project, current *firebase.RemoteConfig, source, target string, yes bool) (bool, *firebase.RemoteConfig, error) {
	changed, finalCfg, sourceTarget, err := duplicateParameter(current, source, target)
	if err != nil || !changed {
		return changed, finalCfg, err
	}
	diffText, hasChanges := sharedrc.RenderRemoteConfigDiff(current, finalCfg)
	if !hasChanges {
		return false, finalCfg, nil
	}
	prompt := fmt.Sprintf(
		"Duplicate %s as %s in %s?",
		shared.FormatParameterHeader(sourceTarget.Key, sourceTarget.Group),
		target,
		project.ProjectID,
	)
	confirmed, err := shared.PrintDiffAndConfirm(cmd, yes, cmd.ErrOrStderr(), diffText, prompt, false)
	if err != nil || !confirmed {
		return false, finalCfg, err
	}
	return true, finalCfg, nil
}

func duplicateParameter(cfg *firebase.RemoteConfig, source, target string) (bool, *firebase.RemoteConfig, shared.ParamTarget, error) {
	finalCfg, err := firebase.CloneRemoteConfig(cfg)
	if err != nil {
		return false, nil, shared.ParamTarget{}, err
	}
	sourceTarget, found, err := resolveSource(finalCfg, source)
	if err != nil || !found {
		return false, finalCfg, sourceTarget, err
	}
	if paramExistsFold(finalCfg, target) {
		return false, finalCfg, sourceTarget, fmt.Errorf("target parameter %q already exists", target)
	}
	if err := draft.DuplicateParameterNamed(sourceTarget.Group, sourceTarget.Key, target)(finalCfg); err != nil {
		return false, finalCfg, sourceTarget, err
	}
	return true, finalCfg, sourceTarget, nil
}

func paramExistsFold(cfg *firebase.RemoteConfig, requested string) bool {
	for _, target := range shared.CollectParamTargets(cfg) {
		if strings.EqualFold(target.Key, requested) {
			return true
		}
	}
	return false
}

func resolveSource(cfg *firebase.RemoteConfig, requested string) (shared.ParamTarget, bool, error) {
	var matches []shared.ParamTarget
	for _, target := range shared.CollectParamTargets(cfg) {
		if strings.EqualFold(target.Key, requested) {
			matches = append(matches, target)
		}
	}
	if len(matches) == 0 {
		return shared.ParamTarget{}, false, nil
	}
	if len(matches) > 1 {
		return shared.ParamTarget{}, false, fmt.Errorf("source parameter %q is ambiguous across groups", requested)
	}
	return matches[0], true, nil
}

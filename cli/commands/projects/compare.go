package projects

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
)

type compareOptions struct {
	Groups         []string
	ParamFilters   []string
	Expr           string
	Search         shared.ParameterSearch
	JSON           bool
	Cached         bool
	ParametersOnly bool
	ConditionsOnly bool
	Prune          bool
	All            bool
	Interactive    bool
	DryRun         bool
	Yes            bool
}

func newDiffCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <source-project> <target-project>",
		Short: "Compare Remote Config between two projects",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := readCompareOptions(cmd)
			if err != nil {
				return err
			}
			return runProjectsDiff(cmd, svc, args[0], args[1], opts)
		},
	}
	addCompareSelectionFlags(cmd)
	cmd.Flags().Bool("cached", false, "Use cached Remote Config instead of live Firebase fetch")
	cmd.Flags().Bool("json", false, "Print diff as JSON")
	return cmd
}

func newPromoteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote <source-project> <target-project>",
		Short: "Promote selected Remote Config changes between projects",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := readCompareOptions(cmd)
			if err != nil {
				return err
			}
			return runProjectsPromote(cmd, svc, args[0], args[1], opts)
		},
	}
	addCompareSelectionFlags(cmd)
	cmd.Flags().Bool("interactive", false, "Review each promotion item interactively")
	cmd.Flags().Bool("all", false, "Select all eligible promotion items")
	cmd.Flags().Bool("prune", false, "Include target-only removals")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Skip final publish confirmation")
	cmd.Flags().Bool("json", false, "Print promotion result as JSON")
	return cmd
}

func addCompareSelectionFlags(cmd *cobra.Command) {
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().StringArray("group", nil, "Select parameters in group; may be repeated")
	cmd.Flags().String("expr", "", "Filter parameter changes by expr-lang expression")
	cmd.Flags().Bool("parameters", false, "Include only parameter and group description changes")
	cmd.Flags().Bool("conditions", false, "Include only condition changes")
}

func readCompareOptions(cmd *cobra.Command) (compareOptions, error) {
	var opts compareOptions
	var err error
	opts.Groups, err = cmd.Flags().GetStringArray("group")
	if err != nil {
		return opts, err
	}
	opts.ParamFilters, err = cmd.Flags().GetStringArray("filter")
	if err != nil {
		return opts, err
	}
	opts.Expr, err = cmd.Flags().GetString("expr")
	if err != nil {
		return opts, err
	}
	searchValue, err := cmd.Flags().GetString("search")
	if err != nil {
		return opts, err
	}
	opts.Search = shared.NewParameterSearch(searchValue)
	opts.JSON, err = cmd.Flags().GetBool("json")
	if err != nil {
		return opts, err
	}
	opts.ParametersOnly, err = cmd.Flags().GetBool("parameters")
	if err != nil {
		return opts, err
	}
	opts.ConditionsOnly, err = cmd.Flags().GetBool("conditions")
	if err != nil {
		return opts, err
	}
	if cmd.Flags().Lookup("cached") != nil {
		opts.Cached, err = cmd.Flags().GetBool("cached")
		if err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("prune") != nil {
		opts.Prune, err = cmd.Flags().GetBool("prune")
		if err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("all") != nil {
		opts.All, err = cmd.Flags().GetBool("all")
		if err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("interactive") != nil {
		opts.Interactive, err = cmd.Flags().GetBool("interactive")
		if err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("dry-run") != nil {
		opts.DryRun, err = cmd.Flags().GetBool("dry-run")
		if err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("yes") != nil {
		opts.Yes, err = cmd.Flags().GetBool("yes")
		if err != nil {
			return opts, err
		}
	}
	opts.Expr = strings.TrimSpace(opts.Expr)
	opts.Groups = normalizeGroups(opts.Groups)
	return opts, nil
}

func runProjectsDiff(cmd *cobra.Command, svc *core.Core, sourceQuery, targetQuery string, opts compareOptions) error {
	ctx := context.Background()
	source, target, sourceCfg, targetCfg, err := loadCompareConfigs(ctx, cmd, svc, sourceQuery, targetQuery, opts.Cached)
	if err != nil {
		return err
	}
	result := filterDiffResult(source, sourceCfg, target, targetCfg, rcdiff.CompareRemoteConfigs(targetCfg, sourceCfg), opts)
	if opts.JSON {
		return shared.WriteJSON(cmd, compareJSON(source, target, result))
	}
	if !result.HasChanges() {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
		return nil
	}
	text, _ := rcdiff.RenderResult(result)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), text)
	return nil
}

func runProjectsPromote(cmd *cobra.Command, svc *core.Core, sourceQuery, targetQuery string, opts compareOptions) error {
	ctx := context.Background()
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	source, target, sourceCfg, targetCfg, err := loadCompareConfigs(ctx, cmd, svc, sourceQuery, targetQuery, false)
	if err != nil {
		return err
	}

	plan := rcpromote.BuildPlan(sourceCfg, targetCfg, rcpromote.Options{Prune: opts.Prune})
	plan.Diff = filterDiffResult(source, sourceCfg, target, targetCfg, plan.Diff, opts)
	plan.Items = filterPromotionItems(plan.Items, plan.Diff, opts)
	if len(plan.Items) == 0 {
		return writePromoteNoChanges(cmd, source, target, opts)
	}

	selected, err := selectPromotionItems(cmd, plan, opts)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return writePromoteNoChanges(cmd, source, target, opts)
	}

	finalCfg, applied, err := rcpromote.Apply(plan, selected, rcpromote.Options{Prune: opts.Prune})
	if err != nil {
		return err
	}
	diffText, hasChanges := rc.RenderRemoteConfigDiff(targetCfg, finalCfg)
	if !hasChanges {
		return writePromoteNoChanges(cmd, source, target, opts)
	}
	if !opts.JSON {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nPromote %s -> %s\n", source.ProjectID, target.ProjectID)
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
	}
	if !opts.Yes {
		confirm := shared.NewConfirmation(
			fmt.Sprintf("Publish selected Remote Config changes to %s?", target.ProjectID),
			shared.ConfirmationOptions{},
		)
		ok, err := confirm.RunPrompt()
		if err != nil || !ok {
			return err
		}
	}

	published, err := publishPromotePlan(ctx, cmd, svc, target, sourceCfg, opts, selected)
	if err != nil {
		return err
	}
	if opts.JSON {
		return shared.WriteJSON(cmd, promoteJSON(source, target, opts, published, applied, rcdiff.CompareRemoteConfigs(targetCfg, finalCfg)))
	}
	if published {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🚀 promoted: %s -> %s\n", source.ProjectID, target.ProjectID)
	}
	return nil
}

func loadCompareConfigs(ctx context.Context, cmd *cobra.Command, svc *core.Core, sourceQuery, targetQuery string, cached bool) (core.Project, core.Project, *firebase.RemoteConfig, *firebase.RemoteConfig, error) {
	source, err := shared.ResolveProjectArg(ctx, cmd, svc, sourceQuery)
	if err != nil {
		return core.Project{}, core.Project{}, nil, nil, err
	}
	target, err := shared.ResolveProjectArg(ctx, cmd, svc, targetQuery)
	if err != nil {
		return core.Project{}, core.Project{}, nil, nil, err
	}
	sourceCfg, err := loadProjectConfig(ctx, svc, source.ProjectID, cached)
	if err != nil {
		return core.Project{}, core.Project{}, nil, nil, err
	}
	targetCfg, err := loadProjectConfig(ctx, svc, target.ProjectID, cached)
	if err != nil {
		return core.Project{}, core.Project{}, nil, nil, err
	}
	sourceCfg.Version = firebase.RemoteConfigVersion{}
	targetCfg.Version = firebase.RemoteConfigVersion{}
	return source, target, sourceCfg, targetCfg, nil
}

func loadProjectConfig(ctx context.Context, svc *core.Core, projectID string, cached bool) (*firebase.RemoteConfig, error) {
	if cached {
		cache, _, err := svc.GetParameters(ctx, projectID, false)
		if err != nil {
			return nil, err
		}
		return firebase.ParseCloneRemoteConfig(cache.RemoteConfig)
	}
	raw, _, err := svc.ExportRemoteConfig(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return firebase.ParseCloneRemoteConfig(raw)
}

func publishPromotePlan(ctx context.Context, cmd *cobra.Command, svc *core.Core, target core.Project, sourceCfg *firebase.RemoteConfig, opts compareOptions, selected map[rcpromote.ItemID]bool) (bool, error) {
	if hasDraft, err := svc.HasDraft(target.ProjectID); err != nil {
		return false, err
	} else if hasDraft {
		return false, fmt.Errorf("project %s has an unpublished draft; publish or discard it before promoting", target.ProjectID)
	}
	for {
		raw, etag, err := svc.ExportRemoteConfig(ctx, target.ProjectID)
		if err != nil {
			return false, err
		}
		latestTarget, err := firebase.ParseCloneRemoteConfig(raw)
		if err != nil {
			return false, err
		}
		latestTarget.Version = firebase.RemoteConfigVersion{}
		plan := rcpromote.BuildPlan(sourceCfg, latestTarget, rcpromote.Options{Prune: opts.Prune})
		finalCfg, _, err := rcpromote.Apply(plan, selected, rcpromote.Options{Prune: opts.Prune})
		if err != nil {
			return false, err
		}
		diffText, hasChanges := rc.RenderRemoteConfigDiff(latestTarget, finalCfg)
		if !hasChanges {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No changes")
			return false, nil
		}
		finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
		if err != nil {
			return false, err
		}
		retry, err := rc.ValidateAndPublishRemoteConfig(ctx, svc, target.ProjectID, finalRaw, etag, "promote", cmd.ErrOrStderr())
		if err != nil {
			return false, err
		}
		if retry {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
			continue
		}
		return true, nil
	}
}

func selectPromotionItems(cmd *cobra.Command, plan rcpromote.Plan, opts compareOptions) (map[rcpromote.ItemID]bool, error) {
	if opts.All {
		return rcpromote.SelectAll(plan.Items), nil
	}
	if !opts.Interactive && !hasSelectionIntent(opts) && !isTerminal() {
		return nil, fmt.Errorf("non-interactive promote requires --all, --filter, --group, --expr, or --search")
	}
	if opts.Interactive || isTerminal() {
		return promptPromotionItems(cmd, plan)
	}
	return rcpromote.SelectAll(plan.Items), nil
}

func promptPromotionItems(cmd *cobra.Command, plan rcpromote.Plan) (map[rcpromote.ItemID]bool, error) {
	selected := make(map[rcpromote.ItemID]bool)
	promoteRest := map[rcdiff.ItemKind]bool{}
	skipRest := map[rcdiff.ItemKind]bool{}
	for _, item := range plan.Items {
		if promoteRest[item.Kind] {
			selected[item.ID] = true
			continue
		}
		if skipRest[item.Kind] {
			continue
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\n%s: %s (%s)\n", itemKindTitle(item.Kind), item.Label, item.Change)
		preview, _ := renderPromotionItemDiff(plan, item)
		if preview != "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), preview)
		}
		choice, err := promptPromotionChoice("Promote this item?")
		if err != nil {
			return nil, err
		}
		switch choice {
		case "promote":
			selected[item.ID] = true
		case "skip":
		case "promote-section":
			promoteRest[item.Kind] = true
			selected[item.ID] = true
		case "skip-section":
			skipRest[item.Kind] = true
		case "quit":
			return nil, fmt.Errorf("promotion cancelled")
		}
	}
	return selected, nil
}

func promptPromotionChoice(prompt string) (string, error) {
	p := selection.New(prompt, []promotionChoice{
		{label: "Promote", value: "promote"},
		{label: "Skip", value: "skip"},
		{label: "Promote remaining in this section", value: "promote-section"},
		{label: "Skip remaining in this section", value: "skip-section"},
		{label: "Quit without publishing", value: "quit"},
	})
	choice, err := p.RunPrompt()
	if err != nil {
		return "", err
	}
	return choice.value, nil
}

type promotionChoice struct {
	label string
	value string
}

func (c promotionChoice) String() string {
	return c.label
}

func filterPromotionItems(items []rcpromote.Item, result rcdiff.Result, opts compareOptions) []rcpromote.Item {
	allowed := make(map[rcpromote.ItemID]struct{})
	for _, change := range result.Parameters {
		allowed[rcpromote.ItemID{Kind: rcdiff.ItemParameter, Name: change.Key, Group: change.Group}] = struct{}{}
	}
	for _, change := range result.GroupDescriptions {
		allowed[rcpromote.ItemID{Kind: rcdiff.ItemGroupDescription, Name: change.Group}] = struct{}{}
	}
	for _, change := range result.Conditions {
		allowed[rcpromote.ItemID{Kind: rcdiff.ItemCondition, Name: change.Name}] = struct{}{}
	}

	out := make([]rcpromote.Item, 0, len(items))
	for _, item := range items {
		if _, ok := allowed[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func filterDiffResult(source core.Project, sourceCfg *firebase.RemoteConfig, target core.Project, targetCfg *firebase.RemoteConfig, result rcdiff.Result, opts compareOptions) rcdiff.Result {
	if opts.ParametersOnly && !opts.ConditionsOnly {
		result.Conditions = nil
	}
	if opts.ConditionsOnly && !opts.ParametersOnly {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}

	filters := shared.ParseFilters(opts.ParamFilters)
	groups := groupsSet(opts.Groups)
	compiledExpr, exprOK := shared.CompileExpr(opts.Expr, source.ProjectID)
	if opts.Expr != "" && !exprOK {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}

	params := make([]rcdiff.ParameterChange, 0, len(result.Parameters))
	for _, change := range result.Parameters {
		cfg := sourceCfg
		project := source
		param := change.Final
		group := change.Group
		if change.Kind == rcdiff.ChangeRemoved {
			cfg = targetCfg
			project = target
			param = change.Current
		}
		if param == nil {
			continue
		}
		if len(groups) > 0 && !groups[group] {
			continue
		}
		if !shared.MatchAnyFilter(change.Key, filters) {
			continue
		}
		if !shared.MatchParameterSearch(change.Key, *param, cfg, opts.Search) {
			continue
		}
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, change.Key, groupOrDefault(group))
		if !ok || !match {
			continue
		}
		params = append(params, change)
	}
	result.Parameters = params

	if len(groups) > 0 {
		groupsChanged := make([]rcdiff.GroupDescriptionChange, 0, len(result.GroupDescriptions))
		for _, change := range result.GroupDescriptions {
			if groups[change.Group] {
				groupsChanged = append(groupsChanged, change)
			}
		}
		result.GroupDescriptions = groupsChanged
	}
	if len(filters) > 0 || !opts.Search.Empty() || opts.Expr != "" {
		result.GroupDescriptions = nil
	}
	return result
}

func groupsSet(groups []string) map[string]bool {
	out := make(map[string]bool, len(groups))
	for _, group := range groups {
		out[group] = true
	}
	return out
}

func groupOrDefault(group string) string {
	if group == "" {
		return shared.DefaultRootGroupLabel
	}
	return group
}

func renderPromotionItemDiff(plan rcpromote.Plan, item rcpromote.Item) (string, bool) {
	selected := map[rcpromote.ItemID]bool{item.ID: true}
	finalCfg, _, err := rcpromote.Apply(plan, selected, rcpromote.Options{Prune: item.Change == rcdiff.ChangeRemoved})
	if err != nil {
		return "", false
	}
	return rc.RenderRemoteConfigDiff(plan.Target, finalCfg)
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func hasSelectionIntent(opts compareOptions) bool {
	return opts.All || len(opts.ParamFilters) > 0 || len(opts.Groups) > 0 || opts.Expr != "" || !opts.Search.Empty()
}

func writePromoteNoChanges(cmd *cobra.Command, source, target core.Project, opts compareOptions) error {
	if opts.JSON {
		return shared.WriteJSON(cmd, promoteJSON(source, target, opts, false, nil, rcdiff.Result{}))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No changes")
	return nil
}

func itemKindTitle(kind rcdiff.ItemKind) string {
	switch kind {
	case rcdiff.ItemCondition:
		return "Condition"
	case rcdiff.ItemGroupDescription:
		return "Group"
	default:
		return "Parameter"
	}
}

func normalizeGroups(groups []string) []string {
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		out = append(out, group)
	}
	return out
}

func compareJSON(source, target core.Project, result rcdiff.Result) any {
	return map[string]any{
		"source_project": source.ProjectID,
		"target_project": target.ProjectID,
		"has_changes":    result.HasChanges(),
		"summary": map[string]rcdiff.Summary{
			"conditions":         result.ConditionSummary(),
			"parameters":         result.ParameterSummary(),
			"group_descriptions": result.GroupDescriptionSummary(),
		},
		"changes": result,
	}
}

func promoteJSON(source, target core.Project, opts compareOptions, published bool, applied []rcpromote.Item, result rcdiff.Result) any {
	return map[string]any{
		"source_project": source.ProjectID,
		"target_project": target.ProjectID,
		"dry_run":        opts.DryRun,
		"published":      published,
		"selected":       len(applied),
		"summary": map[string]rcdiff.Summary{
			"conditions":         result.ConditionSummary(),
			"parameters":         result.ParameterSummary(),
			"group_descriptions": result.GroupDescriptionSummary(),
		},
	}
}

package conditions

import (
	"context"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

type loadedConditions struct {
	Project  core.Project         `json:"project"`
	Version  string               `json:"version"`
	Source   string               `json:"source"`
	HasDraft bool                 `json:"hasDraft"`
	Tree     *core.ConditionsTree `json:"-"`
}

type conditionShowResult struct {
	Project   core.Project        `json:"project"`
	Version   string              `json:"version"`
	Source    string              `json:"source" contract:"enum=cache|cache-verified|firebase|draft"`
	HasDraft  bool                `json:"has_draft"`
	Condition core.ConditionEntry `json:"condition"`
}

func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "conditions",
		Short: "Inspect and manage Remote Config conditions",
		Long:  "Inspect and manage condition definitions, evaluation priority, expressions, colors, and parameter usage.",
	}
	cmd.AddCommand(
		newListCommandDefinition(svc),
		newShowCommandDefinition(svc),
		newAddCommandDefinition(svc),
		newEditCommandDefinition(svc),
		newRenameCommandDefinition(svc),
		newMoveCommandDefinition(svc),
		newDeleteCommandDefinition(svc),
		newValidateCommandDefinition(svc),
	)
	invocation.MustRegisterResponsePath(cmd, "list", []core.ConditionEntry{})
	invocation.MustRegisterResponsePath(cmd, "show", conditionShowResult{})
	for _, path := range []string{"add", "edit", "rename", "move", "delete"} {
		invocation.MustRegisterResponsePath(cmd, path, []sharedrc.RemoteMutationJSONResult{}, sharedrc.PlanCreatedResult{})
	}
	invocation.MustRegisterResponsePath(cmd, "validate", conditionValidationResult{})
	return cmd
}

func newListCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "list <project>",
		Short: "List conditions in evaluation priority order",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			loaded, err := load(cmd, svc, args[0])
			if err != nil {
				return err
			}
			filters, _ := cmd.Flags().GetStringArray("filter")
			search, _ := cmd.Flags().GetString("search")
			rawExpr, _ := cmd.Flags().GetString("expr")
			entries := filterEntries(loaded.Tree.Conditions, filters, search)
			entries, err = filterEntriesByExpr(loaded.Project, entries, rawExpr)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, entries)
			}
			printContext(cmd, loaded)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderConditionsTable(entries))
			return nil
		},
	}
	addReadFlags(cmd)
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter conditions by mode-prefixed name query (^, /, ~, =); may be repeated")
	cmd.Flags().String("search", "", "Search condition names and expressions")
	cmd.Flags().String("expr", "", "Filter conditions by expr-lang expression")
	return cmd
}

func newShowCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "show <project> <condition>",
		Short: "Show a condition and every parameter that uses it",
		Args:  invocation.ExactArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			loaded, err := load(cmd, svc, args[0])
			if err != nil {
				return err
			}
			condition, ok := findCondition(loaded.Tree, args[1])
			if !ok {
				return &shared.SelectionError{Resource: "condition", Kind: "not_found", Query: args[1], Err: fmt.Errorf("condition %q not found in project %s", args[1], loaded.Project.ProjectID)}
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, conditionShowResult{Project: loaded.Project, Version: loaded.Version, Source: loaded.Source, HasDraft: loaded.HasDraft, Condition: condition})
			}
			printContext(cmd, loaded)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderConditionDetails(condition))
			return nil
		},
	}
	addReadFlags(cmd)
	return cmd
}

func addReadFlags(cmd invocation.FlagGroups) {
	cmd.Flags().Bool("update", false, "Revalidate cached Remote Config before printing")
	cmd.Flags().Bool("json", false, "Print conditions as JSON")
}

func load(cmd invocation.Call, svc *core.Core, query string) (loadedConditions, error) {
	ctx := shared.CommandContext(cmd)
	update, _ := cmd.Flags().GetBool("update")
	if !core.ExecutionPolicyFromContext(ctx).ReadLocalState && update {
		return loadedConditions{}, shared.InvalidArgument(fmt.Errorf("--update cannot be used with --stateless; Remote Config reads are already live"))
	}
	project, err := shared.ResolveProjectTargetForExecution(ctx, cmd, svc, query)
	if err != nil {
		return loadedConditions{}, err
	}
	ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return loadedConditions{}, err
	}
	cmd.SetContext(ctx)
	cache, source, err := loadCache(ctx, svc, project.ProjectID, update)
	if err != nil {
		return loadedConditions{}, err
	}
	var tree *core.ConditionsTree
	var hasDraft bool
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		tree, hasDraft, err = svc.BuildDraftAwareConditionsTree(project.ProjectID, cache)
	} else {
		tree, err = svc.BuildConditionsTree(cache)
	}
	if err != nil {
		return loadedConditions{}, err
	}
	if hasDraft {
		source = "draft"
	}
	return loadedConditions{Project: project, Version: tree.Version, Source: source, HasDraft: hasDraft, Tree: tree}, nil
}

func loadCache(ctx context.Context, svc *core.Core, projectID string, update bool) (*core.ParametersCache, string, error) {
	var cache *core.ParametersCache
	var source string
	var err error
	if update {
		cache, source, err = svc.RevalidateParameters(ctx, projectID)
	} else {
		cache, source, err = svc.GetParameters(ctx, projectID, false)
	}
	if err == nil {
		return cache, source, nil
	}
	if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return nil, "", err
	}

	stale, state, inspectErr := svc.InspectParametersCache(projectID)
	if inspectErr == nil && state != core.ParametersCacheMissing && stale != nil {
		return stale, "cache-stale", nil
	}
	return nil, "", err
}

func filterEntries(entries []core.ConditionEntry, rawFilters []string, search string) []core.ConditionEntry {
	filters := shared.ParseFilters(rawFilters)
	search = strings.ToLower(strings.TrimSpace(search))
	out := make([]core.ConditionEntry, 0, len(entries))
	for _, entry := range entries {
		if !shared.MatchAnyFilter(entry.Name, filters) {
			continue
		}
		if search != "" {
			haystack := strings.ToLower(entry.Name + "\n" + entry.Expression)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		out = append(out, entry)
	}
	return out
}

func filterEntriesByExpr(project core.Project, entries []core.ConditionEntry, rawExpr string) ([]core.ConditionEntry, error) {
	compiled, err := shared.CompileExpr(rawExpr, project.ProjectID)
	if err != nil {
		return nil, err
	}
	out := make([]core.ConditionEntry, 0, len(entries))
	for _, entry := range entries {
		match, err := shared.MatchConditionByCompiledExpr(compiled, project, entry)
		if err != nil {
			return nil, err
		}
		if match {
			out = append(out, entry)
		}
	}
	return out, nil
}

func findCondition(tree *core.ConditionsTree, name string) (core.ConditionEntry, bool) {
	return tree.Find(name)
}

func printContext(cmd invocation.Call, loaded loadedConditions) {
	source := loaded.Source
	if loaded.HasDraft {
		source = "draft"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nVersion: %s · Source: %s\n\n", loaded.Project.Name, loaded.Project.ProjectID, loaded.Version, source)
}

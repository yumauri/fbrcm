package conditions

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
)

type loadedConditions struct {
	Project  core.Project         `json:"project"`
	Version  string               `json:"version"`
	Source   string               `json:"source"`
	HasDraft bool                 `json:"hasDraft"`
	Tree     *core.ConditionsTree `json:"-"`
}

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conditions",
		Short: "Inspect Remote Config conditions",
		Long:  "Inspect condition priority, expressions, and the parameters that use each condition.",
	}
	cmd.AddCommand(newListCommand(svc), newShowCommand(svc))
	return cmd
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List conditions in evaluation priority order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := load(cmd, svc, args[0])
			if err != nil {
				return err
			}
			filters, _ := cmd.Flags().GetStringArray("filter")
			search, _ := cmd.Flags().GetString("search")
			entries := filterEntries(loaded.Tree.Conditions, filters, search)
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{
					"project":    loaded.Project,
					"version":    loaded.Version,
					"source":     loaded.Source,
					"has_draft":  loaded.HasDraft,
					"conditions": entries,
				})
			}
			printContext(cmd, loaded)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderConditionsTable(entries))
			return nil
		},
	}
	addReadFlags(cmd)
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter conditions by mode-prefixed name query (^, /, ~, =); may be repeated")
	cmd.Flags().String("search", "", "Search condition names, expressions, and descriptions")
	return cmd
}

func newShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <project> <condition>",
		Short: "Show a condition and every parameter that uses it",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := load(cmd, svc, args[0])
			if err != nil {
				return err
			}
			condition, ok := findCondition(loaded.Tree, args[1])
			if !ok {
				return fmt.Errorf("condition %q not found in project %s", args[1], loaded.Project.ProjectID)
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{
					"project":   loaded.Project,
					"version":   loaded.Version,
					"source":    loaded.Source,
					"has_draft": loaded.HasDraft,
					"condition": condition,
				})
			}
			printContext(cmd, loaded)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderConditionDetails(condition))
			return nil
		},
	}
	addReadFlags(cmd)
	return cmd
}

func addReadFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("update", false, "Revalidate cached Remote Config before printing")
	cmd.Flags().Bool("json", false, "Print conditions as JSON")
}

func load(cmd *cobra.Command, svc *core.Core, query string) (loadedConditions, error) {
	project, err := shared.ResolveProjectArg(context.Background(), cmd, svc, query)
	if err != nil {
		return loadedConditions{}, err
	}
	update, _ := cmd.Flags().GetBool("update")
	cache, source, err := loadCache(context.Background(), svc, project.ProjectID, update)
	if err != nil {
		return loadedConditions{}, err
	}
	tree, hasDraft, err := svc.BuildDraftAwareConditionsTree(project.ProjectID, cache)
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
			haystack := strings.ToLower(entry.Name + "\n" + entry.Expression + "\n" + entry.Description)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		out = append(out, entry)
	}
	return out
}

func findCondition(tree *core.ConditionsTree, name string) (core.ConditionEntry, bool) {
	if condition, ok := tree.Find(name); ok {
		return condition, true
	}
	for _, condition := range tree.Conditions {
		if strings.EqualFold(condition.Name, name) {
			return condition, true
		}
	}
	return core.ConditionEntry{}, false
}

func printContext(cmd *cobra.Command, loaded loadedConditions) {
	source := loaded.Source
	if loaded.HasDraft {
		source = "draft"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nVersion: %s · Source: %s\n\n", loaded.Project.Name, loaded.Project.ProjectID, loaded.Version, source)
}

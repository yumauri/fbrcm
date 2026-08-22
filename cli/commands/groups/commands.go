package groups

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/core/strfold"
)

type loadedGroups struct {
	Project  core.Project
	Version  string
	Source   string
	HasDraft bool
	Groups   []core.ParametersGroup
}

type groupListJSON struct {
	Project        string `json:"project"`
	ProjectID      string `json:"project_id"`
	Version        string `json:"version"`
	Source         string `json:"source" contract:"enum=cache|cache-verified|firebase|draft"`
	HasDraft       bool   `json:"has_draft"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	ParameterCount int    `json:"parameter_count"`
}

type projectGroup struct {
	Project  core.Project
	Version  string
	Source   string
	HasDraft bool
	Group    core.ParametersGroup
}

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "groups", Short: "Inspect and manage Remote Config parameter groups"}
	cmd.AddCommand(newListCommand(svc), newAddCommand(svc), newEditCommand(svc), newRenameCommand(svc), newDeleteCommand(svc))
	contract.MustRegisterResponsePath(cmd, "list", []groupListJSON{})
	for _, path := range []string{"add", "edit", "rename", "delete"} {
		contract.MustRegisterResponsePath(cmd, path, []sharedrc.RemoteMutationJSONResult{})
	}
	return cmd
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list", Short: "List parameter groups across projects", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			projectFilters, _ := cmd.Flags().GetStringArray("project")
			loaded, err := loadProjects(cmd, svc)
			if err != nil {
				return err
			}
			filters, _ := cmd.Flags().GetStringArray("filter")
			search, _ := cmd.Flags().GetString("search")
			entries := filterLoadedGroups(loaded, filters, search)
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, groupsJSON(entries))
			}
			oneExactTarget := shared.SingleExactProjectTargetFilter(projectFilters) && len(loaded) == 1
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderGroupsTable(entries, !oneExactTarget))
			return nil
		},
	}
	addReadFlags(cmd)
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter group names by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().String("search", "", "Search group names and descriptions")
	return cmd
}

func addReadFlags(cmd *cobra.Command) {
	shared.AddProjectTargetFilterFlag(cmd)
	cmd.Flags().Bool("update", false, "Revalidate cached Remote Config before printing")
	cmd.Flags().Bool("json", false, "Print groups as JSON")
}

func loadProjects(cmd *cobra.Command, svc *core.Core) ([]loadedGroups, error) {
	ctx := shared.CommandContext(cmd)
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return nil, err
	}
	update, _ := cmd.Flags().GetBool("update")
	if !core.ExecutionPolicyFromContext(ctx).ReadLocalState && update {
		return nil, shared.InvalidArgument(fmt.Errorf("--update cannot be used with --stateless; Remote Config reads are already live"))
	}
	projects, ctx, err := shared.ResolveProjectTargetsForExecution(ctx, cmd, svc, projectFilters)
	if err != nil {
		return nil, err
	}
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })
	loaded := make([]loadedGroups, 0, len(projects))
	for _, project := range projects {
		projectCtx, err := shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
		if err != nil {
			return nil, err
		}
		item, err := loadProjectGroups(projectCtx, svc, project, update)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, item)
	}
	return loaded, nil
}

func loadProjectGroups(ctx context.Context, svc *core.Core, project core.Project, update bool) (loadedGroups, error) {
	cache, source, err := loadCache(ctx, svc, project.ProjectID, update)
	if err != nil {
		return loadedGroups{}, err
	}
	var tree *core.ParametersTree
	var hasDraft bool
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		tree, hasDraft, err = svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
	} else {
		tree, err = svc.BuildParametersTree(cache)
	}
	if err != nil {
		return loadedGroups{}, err
	}
	groups := make([]core.ParametersGroup, 0, len(tree.Groups))
	for _, group := range tree.Groups {
		if core.NormalizeRemoteConfigGroupKey(group.Key) != rootgroup.WireKey {
			groups = append(groups, group)
		}
	}
	if hasDraft {
		source = "draft"
	}
	return loadedGroups{Project: project, Version: tree.Version, Source: source, HasDraft: hasDraft, Groups: groups}, nil
}

func filterLoadedGroups(loaded []loadedGroups, rawFilters []string, search string) []projectGroup {
	out := make([]projectGroup, 0)
	for _, item := range loaded {
		for _, group := range filterGroups(item.Groups, rawFilters, search) {
			out = append(out, projectGroup{Project: item.Project, Version: item.Version, Source: item.Source, HasDraft: item.HasDraft, Group: group})
		}
	}
	return out
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

func filterGroups(groups []core.ParametersGroup, rawFilters []string, search string) []core.ParametersGroup {
	filters := shared.ParseFilters(rawFilters)
	search = strings.ToLower(strings.TrimSpace(search))
	out := make([]core.ParametersGroup, 0, len(groups))
	for _, group := range groups {
		if !shared.MatchAnyFilter(group.Key, filters) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(group.Key+"\n"+group.Description), search) {
			continue
		}
		out = append(out, group)
	}
	return out
}

func groupsJSON(groups []projectGroup) []groupListJSON {
	out := make([]groupListJSON, len(groups))
	for i, item := range groups {
		out[i] = groupListJSON{
			Project: item.Project.Name, ProjectID: item.Project.ProjectID, Version: item.Version, Source: item.Source, HasDraft: item.HasDraft,
			Name: item.Group.Key, Description: item.Group.Description, ParameterCount: len(item.Group.Parameters),
		}
	}
	return out
}

package projects

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared"
)

type visibilityEntry struct {
	ProjectID    string `json:"project_id"`
	Project      string `json:"project"`
	Known        bool   `json:"known"`
	HiddenReason string `json:"hidden_reason,omitempty" contract:"enum=blacklisted|not_whitelisted"`
}

type visibilityListResult struct {
	List                string            `json:"list" contract:"enum=blacklist|whitelist"`
	Enabled             bool              `json:"enabled"`
	Count               int               `json:"count"`
	Items               []visibilityEntry `json:"items"`
	VisibleProjectCount int               `json:"visible_project_count"`
}

type visibilitySelection struct {
	Kind             string   `json:"kind" contract:"enum=ids|filters|list"`
	ProjectIDs       []string `json:"project_ids"`
	Filters          []string `json:"filters"`
	MatchedItemCount int      `json:"matched_item_count"`
}

type visibilityMutationResult struct {
	List                string              `json:"list" contract:"enum=blacklist|whitelist"`
	Operation           string              `json:"operation" contract:"enum=add|remove|clear|enable|disable"`
	Enabled             bool                `json:"enabled"`
	Changed             bool                `json:"changed"`
	AffectedProjectIDs  []string            `json:"affected_project_ids"`
	Count               int                 `json:"count"`
	VisibleProjectCount int                 `json:"visible_project_count"`
	Selection           visibilitySelection `json:"selection"`
	NoOpReason          *string             `json:"no_op_reason" contract:"enum=no_match|already_applied"`
}

func newVisibilityCommandDefinition(name string) *invocation.Definition {
	group := &invocation.Definition{Use: name, Short: "Manage the profile project " + name}
	list := &invocation.Definition{Use: "list", Short: "List stored project IDs and activation state", Args: invocation.NoArgs,
		RunE: func(cmd invocation.Call, _ []string) error { return listVisibility(cmd, name) },
	}
	invocation.RegisterResponse(list, visibilityListResult{})
	group.AddCommand(list)
	for _, action := range []string{"add", "remove", "clear", "enable", "disable"} {
		description := map[string]string{
			"add": "Add project IDs to the profile ", "remove": "Remove project IDs from the profile ",
			"clear": "Clear the profile project ", "enable": "Enable the profile project ", "disable": "Disable the profile project ",
		}[action] + name
		command := &invocation.Definition{Use: action, Short: description, Args: invocation.NoArgs,
			RunE: func(cmd invocation.Call, args []string) error { return mutateVisibility(cmd, name, action, args) },
		}
		if action == "add" || action == "remove" {
			command.Use += " [project-id...]"
			command.Args = invocation.ArbitraryArgs
			shared.AddProjectListFilterFlag(command)
		}
		invocation.RegisterResponse(command, visibilityMutationResult{})
		group.AddCommand(command)
	}
	return group
}

func visibilityList(policy *config.ProjectVisibility, name string) *config.ProjectVisibilityList {
	if name == "blacklist" {
		return &policy.Blacklist
	}
	return &policy.Whitelist
}

// These commands are local inventory administration and never discover projects.
func visibilityInventory() ([]core.Project, error) {
	projects, err := config.LoadProjects()
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, config.ErrEmptyProjectsFile) {
		return []core.Project{}, nil
	}
	return projects, err
}

func visibilityEntries(policy config.ProjectVisibility, ids []string, inventory []core.Project) []visibilityEntry {
	byID := make(map[string]core.Project, len(inventory))
	for _, project := range inventory {
		byID[project.ProjectID] = project
	}
	entries := make([]visibilityEntry, 0, len(ids))
	for _, id := range ids {
		project, known := byID[id]
		entries = append(entries, visibilityEntry{ProjectID: id, Project: project.Name, Known: known, HiddenReason: policy.HiddenReason(id)})
	}
	return entries
}

func visibleProjectCount(policy config.ProjectVisibility, inventory []core.Project) int {
	count := 0
	for _, project := range inventory {
		if policy.HiddenReason(project.ProjectID) == "" {
			count++
		}
	}
	return count
}

func listVisibility(cmd invocation.Call, name string) error {
	policy, err := config.LoadProjectVisibility()
	if err != nil {
		return err
	}
	projects, err := visibilityInventory()
	if err != nil {
		return err
	}
	list := visibilityList(&policy, name)
	result := visibilityListResult{List: name, Enabled: list.Enabled, Count: len(list.ProjectIDs), Items: visibilityEntries(policy, list.ProjectIDs, projects), VisibleProjectCount: visibleProjectCount(policy, projects)}
	if shared.MachineMode(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	state := "disabled"
	if list.Enabled {
		state = "enabled"
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %s; %s stored; %s visible\n", name, state, rcdisplay.FormatCount(result.Count, "project", "projects"), rcdisplay.FormatCount(result.VisibleProjectCount, "project", "projects")); err != nil {
		return err
	}
	rows := make([]core.Project, 0, len(result.Items))
	for _, entry := range result.Items {
		label := entry.Project
		if !entry.Known {
			label = "Unknown project"
		}
		if entry.HiddenReason != "" {
			label += " (" + core.ProjectVisibilityReasonText(entry.HiddenReason) + ")"
		}
		rows = append(rows, core.Project{Name: label, ProjectID: entry.ProjectID})
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), shared.RenderProjectsChoiceTable(rows))
	return err
}

func selectVisibilityIDs(cmd invocation.Call, action string, args []string, list config.ProjectVisibilityList, inventory []core.Project) (visibilitySelection, error) {
	selection := visibilitySelection{Kind: "list", ProjectIDs: []string{}, Filters: []string{}}
	if action != "add" && action != "remove" {
		selection.ProjectIDs = append(selection.ProjectIDs, list.ProjectIDs...)
		selection.MatchedItemCount = len(selection.ProjectIDs)
		return selection, nil
	}
	filters, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return selection, err
	}
	if len(args) > 0 && cmd.Flags().Changed("filter") {
		return selection, shared.InvalidArgument(fmt.Errorf("project IDs and --filter cannot be used together"))
	}
	if len(args) == 0 && len(filters) == 0 {
		return selection, shared.InvalidArgument(fmt.Errorf("provide at least one literal project ID or --filter"))
	}
	selection.Kind = "ids"
	for _, id := range args {
		if err := config.ValidatePhysicalProjectID(id); err != nil {
			return selection, shared.InvalidArgument(err)
		}
	}
	selection.ProjectIDs = append(selection.ProjectIDs, args...)
	if len(filters) > 0 {
		if err := shared.RejectTemplateFilters(filters, "filter"); err != nil {
			return selection, err
		}
		selection.Kind = "filters"
		selection.Filters = append(selection.Filters, filters...)
		if action == "remove" {
			entries := visibilityEntries(config.DefaultProjectVisibility(), list.ProjectIDs, inventory)
			inventory = make([]core.Project, 0, len(entries))
			for _, entry := range entries {
				inventory = append(inventory, core.Project{ProjectID: entry.ProjectID, Name: entry.Project})
			}
		}
		matched, err := shared.FilterProjects(inventory, filters)
		if err != nil {
			return selection, err
		}
		for _, project := range matched {
			selection.ProjectIDs = append(selection.ProjectIDs, project.ProjectID)
		}
	}
	strfold.Sort(selection.ProjectIDs)
	selection.ProjectIDs = slices.Compact(selection.ProjectIDs)
	selection.MatchedItemCount = len(selection.ProjectIDs)
	return selection, nil
}

func mutateVisibility(cmd invocation.Call, name, action string, args []string) error {
	policy, err := config.LoadProjectVisibility()
	if err != nil {
		return err
	}
	projects, err := visibilityInventory()
	if err != nil {
		return err
	}
	list := visibilityList(&policy, name)
	selection, err := selectVisibilityIDs(cmd, action, args, *list, projects)
	if err != nil {
		return err
	}
	result := visibilityMutationResult{List: name, Operation: action, Selection: selection, AffectedProjectIDs: []string{}}
	for _, id := range selection.ProjectIDs {
		contains := slices.Contains(list.ProjectIDs, id)
		switch {
		case action == "add" && !contains:
			list.ProjectIDs = append(list.ProjectIDs, id)
			result.AffectedProjectIDs = append(result.AffectedProjectIDs, id)
		case (action == "remove" || action == "clear") && contains:
			list.ProjectIDs = slices.DeleteFunc(list.ProjectIDs, func(value string) bool { return value == id })
			result.AffectedProjectIDs = append(result.AffectedProjectIDs, id)
		}
	}
	result.Changed = len(result.AffectedProjectIDs) > 0
	if action == "enable" || action == "disable" {
		enabled := action == "enable"
		result.Changed = list.Enabled != enabled
		list.Enabled = enabled
	}
	result.Enabled = list.Enabled
	result.Count = len(list.ProjectIDs)
	result.VisibleProjectCount = visibleProjectCount(policy, projects)
	if !result.Changed {
		reason := "already_applied"
		if selection.MatchedItemCount == 0 && (action == "add" || action == "remove" || action == "clear") {
			reason = "no_match"
		}
		result.NoOpReason = &reason
	} else if err := config.SaveProjectVisibility(policy); err != nil {
		return err
	}
	if name == "whitelist" && action == "add" {
		for _, id := range selection.ProjectIDs {
			if policy.HiddenReason(id) == "blacklisted" {
				machine.FromContext(shared.CommandContext(cmd)).AddWarning(machine.Warning{
					Code: "project.remains_hidden", Message: "Project remains hidden by the blacklist.", Target: id,
					Details: struct {
						Reason string `json:"reason"`
					}{Reason: "blacklisted"},
					Remediation: []machine.Remediation{{Description: "remove the project from the blacklist", Strategy: machine.RemediationRunCommand, Argv: []string{"projects", "blacklist", "remove", "--", id}}},
				})
			}
		}
	}
	if shared.MachineMode(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	verb := "updated"
	if !result.Changed {
		verb = "unchanged"
	}
	state := "disabled"
	if result.Enabled {
		state = "enabled"
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s): %s affected; %s stored; %s visible\n", verb, name, state, rcdisplay.FormatCount(len(result.AffectedProjectIDs), "project", "projects"), rcdisplay.FormatCount(result.Count, "project", "projects"), rcdisplay.FormatCount(result.VisibleProjectCount, "project", "projects"))
	if err != nil {
		return err
	}
	if name == "whitelist" && action == "add" && policy.Blacklist.Enabled {
		for _, id := range selection.ProjectIDs {
			if policy.HiddenReason(id) == "blacklisted" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Project %s remains hidden by the blacklist.\n", id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

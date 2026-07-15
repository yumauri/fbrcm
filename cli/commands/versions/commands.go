package versions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// New constructs the top-level versions command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "versions", Short: "Inspect and recover project Remote Config versions", Long: "Inspect Firebase Remote Config version history and immutable local snapshots. Version selectors accept a number, current, latest, previous, current~N, or latest~N. Rollback uses Firebase history; restore republishes a locally cached snapshot."}
	cmd.AddCommand(newVersionsListCommand(svc), newVersionsShowCommand(svc), newVersionsDiffCommand(svc), newVersionsExportCommand(svc), newVersionsRollbackCommand(svc, false), newVersionsRollbackCommand(svc, true))
	return cmd
}

func resolveVersionProject(cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	return shared.ResolveProjectArg(context.Background(), cmd, svc, query)
}

func newVersionsListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "list <project>", Short: "List Remote Config versions", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		project, err := resolveVersionProject(cmd, svc, args[0])
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		if limit < 1 {
			return fmt.Errorf("--limit must be greater than zero")
		}
		all, _ := cmd.Flags().GetBool("all")
		cached, _ := cmd.Flags().GetBool("cached")
		before, _ := cmd.Flags().GetString("before")
		since, err := timeFlag(cmd, "since")
		if err != nil {
			return err
		}
		until, err := timeFlag(cmd, "until")
		if err != nil {
			return err
		}
		result, err := svc.ListRemoteConfigVersions(context.Background(), project.ProjectID, core.VersionListOptions{Limit: limit, All: all, Before: before, Since: since, Until: until, CachedOnly: cached})
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			return shared.WriteJSON(cmd, map[string]any{"project": project, "versions": result.Versions, "next_page_token": result.NextPageToken})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\n\n", project.Name, project.ProjectID)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderVersionsTable(result.Versions, cached))
		return nil
	}}
	cmd.Flags().Int("limit", 20, "Maximum versions to print")
	cmd.Flags().Bool("all", false, "Print all available Firebase versions")
	cmd.Flags().String("before", "", "Newest version number to include")
	cmd.Flags().String("since", "", "Only versions at or after RFC3339 time")
	cmd.Flags().String("until", "", "Only versions before RFC3339 time")
	cmd.Flags().Bool("cached", false, "List local cached versions without contacting Firebase")
	cmd.Flags().Bool("json", false, "Print versions as JSON")
	cmd.MarkFlagsMutuallyExclusive("all", "limit")
	return cmd
}

func timeFlag(cmd *cobra.Command, name string) (time.Time, error) {
	raw, _ := cmd.Flags().GetString(name)
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --%s time %q: use RFC3339", name, raw)
	}
	return value, nil
}

func newVersionsShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "show <project> <version>", Short: "Show Remote Config version metadata or JSON", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		project, err := resolveVersionProject(cmd, svc, args[0])
		if err != nil {
			return err
		}
		cached, _ := cmd.Flags().GetBool("cached")
		resolved, err := svc.GetRemoteConfigVersion(context.Background(), project.ProjectID, args[1], cached)
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			return shared.WriteJSON(cmd, map[string]any{"project": project, "version": resolved.Version, "cached": resolved.Cached})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nVersion: %s\nPublished: %s\nUpdated by: %s\nOrigin: %s\nType: %s\nDescription: %s\nRollback source: %s\nCached: %t\n", project.Name, project.ProjectID, resolved.Version.VersionNumber, resolved.Version.UpdateTime, resolved.Version.UpdateUser.Email, resolved.Version.UpdateOrigin, resolved.Version.UpdateType, resolved.Version.Description, resolved.Version.RollbackSource, resolved.Cached)
		return nil
	}}
	cmd.Flags().Bool("cached", false, "Require a local snapshot and do not contact Firebase")
	cmd.Flags().Bool("json", false, "Print metadata as JSON")
	return cmd
}

type versionDiffOptions struct {
	cached, json, parameters, conditions bool
	filters, groups                      []string
	search, expr                         string
}

func newVersionsDiffCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "diff <project> <from> [<to>]", Short: "Compare two Remote Config versions", Args: cobra.RangeArgs(2, 3), RunE: func(cmd *cobra.Command, args []string) error {
		project, err := resolveVersionProject(cmd, svc, args[0])
		if err != nil {
			return err
		}
		to := "current"
		if len(args) == 3 {
			to = args[2]
		}
		opts := readVersionDiffOptions(cmd)
		fromCfg, toCfg, err := svc.GetRemoteConfigVersionPair(context.Background(), project.ProjectID, args[1], to, opts.cached)
		if err != nil {
			return err
		}
		result := filterVersionDiff(project, rcdiff.CompareRemoteConfigs(fromCfg.Config, toCfg.Config), fromCfg.Config, toCfg.Config, opts)
		if opts.json {
			return shared.WriteJSON(cmd, map[string]any{"project": project, "from_version": fromCfg.Version.VersionNumber, "to_version": toCfg.Version.VersionNumber, "diff": result})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s (%s): version %s → version %s\n", project.Name, project.ProjectID, fromCfg.Version.VersionNumber, toCfg.Version.VersionNumber)
		text, changed := rcdiff.RenderResult(result)
		if !changed {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
			return nil
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), text)
		return nil
	}}
	addVersionDiffFlags(cmd)
	return cmd
}

func addVersionDiffFlags(cmd *cobra.Command) {
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().StringArray("group", nil, "Select parameters in group; may be repeated")
	cmd.Flags().String("expr", "", "Filter parameter changes by expr-lang expression")
	cmd.Flags().Bool("parameters", false, "Include only parameter and group description changes")
	cmd.Flags().Bool("conditions", false, "Include only condition changes")
	cmd.Flags().Bool("cached", false, "Require local snapshots and do not contact Firebase")
	cmd.Flags().Bool("json", false, "Print diff as JSON")
	cmd.MarkFlagsMutuallyExclusive("parameters", "conditions")
}
func readVersionDiffOptions(cmd *cobra.Command) versionDiffOptions {
	o := versionDiffOptions{}
	o.cached, _ = cmd.Flags().GetBool("cached")
	o.json, _ = cmd.Flags().GetBool("json")
	o.parameters, _ = cmd.Flags().GetBool("parameters")
	o.conditions, _ = cmd.Flags().GetBool("conditions")
	o.filters, _ = cmd.Flags().GetStringArray("filter")
	o.groups, _ = cmd.Flags().GetStringArray("group")
	o.search, _ = cmd.Flags().GetString("search")
	o.expr, _ = cmd.Flags().GetString("expr")
	return o
}

func filterVersionDiff(project core.Project, result rcdiff.Result, from, to *firebase.RemoteConfig, opts versionDiffOptions) rcdiff.Result {
	if opts.parameters {
		result.Conditions = nil
	}
	if opts.conditions {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}
	filters := shared.ParseFilters(opts.filters)
	groupSet := map[string]bool{}
	for _, g := range opts.groups {
		groupSet[g] = true
	}
	compiledExpr, exprOK := shared.CompileExpr(strings.TrimSpace(opts.expr), project.ProjectID)
	if strings.TrimSpace(opts.expr) != "" && !exprOK {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}
	search := shared.NewParameterSearch(opts.search)
	params := result.Parameters[:0]
	for _, change := range result.Parameters {
		param := change.Final
		cfg := to
		if param == nil {
			param = change.Current
			cfg = from
		}
		if param == nil {
			continue
		}
		if len(groupSet) > 0 && !groupSet[change.Group] {
			continue
		}
		if !shared.MatchAnyFilter(change.Key, filters) || !shared.MatchParameterSearch(change.Key, *param, cfg, search) {
			continue
		}
		group := change.Group
		if group == "" {
			group = "default"
		}
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, change.Key, group)
		if !ok || !match {
			continue
		}
		params = append(params, change)
	}
	result.Parameters = params
	if len(groupSet) > 0 {
		groups := result.GroupDescriptions[:0]
		for _, change := range result.GroupDescriptions {
			if groupSet[change.Group] {
				groups = append(groups, change)
			}
		}
		result.GroupDescriptions = groups
	}
	return result
}

func newVersionsExportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "export <project> <version>", Short: "Export historical Remote Config JSON", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		project, err := resolveVersionProject(cmd, svc, args[0])
		if err != nil {
			return err
		}
		cached, _ := cmd.Flags().GetBool("cached")
		resolved, err := svc.GetRemoteConfigVersion(context.Background(), project.ProjectID, args[1], cached)
		if err != nil {
			return err
		}
		to, _ := cmd.Flags().GetString("to")
		if to == "" {
			body := rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(resolved.Cache.RemoteConfig))
			_, err = cmd.OutOrStdout().Write(body)
			return err
		}
		if err := rc.WriteRemoteConfigFile(to, resolved.Cache.RemoteConfig); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported version %s: %s\n", resolved.Version.VersionNumber, to)
		return nil
	}}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	cmd.Flags().Bool("cached", false, "Require a local snapshot and do not contact Firebase")
	return cmd
}

func newVersionsRollbackCommand(svc *core.Core, restore bool) *cobra.Command {
	name, short := "rollback", "Roll back to a Firebase Remote Config version"
	if restore {
		name, short = "restore", "Republish a locally cached Remote Config version"
	}
	cmd := &cobra.Command{Use: name + " <project> <version>", Short: short, Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionPublish(cmd, svc, args[0], args[1], restore)
	}}
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Skip final publish confirmation")
	cmd.Flags().Bool("json", false, "Print result as JSON")
	return cmd
}

func runVersionPublish(cmd *cobra.Command, svc *core.Core, query, selector string, restore bool) error {
	ctx := context.Background()
	dry, _ := cmd.Flags().GetBool("dry-run")
	if dry {
		ctx = firebase.WithDryRun(ctx)
	}
	project, err := resolveVersionProject(cmd, svc, query)
	if err != nil {
		return err
	}
	if hasDraft, draftErr := svc.HasDraft(project.ProjectID); draftErr != nil {
		return draftErr
	} else if hasDraft {
		return fmt.Errorf("project %s has an unpublished draft; publish or discard it before changing versions", project.ProjectID)
	}
	target, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, selector, restore)
	if err != nil {
		return err
	}
	current, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, "current", false)
	if err != nil {
		return err
	}
	if current.Version.VersionNumber == target.Version.VersionNumber {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "version %s is already current; no operation performed\n", target.Version.VersionNumber)
		return nil
	}
	diffText, changed := rc.RenderRemoteConfigDiff(current.Config, target.Config)
	if !changed {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
		return nil
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	if !jsonOut {
		op := "Rollback"
		note := "Rollback publishes the selected historical template as a new Remote Config version."
		if restore {
			op = "Restore"
			note = "Restore publishes the cached snapshot as a normal new Remote Config version."
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s %s (%s)\nCurrent version: %s\nSource version:  %s\n%s\n%s\n", op, project.Name, project.ProjectID, current.Version.VersionNumber, target.Version.VersionNumber, diffText, note)
	}
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && !dry {
		confirm := shared.NewConfirmation(fmt.Sprintf("Publish this %s to %s?", map[bool]string{true: "restore", false: "rollback"}[restore], project.ProjectID), shared.ConfirmationOptions{})
		confirm.Output = cmd.ErrOrStderr()
		ok, err := confirm.RunPrompt()
		if err != nil || !ok {
			return err
		}
	}
	if dry {
		if restore {
			if _, err := svc.RestoreRemoteConfigVersion(ctx, project.ProjectID, target.Version.VersionNumber); err != nil {
				return err
			}
		}
		result := map[string]any{"project_id": project.ProjectID, "operation": map[bool]string{true: "restore", false: "rollback"}[restore], "previous_version": current.Version.VersionNumber, "source_version": target.Version.VersionNumber, "published_version": nil, "dry_run": true, "changed": true}
		if jsonOut {
			return shared.WriteJSON(cmd, result)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧪 dry run: %s would use version %s\n", project.ProjectID, target.Version.VersionNumber)
		return nil
	}
	latest, err := svc.GetRemoteConfigVersion(context.Background(), project.ProjectID, "current", false)
	if err != nil {
		return err
	}
	if latest.Version.VersionNumber != current.Version.VersionNumber {
		return fmt.Errorf("remote config changed from version %s to %s during preview; rerun the command", current.Version.VersionNumber, latest.Version.VersionNumber)
	}
	var result core.VersionPublishResult
	if restore {
		result, err = svc.RestoreRemoteConfigVersion(context.Background(), project.ProjectID, target.Version.VersionNumber)
	} else {
		result, err = svc.RollbackRemoteConfig(context.Background(), project.ProjectID, target.Version.VersionNumber)
	}
	if err != nil {
		if !restore && target.Cached {
			return fmt.Errorf("%w; if Firebase no longer retains version %s, republish the local snapshot with: fbrcm versions restore %s %s", err, target.Version.VersionNumber, project.ProjectID, target.Version.VersionNumber)
		}
		return err
	}
	payload := map[string]any{"project_id": project.ProjectID, "operation": map[bool]string{true: "restore", false: "rollback"}[restore], "previous_version": result.PreviousVersion, "source_version": result.SourceVersion, "published_version": result.PublishedVersion, "dry_run": false, "changed": true}
	if jsonOut {
		return shared.WriteJSON(cmd, payload)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s v%s → v%s, using v%s\n", map[bool]string{true: "♻️ restored", false: "⏪ rolled back"}[restore], project.ProjectID, result.PreviousVersion, result.PublishedVersion, result.SourceVersion)
	return nil
}

package versions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/core/rc/publication"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

// New constructs the top-level versions command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "versions", Short: "Inspect and recover project Remote Config versions", Long: "Inspect Firebase Remote Config version history and immutable local snapshots. Version selectors accept a number, current, latest, previous, current~N, or latest~N. Rollback uses Firebase history; restore republishes a locally cached snapshot."}
	cmd.AddCommand(newVersionsListCommand(svc), newVersionsShowCommand(svc), newVersionsDiffCommand(svc), newVersionsExportCommand(svc), newVersionsRollbackCommand(svc, false), newVersionsRollbackCommand(svc, true))
	contract.MustRegisterResponsePath(cmd, "list", []versionJSON{})
	contract.MustRegisterResponsePath(cmd, "show", versionShowResult{})
	contract.MustRegisterResponsePath(cmd, "diff", versionDiffResult{})
	contract.MustRegisterResponsePath(cmd, "export", contract.ArtifactData{})
	contract.MustRegisterResponsePath(cmd, "rollback", versionPublishResult{})
	contract.MustRegisterResponsePath(cmd, "restore", versionPublishResult{}, rc.PlanCreatedResult{})
	return cmd
}

func resolveVersionProject(cmd *cobra.Command, svc *core.Core, query string, cachedOnly bool) (core.Project, error) {
	ctx := shared.CommandContext(cmd)
	if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		if cachedOnly {
			return core.Project{}, shared.InvalidArgument(fmt.Errorf("--cached cannot be used with --stateless"))
		}
		return shared.ResolveProjectTargetForExecution(ctx, cmd, svc, query)
	}
	if cachedOnly {
		return shared.ResolveCachedProjectTargetArg(cmd, query)
	}
	return shared.ResolveProjectTargetArg(ctx, cmd, svc, query)
}

func newVersionsListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "list <project>", Short: "List Remote Config versions", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cached, _ := cmd.Flags().GetBool("cached")
		before, _ := cmd.Flags().GetString("before")
		if cmd.Flags().Changed("before") {
			if err := validateBeforeVersion(before); err != nil {
				return err
			}
		}
		project, err := resolveVersionProject(cmd, svc, args[0], cached)
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		if limit < 1 {
			return shared.InvalidArgument(fmt.Errorf("--limit must be greater than zero"))
		}
		all, _ := cmd.Flags().GetBool("all")
		since, err := timeFlag(cmd, "since")
		if err != nil {
			return err
		}
		until, err := timeFlag(cmd, "until")
		if err != nil {
			return err
		}
		ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
		if err != nil {
			return err
		}
		cmd.SetContext(ctx)
		result, err := svc.ListRemoteConfigVersions(ctx, project.ProjectID, core.VersionListOptions{Limit: limit, All: all, Before: before, Since: since, Until: until, CachedOnly: cached})
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			return shared.WriteJSON(cmd, versionListJSON(result))
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

func validateBeforeVersion(value string) error {
	if value == "" || value[0] < '1' || value[0] > '9' {
		return shared.InvalidArgument(fmt.Errorf("--before must be a canonical positive version number"))
	}
	for index := 1; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return shared.InvalidArgument(fmt.Errorf("--before must be a canonical positive version number"))
		}
	}
	return nil
}

type versionJSON struct {
	VersionNumber  string                    `json:"versionNumber,omitempty"`
	UpdateTime     string                    `json:"updateTime,omitempty" contract:"format=date-time"`
	UpdateUser     firebase.RemoteConfigUser `json:"updateUser,omitzero"`
	ChangeNote     *string                   `json:"change_note"`
	UpdateOrigin   string                    `json:"updateOrigin,omitempty"`
	UpdateType     string                    `json:"updateType,omitempty"`
	RollbackSource string                    `json:"rollbackSource,omitempty"`
	IsLegacy       bool                      `json:"isLegacy,omitempty"`
	Current        bool                      `json:"current"`
	Cached         bool                      `json:"cached"`
	CachedAt       time.Time                 `json:"cached_at,omitzero"`
	Size           int64                     `json:"size,omitempty"`
	Path           string                    `json:"path,omitempty"`
}

func versionListJSON(result core.RemoteConfigVersionList) []versionJSON {
	items := make([]versionJSON, 0, len(result.Versions))
	for _, entry := range result.Versions {
		items = append(items, versionEntryJSON(entry))
	}
	return items
}

func versionEntryJSON(entry core.RemoteConfigVersionEntry) versionJSON {
	return versionJSON{
		VersionNumber: entry.VersionNumber, UpdateTime: entry.UpdateTime, UpdateUser: entry.UpdateUser,
		ChangeNote: optionalVersionString(entry.ChangeNote), UpdateOrigin: entry.UpdateOrigin,
		UpdateType: entry.UpdateType, RollbackSource: entry.RollbackSource, IsLegacy: entry.IsLegacy,
		Current: entry.Current, Cached: entry.Cached, CachedAt: entry.CachedAt, Size: entry.Size, Path: entry.Path,
	}
}

func optionalVersionString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func timeFlag(cmd *cobra.Command, name string) (time.Time, error) {
	raw, _ := cmd.Flags().GetString(name)
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, shared.InvalidArgument(fmt.Errorf("invalid --%s time %q: use RFC3339", name, raw))
	}
	return value, nil
}

func newVersionsShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "show <project> <version>", Short: "Show Remote Config version metadata or JSON", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		cached, _ := cmd.Flags().GetBool("cached")
		project, err := resolveVersionProject(cmd, svc, args[0], cached)
		if err != nil {
			return err
		}
		ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
		if err != nil {
			return err
		}
		cmd.SetContext(ctx)
		resolved, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, args[1], cached)
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			entry := core.RemoteConfigVersionEntry{RemoteConfigVersion: resolved.Version, Cached: resolved.Cached}
			return shared.WriteJSON(cmd, versionShowResult{Project: project, Version: versionEntryJSON(entry), Cached: resolved.Cached})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nVersion: %s\nPublished: %s\nUpdated by: %s\nOrigin: %s\nType: %s\nChange note: %s\nRollback source: %s\nCached: %t\n", project.Name, project.ProjectID, resolved.Version.VersionNumber, resolved.Version.UpdateTime, resolved.Version.UpdateUser.Email, resolved.Version.UpdateOrigin, resolved.Version.UpdateType, resolved.Version.ChangeNote, resolved.Version.RollbackSource, resolved.Cached)
		return nil
	}}
	cmd.Flags().Bool("cached", false, "Require a local snapshot and do not contact Firebase")
	cmd.Flags().Bool("json", false, "Print metadata as JSON")
	return cmd
}

type versionDiffOptions struct {
	cached, json, sideBySide, parameters, conditions bool
	filters, groups                                  []string
	search, expr                                     string
}

func newVersionsDiffCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "diff <project> <from> [<to>]", Short: "Compare two Remote Config versions", Args: cobra.RangeArgs(2, 3), RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionsDiff(cmd, svc, args)
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
	cmd.Flags().Bool("side-by-side", false, "Print a two-column terminal diff")
	cmd.MarkFlagsMutuallyExclusive("parameters", "conditions")
	cmd.MarkFlagsMutuallyExclusive("json", "side-by-side")
}

func runVersionsDiff(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts := readVersionDiffOptions(cmd)
	project, err := resolveVersionProject(cmd, svc, args[0], opts.cached)
	if err != nil {
		return err
	}
	ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	to := "current"
	if len(args) == 3 {
		to = args[2]
	}
	fromCfg, toCfg, err := svc.GetRemoteConfigVersionPair(ctx, project.ProjectID, args[1], to, opts.cached)
	if err != nil {
		return err
	}
	result, err := filterVersionDiff(project, rcdiff.CompareRemoteConfigs(fromCfg.Config, toCfg.Config), fromCfg.Config, toCfg.Config, opts)
	if err != nil {
		return err
	}
	changed := result.HasChanges()
	if opts.json {
		if err := shared.WriteJSON(cmd, versionDiffResult{Project: project, FromVersion: fromCfg.Version.VersionNumber, ToVersion: toCfg.Version.VersionNumber, Changed: changed, Diff: result}); err != nil {
			return err
		}
		if changed {
			return shared.DiffFoundError(cmd)
		}
		return nil
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s): version %s → version %s\n", project.Name, project.ProjectID, fromCfg.Version.VersionNumber, toCfg.Version.VersionNumber); err != nil {
		return err
	}
	if opts.sideBySide {
		text, err := renderVersionSideBySide(
			result,
			shared.TerminalWidth(),
		)
		if err != nil {
			return err
		}
		if !changed {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
			return err
		}
		if _, err = fmt.Fprintln(cmd.OutOrStdout(), "\n"+renderVersionDiffSummary(result)+"\n\n"+text); err != nil {
			return err
		}
		return shared.DiffFoundError(cmd)
	}
	text, changed := rcdiff.RenderResult(result)
	if !changed {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
		return err
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), text); err != nil {
		return err
	}
	return shared.DiffFoundError(cmd)
}
func readVersionDiffOptions(cmd *cobra.Command) versionDiffOptions {
	o := versionDiffOptions{}
	o.cached, _ = cmd.Flags().GetBool("cached")
	o.json, _ = cmd.Flags().GetBool("json")
	o.sideBySide, _ = cmd.Flags().GetBool("side-by-side")
	o.parameters, _ = cmd.Flags().GetBool("parameters")
	o.conditions, _ = cmd.Flags().GetBool("conditions")
	o.filters, _ = cmd.Flags().GetStringArray("filter")
	o.groups, _ = cmd.Flags().GetStringArray("group")
	o.search, _ = cmd.Flags().GetString("search")
	o.expr, _ = cmd.Flags().GetString("expr")
	return o
}

func filterVersionDiff(project core.Project, result rcdiff.Result, from, to *firebase.RemoteConfig, opts versionDiffOptions) (rcdiff.Result, error) {
	if opts.parameters {
		result.Conditions = nil
	}
	if opts.conditions {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result, nil
	}
	filters := shared.ParseFilters(opts.filters)
	groupSet := map[string]bool{}
	for _, g := range opts.groups {
		groupSet[g] = true
	}
	compiledExpr, err := shared.CompileExpr(strings.TrimSpace(opts.expr), project.ProjectID)
	if err != nil {
		return rcdiff.Result{}, err
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
		match, err := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, change.Key, group)
		if err != nil {
			return rcdiff.Result{}, err
		}
		if !match {
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
	return result, nil
}

func newVersionsExportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "export <project> <version>", Short: "Export historical Remote Config JSON", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		cached, _ := cmd.Flags().GetBool("cached")
		project, err := resolveVersionProject(cmd, svc, args[0], cached)
		if err != nil {
			return err
		}
		to, _ := cmd.Flags().GetString("to")
		yes, _ := cmd.Flags().GetBool("yes")
		overwrite := false
		if to != "" {
			var proceed bool
			overwrite, proceed, err = shared.ConfirmFileOverwrite(cmd, to, yes)
			if err != nil || !proceed {
				return err
			}
		}
		ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
		if err != nil {
			return err
		}
		cmd.SetContext(ctx)
		resolved, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, args[1], cached)
		if err != nil {
			return err
		}
		body := rc.NormalizeExportBytes(resolved.Cache.RemoteConfig)
		if to == "" {
			if contract.Enabled(cmd) {
				target := project.ProjectID
				return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, nil, false))
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		}
		write := rc.CreateRemoteConfigFile
		if overwrite {
			write = rc.WriteRemoteConfigFile
		}
		if err := write(to, body); err != nil {
			return err
		}
		if contract.Enabled(cmd) {
			target, destination := project.ProjectID, to
			return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, &destination, overwrite))
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported version %s: %s\n", resolved.Version.VersionNumber, to)
		return nil
	}}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	cmd.Flags().Bool("cached", false, "Require a local snapshot and do not contact Firebase")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
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
	if restore {
		shared.AddChangeNoteFlag(cmd)
	}
	shared.AddYesFlag(cmd, "Skip final publish confirmation")
	if restore {
		shared.AddPlanOutFlag(cmd)
	}
	cmd.Flags().Bool("json", false, "Print result as JSON")
	return cmd
}

func runVersionPublish(cmd *cobra.Command, svc *core.Core, query, selector string, restore bool) error {
	ctx := shared.CommandContext(cmd)
	dry, _ := cmd.Flags().GetBool("dry-run")
	jsonOut, _ := cmd.Flags().GetBool("json")
	if dry {
		ctx = firebase.WithDryRun(ctx)
	}
	var changeNote *string
	var err error
	if restore {
		changeNote, err = shared.ReadChangeNoteFlag(cmd)
		if err != nil {
			return err
		}
		ctx, err = shared.WithChangeNote(ctx, changeNote)
		if err != nil {
			return err
		}
	}
	project, err := resolveVersionProject(cmd, svc, query, false)
	if err != nil {
		return err
	}
	ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		if hasDraft, draftErr := svc.HasDraft(project.ProjectID); draftErr != nil {
			return draftErr
		} else if hasDraft {
			return &shared.ConflictError{Code: "draft.exists", Resource: "draft", Target: project.ProjectID, Remediation: []shared.Remediation{
				{Description: "publish the existing draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "publish", project.ProjectID}},
				{Description: "discard the existing draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "discard", project.ProjectID}},
			}, Err: fmt.Errorf("project %s has an unpublished draft; publish or discard it before changing versions", project.ProjectID)}
		}
	}
	target, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, selector, restore)
	if err != nil {
		return err
	}
	current, err := svc.GetRemoteConfigVersion(ctx, project.ProjectID, "current", false)
	if err != nil {
		return err
	}
	if planPath, planning, planErr := shared.PlanOutputPath(cmd); planErr != nil {
		return planErr
	} else if planning {
		if current.Cache == nil || target.Cache == nil {
			return fmt.Errorf("version plan requires complete current and source snapshots")
		}
		currentDigest, digestErr := publication.RemoteConfigDigest(current.Cache.RemoteConfig)
		if digestErr != nil {
			return digestErr
		}
		targetDigest, digestErr := publication.RemoteConfigDigest(target.Cache.RemoteConfig)
		if digestErr != nil {
			return digestErr
		}
		action := publication.ActionNone
		validationSource := core.ValidationSourceLocal
		if currentDigest != targetDigest {
			action = publication.ActionPublish
			validationSource = core.ValidationSourceFirebase
			if validationErr := svc.ValidatePublicationCandidate(ctx, project.ProjectID, current.Cache.RemoteConfig, target.Cache.RemoteConfig, current.Cache.ETag, "versions-restore"); validationErr != nil {
				return validationErr
			}
		}
		environment, environmentErr := core.PublicationEnvironmentForContext(ctx)
		if environmentErr != nil {
			return environmentErr
		}
		parsedTarget, targetErr := rctarget.Parse(project.ProjectID)
		if targetErr != nil {
			return targetErr
		}
		publicationPlan := publication.New(cmd.Root().Version, "versions.restore", environment.Policy, changeNote)
		publicationPlan.Execution.HooksEnabled = environment.HooksEnabled
		publicationPlan.Execution.HookDefinitionSHA256 = environment.HookDefinitionSHA256
		publicationPlan.Operation.Selection, targetErr = json.Marshal(struct {
			RequestedVersion string `json:"requested_version"`
			SourceVersion    string `json:"source_version"`
		}{RequestedVersion: selector, SourceVersion: target.Version.VersionNumber})
		if targetErr != nil {
			return targetErr
		}
		publicationPlan.Targets = append(publicationPlan.Targets, publication.Target{
			Target: project.ProjectID, ProjectID: parsedTarget.ProjectID, Template: string(parsedTarget.Kind), Action: action, ChangeNote: changeNote,
			Base:       publication.Snapshot{Version: current.Version.VersionNumber, ETag: current.Cache.ETag, RemoteConfig: current.Cache.RemoteConfig},
			Candidate:  publication.Snapshot{Version: target.Version.VersionNumber, RemoteConfig: target.Cache.RemoteConfig},
			Validation: publication.Validation{Source: validationSource, ValidatedAt: publicationPlan.CreatedAt},
			Source:     publication.Source{Kind: "version_restore", Fingerprint: target.Version.VersionNumber},
		})
		_, writeErr := rc.WritePublicationPlan(cmd, publicationPlan, planPath)
		return writeErr
	}
	if current.Version.VersionNumber == target.Version.VersionNumber {
		if jsonOut {
			return shared.WriteJSON(cmd, versionPublishJSON(project.ProjectID, restore, current.Version.VersionNumber, target.Version.VersionNumber, "", changeNote, dry, false, true, core.ValidationSourceLocal))
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "version %s is already current; no operation performed\nvalidated: true · validation_source: local\n", target.Version.VersionNumber)
		return err
	}
	diffText, changed := rc.RenderRemoteConfigDiff(current.Config, target.Config)
	if !changed {
		if jsonOut {
			return shared.WriteJSON(cmd, versionPublishJSON(project.ProjectID, restore, current.Version.VersionNumber, target.Version.VersionNumber, "", changeNote, dry, false, true, core.ValidationSourceLocal))
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences\nvalidated: true · validation_source: local")
		return err
	}
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
		if err := shared.RequireYesInMachineMode(cmd, yes, "publishing the version change to "+project.ProjectID, true); err != nil {
			return err
		}
		confirm := shared.NewConfirmation(fmt.Sprintf("Publish this %s to %s?", map[bool]string{true: "restore", false: "rollback"}[restore], project.ProjectID), shared.ConfirmationOptions{Destructive: true})
		confirm.Output = cmd.ErrOrStderr()
		ok, err := confirm.RunPrompt()
		if err != nil || !ok {
			return err
		}
	}
	action := "Rolling back Remote Config for "
	if restore {
		action = "Restoring Remote Config for "
	}
	if dry {
		action = "Previewing version change for "
	}
	progress.Start(action + project.ProjectID + "…")
	if dry {
		if restore {
			_, err = svc.RestoreRemoteConfigVersion(ctx, project.ProjectID, target.Version.VersionNumber)
		} else {
			_, err = svc.RollbackRemoteConfig(ctx, project.ProjectID, target.Version.VersionNumber)
		}
		if err != nil {
			validated := false
			validationSource := core.ValidationSourceLocal
			status, stage := "failed", "preparation"
			if source, ok := core.RemoteConfigValidationSource(err); ok {
				validationSource = source
				status, stage = "validation-failed", "validation"
			}
			if jsonOut {
				result := versionPublishJSON(project.ProjectID, restore, current.Version.VersionNumber, target.Version.VersionNumber, "", changeNote, true, true, validated, validationSource)
				result.Status = status
				result.Error = &versionOperationError{Stage: stage, Message: shared.SafeErrorText(err)}
				if writeErr := shared.WriteJSON(cmd, result); writeErr != nil {
					return writeErr
				}
			} else {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "validated: false · validation_source: %s\n", validationSource)
			}
			return err
		}
		result := versionPublishJSON(project.ProjectID, restore, current.Version.VersionNumber, target.Version.VersionNumber, "", changeNote, true, true, true, core.ValidationSourceFirebase)
		if jsonOut {
			return shared.WriteJSON(cmd, result)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧪 dry run: %s would use version %s\nvalidated: true · validation_source: firebase\n", project.ProjectID, target.Version.VersionNumber)
		return nil
	}
	latest, err := svc.GetRemoteConfigVersion(shared.CommandContext(cmd), project.ProjectID, "current", false)
	if err != nil {
		return err
	}
	if latest.Version.VersionNumber != current.Version.VersionNumber {
		return &shared.ConflictError{Code: "remote_config.conflict", Resource: "remote_config", Target: project.ProjectID, Retryable: true, Err: fmt.Errorf("remote config changed from version %s to %s during preview; rerun the command", current.Version.VersionNumber, latest.Version.VersionNumber)}
	}
	var result core.VersionPublishResult
	if restore {
		result, err = svc.RestoreRemoteConfigVersion(ctx, project.ProjectID, target.Version.VersionNumber)
	} else {
		result, err = svc.RollbackRemoteConfig(ctx, project.ProjectID, target.Version.VersionNumber)
	}
	publishErr := err
	if publishErr != nil && result.PublishedVersion == "" {
		if !restore && target.Cached {
			return fmt.Errorf("%w; if Firebase no longer retains version %s, republish the local snapshot with: fbrcm versions restore %s %s", publishErr, target.Version.VersionNumber, project.ProjectID, target.Version.VersionNumber)
		}
		return publishErr
	}
	payload := versionPublishJSON(project.ProjectID, restore, result.PreviousVersion, result.SourceVersion, result.PublishedVersion, changeNote, false, true, true, core.ValidationSourceFirebase)
	if publishErr != nil {
		status := "published-local-update-failed"
		var hookErr *core.RemoteConfigPublishedHookError
		var cacheErr *core.RemoteConfigPublishedCacheError
		switch {
		case errors.As(publishErr, &hookErr):
			status = "published-hook-failed"
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.post_publish_hook_failed", Message: "Firebase accepted the version publication, but a post_publish hook failed.", Target: project.ProjectID, Details: struct {
				Stage string `json:"stage"`
			}{Stage: "post_publish_hook"}, Remediation: []shared.Remediation{{Description: "inspect hook trust and status without republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"}}}})
		case errors.As(publishErr, &cacheErr):
			status = "published-cache-failed"
			filter, _ := rctarget.ExactFilter(project.ProjectID)
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.cache_stale", Message: "Firebase accepted the version publication, but the local cache update failed.", Target: project.ProjectID, Details: struct {
				Stage string `json:"stage"`
			}{Stage: "cache"}, Remediation: []shared.Remediation{{Description: "refresh the published target instead of republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", filter}}}})
		}
		payload.Status = status
		payload.Error = &versionOperationError{Stage: map[bool]string{true: "post_publish_hook", false: "cache"}[status == "published-hook-failed"], Message: shared.SafeErrorText(publishErr)}
	}
	if jsonOut {
		if writeErr := shared.WriteJSON(cmd, payload); writeErr != nil {
			return writeErr
		}
		return publishErr
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s v%s → v%s, using v%s\n", map[bool]string{true: "♻️ restored", false: "⏪ rolled back"}[restore], project.ProjectID, result.PreviousVersion, result.PublishedVersion, result.SourceVersion)
	if publishErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Published with warning: %v\n", publishErr)
	}
	return publishErr
}

func versionPublishJSON(projectID string, restore bool, previousVersion, sourceVersion, publishedVersion string, changeNote *string, dryRun, changed, validated bool, validationSource string) versionPublishResult {
	var published *string
	if publishedVersion != "" {
		published = &publishedVersion
	}
	status := "unchanged"
	if changed {
		status = map[bool]string{true: "would-publish", false: "published"}[dryRun]
	}
	return versionPublishResult{ProjectID: projectID, Operation: map[bool]string{true: "restore", false: "rollback"}[restore], PreviousVersion: previousVersion, SourceVersion: sourceVersion, PublishedVersion: published, DryRun: dryRun, Changed: changed, Validated: validated, ValidationSource: validationSource, ChangeNote: changeNote, Status: status}
}

type versionShowResult struct {
	Project core.Project `json:"project"`
	Version versionJSON  `json:"version"`
	Cached  bool         `json:"cached"`
}

type versionDiffResult struct {
	Project     core.Project  `json:"project"`
	FromVersion string        `json:"from_version"`
	ToVersion   string        `json:"to_version"`
	Changed     bool          `json:"changed"`
	Diff        rcdiff.Result `json:"diff"`
}

type versionOperationError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

type versionPublishResult struct {
	ProjectID        string                 `json:"project_id"`
	Operation        string                 `json:"operation" contract:"enum=restore|rollback"`
	Status           string                 `json:"status" contract:"enum=unchanged|would-publish|published|failed|validation-failed|published-local-update-failed|published-hook-failed|published-cache-failed"`
	PreviousVersion  string                 `json:"previous_version"`
	SourceVersion    string                 `json:"source_version"`
	PublishedVersion *string                `json:"published_version"`
	DryRun           bool                   `json:"dry_run"`
	Changed          bool                   `json:"changed"`
	Validated        bool                   `json:"validated"`
	ValidationSource string                 `json:"validation_source"`
	ChangeNote       *string                `json:"change_note"`
	Error            *versionOperationError `json:"error"`
}

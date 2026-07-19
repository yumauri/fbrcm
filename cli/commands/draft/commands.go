package draft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "draft", Short: "Inspect, publish, and discard Remote Config drafts"}
	cmd.AddCommand(newListCommand(), newPathCommand(), newShowCommand(), newDiffCommand(svc), newPublishCommand(svc), newDiscardCommand())
	return cmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print Remote Config drafts directory path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.GetDraftsDirPath()
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]string{"path": path})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print path as JSON")
	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "list", Short: "List local Remote Config drafts", Args: cobra.NoArgs, RunE: runList}
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().Bool("json", false, "Print drafts as JSON")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	filters, _ := cmd.Flags().GetStringArray("filter")
	items, err := loadItems(filters)
	if err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		return shared.WriteJSON(cmd, items)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderList(items))
	return nil
}

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "show <project>", Short: "Show a local Remote Config draft", Args: cobra.ExactArgs(1), RunE: runShow}
	cmd.Flags().Bool("raw", false, "Print the exact stored draft envelope without parsing")
	cmd.Flags().String("to", "", "Write output to file path")
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	projectID, _, err := resolveDraft(args[0])
	if err != nil {
		return err
	}
	rawOut, _ := cmd.Flags().GetBool("raw")
	var body []byte
	if rawOut {
		body, err = os.ReadFile(config.GetDraftPath(projectID))
	} else {
		stored, loadErr := config.LoadDraft(projectID)
		if loadErr != nil {
			return loadErr
		}
		cfg, parseErr := firebase.ParseRemoteConfig(stored.RemoteConfig)
		if parseErr != nil {
			return parseErr
		}
		body, err = firebase.MarshalRemoteConfig(cfg)
	}
	if err != nil {
		return err
	}
	if !rawOut {
		body = rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(body))
	}
	to, _ := cmd.Flags().GetString("to")
	if to == "" {
		_, err = cmd.OutOrStdout().Write(body)
		return err
	}
	if dir := filepath.Dir(to); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create destination dir: %w", err)
		}
	}
	if err := os.WriteFile(to, body, 0o600); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}
	if err := os.Chmod(to, 0o600); err != nil {
		return fmt.Errorf("set destination permissions: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported draft: %s\n", to)
	return nil
}

type diffOptions struct {
	against                              string
	cached, json, parameters, conditions bool
	filters, groups                      []string
	search, expr                         string
}

func newDiffCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "diff <project>", Short: "Compare a draft with its base or current Remote Config", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return shared.DiffCommandError(cmd, runDiff(cmd, svc, args[0]))
	}}
	cmd.Flags().String("against", "base", "Comparison target: base or current")
	cmd.Flags().Bool("cached", false, "Use the latest local snapshot and do not contact Firebase")
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().StringArray("group", nil, "Select parameters in group; may be repeated")
	cmd.Flags().String("expr", "", "Filter parameter changes by expr-lang expression")
	cmd.Flags().Bool("parameters", false, "Include only parameter and group description changes")
	cmd.Flags().Bool("conditions", false, "Include only condition changes")
	cmd.Flags().Bool("json", false, "Print diff as JSON")
	shared.AddDiffExitCodeFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("parameters", "conditions")
	return cmd
}

func runDiff(cmd *cobra.Command, svc *core.Core, query string) error {
	projectID, project, err := resolveDraft(query)
	if err != nil {
		return err
	}
	stored, err := config.LoadDraft(projectID)
	if err != nil {
		return err
	}
	opts := readDiffOptions(cmd)
	var fromRaw, toRaw json.RawMessage
	fromRaw, toRaw = stored.BaseRemoteConfig, stored.RemoteConfig
	currentVersion := ""
	if opts.against == "current" {
		var cache *core.ParametersCache
		if opts.cached {
			cache, _, err = svc.InspectParametersCache(projectID)
			if err == nil && cache == nil {
				err = fmt.Errorf("parameters cache not found")
			}
		} else {
			cache, _, err = svc.GetParameters(context.Background(), projectID, true)
		}
		if err != nil {
			return err
		}
		var changed bool
		toRaw, changed, err = core.MergeDraftWithLatest(stored.BaseRemoteConfig, stored.RemoteConfig, cache.RemoteConfig)
		if err != nil {
			return err
		}
		if !changed {
			toRaw = cache.RemoteConfig
		}
		fromRaw = cache.RemoteConfig
		currentCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
		currentVersion = currentCfg.Version.VersionNumber
	} else if opts.against != "base" {
		return fmt.Errorf("--against must be base or current")
	} else if opts.cached {
		return fmt.Errorf("--cached requires --against current")
	}
	fromCfg, err := firebase.ParseRemoteConfig(fromRaw)
	if err != nil {
		return err
	}
	toCfg, err := firebase.ParseRemoteConfig(toRaw)
	if err != nil {
		return err
	}
	result := filterDiff(project, rcdiff.CompareRemoteConfigs(fromCfg, toCfg), fromCfg, toCfg, opts)
	if opts.json {
		if err := shared.WriteJSON(cmd, map[string]any{"project": project, "against": opts.against, "base_version": stored.BaseVersion, "current_version": currentVersion, "changed": result.HasChanges(), "diff": result}); err != nil {
			return err
		}
		if result.HasChanges() {
			return shared.DiffFoundError(cmd)
		}
		return nil
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s): %s → draft\n", project.Name, projectID, opts.against); err != nil {
		return err
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

func newPublishCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "publish [project...]", Short: "Safely rebase and publish Remote Config drafts", Args: cobra.ArbitraryArgs, RunE: func(cmd *cobra.Command, args []string) error {
		return runPublish(cmd, svc, args)
	}}
	cmd.Flags().Bool("all", false, "Publish every valid draft in the active profile")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Skip publish confirmations")
	cmd.Flags().Bool("json", false, "Print results as JSON")
	return cmd
}

type publishResult struct {
	ProjectID        string `json:"project_id"`
	Status           string `json:"status"`
	BaseVersion      string `json:"base_version,omitempty"`
	PreviousVersion  string `json:"previous_version,omitempty"`
	PublishedVersion string `json:"published_version,omitempty"`
	Rebased          bool   `json:"rebased"`
	Changed          bool   `json:"changed"`
	DraftDeleted     bool   `json:"draft_deleted"`
	DryRun           bool   `json:"dry_run"`
	Error            string `json:"error,omitempty"`
}

func runPublish(cmd *cobra.Command, svc *core.Core, args []string) error {
	ids, err := selectedDraftIDs(cmd, args)
	if err != nil {
		return err
	}
	dry, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	jsonOut, _ := cmd.Flags().GetBool("json")
	ctx := context.Background()
	if dry {
		ctx = firebase.WithDryRun(ctx)
	}
	results := make([]publishResult, 0, len(ids))
	failed := false
	for _, projectID := range ids {
		result := publishResult{ProjectID: projectID, DryRun: dry}
		plan, prepareErr := svc.PrepareDraftPublish(ctx, projectID)
		if prepareErr != nil {
			result.Status, result.Error = "failed", prepareErr.Error()
			results = append(results, result)
			failed = true
			continue
		}
		result.BaseVersion = plan.Draft.BaseVersion
		result.Rebased = plan.Rebased
		result.Changed = plan.HasChanges
		latestCfg, _ := firebase.ParseRemoteConfig(plan.Latest.RemoteConfig)
		result.PreviousVersion = latestCfg.Version.VersionNumber
		if plan.HasChanges && !jsonOut {
			fromCfg, _ := firebase.ParseRemoteConfig(plan.Latest.RemoteConfig)
			toCfg, _ := firebase.ParseRemoteConfig(plan.Candidate)
			diffText, _ := rcdiff.RenderRemoteConfigDiff(fromCfg, toCfg)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Publish draft for %s\n%s\n", projectID, diffText)
		}
		if plan.HasChanges && !dry && !yes {
			confirm := shared.NewConfirmation(fmt.Sprintf("Publish this draft to %s?", projectID), shared.ConfirmationOptions{})
			confirm.Output = cmd.ErrOrStderr()
			ok, confirmErr := confirm.RunPrompt()
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				result.Status = "canceled"
				results = append(results, result)
				continue
			}
		}
		cache, _, publishErr := svc.ExecuteDraftPublish(ctx, projectID, plan)
		if publishErr != nil {
			var cleanupErr *core.DraftPublishedCleanupError
			if errors.As(publishErr, &cleanupErr) && cache != nil {
				publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
				result.Status = "published-cleanup-failed"
				result.PublishedVersion = publishedCfg.Version.VersionNumber
				result.Error = publishErr.Error()
				results = append(results, result)
				failed = true
				continue
			}
			result.Status, result.Error = "failed", publishErr.Error()
			results = append(results, result)
			failed = true
			continue
		}
		if dry {
			result.Status = "would-publish"
		} else if !plan.HasChanges {
			result.Status, result.DraftDeleted = "already-applied", true
		} else {
			publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
			result.Status, result.PublishedVersion, result.DraftDeleted = "published", publishedCfg.Version.VersionNumber, true
		}
		results = append(results, result)
	}
	if jsonOut {
		if err := shared.WriteJSON(cmd, map[string]any{"results": results}); err != nil {
			return err
		}
	} else {
		for _, result := range results {
			if result.Error != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s\n", result.ProjectID, result.Error)
				continue
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", result.Status, result.ProjectID)
		}
	}
	if failed {
		return fmt.Errorf("one or more drafts failed")
	}
	return nil
}

func newDiscardCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "discard [project...]", Short: "Discard local Remote Config drafts", Args: cobra.ArbitraryArgs, RunE: runDiscard}
	cmd.Flags().Bool("all", false, "Discard every draft in the active profile")
	shared.AddYesFlag(cmd, "Skip destructive confirmations")
	cmd.Flags().Bool("json", false, "Print results as JSON")
	return cmd
}

func runDiscard(cmd *cobra.Command, args []string) error {
	ids, err := selectedDraftIDs(cmd, args)
	if err != nil {
		return err
	}
	yes, _ := cmd.Flags().GetBool("yes")
	jsonOut, _ := cmd.Flags().GetBool("json")
	results := make([]map[string]any, 0, len(ids))
	for _, projectID := range ids {
		stored, loadErr := config.LoadDraft(projectID)
		if !jsonOut {
			if loadErr != nil {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Draft preview unavailable:", loadErr)
			} else {
				baseCfg, baseErr := firebase.ParseRemoteConfig(stored.BaseRemoteConfig)
				draftCfg, draftErr := firebase.ParseRemoteConfig(stored.RemoteConfig)
				if baseErr == nil && draftErr == nil {
					if text, changed := rcdiff.RenderRemoteConfigDiff(baseCfg, draftCfg); changed {
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), text)
					}
				}
			}
		}
		if !yes {
			confirm := shared.NewConfirmation(fmt.Sprintf("Discard draft for %s?", projectID), shared.ConfirmationOptions{Destructive: true})
			confirm.Output = cmd.ErrOrStderr()
			ok, confirmErr := confirm.RunPrompt()
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				results = append(results, map[string]any{"project_id": projectID, "status": "canceled"})
				continue
			}
		}
		if err := config.DeleteDraft(projectID); err != nil {
			return err
		}
		baseVersion := ""
		if stored != nil {
			baseVersion = stored.BaseVersion
		}
		results = append(results, map[string]any{"project_id": projectID, "status": "discarded", "base_version": baseVersion})
	}
	if jsonOut {
		return shared.WriteJSON(cmd, map[string]any{"results": results})
	}
	for _, result := range results {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", result["status"], result["project_id"])
	}
	return nil
}

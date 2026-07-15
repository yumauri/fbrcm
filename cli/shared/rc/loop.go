package rc

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// RemoteMutationTotals counts the projects and parameters changed by a publish loop.
type RemoteMutationTotals struct {
	ModifiedProjects int
	ChangedParams    int
}

// RunRemoteDraftLoop revalidates each project, applies mutations on top of any
// existing draft, and stores the result locally without a Firebase write.
func RunRemoteDraftLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, operation string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	var totals RemoteMutationTotals
	for _, project := range projects {
		cfg, err := RevalidateProjectConfig(ctx, svc, project)
		if err != nil {
			return totals, err
		}
		if draftRaw, hasDraft, err := svc.LoadDraft(project.ProjectID); err != nil {
			return totals, err
		} else if hasDraft {
			draftCfg, err := firebase.ParseRemoteConfig(draftRaw)
			if err != nil {
				return totals, err
			}
			cfg.Config = draftCfg
		}
		mutate, err := plan(project, cfg)
		if err != nil {
			return totals, err
		}
		if mutate == nil {
			continue
		}
		changedCount, finalCfg, err := mutate(cfg.Config)
		if err != nil {
			return totals, err
		}
		if changedCount == 0 {
			continue
		}
		finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
		if err != nil {
			return totals, err
		}
		if !firebase.IsDryRun(ctx) {
			if err := svc.SaveDraft(project.ProjectID, finalRaw); err != nil {
				return totals, err
			}
		}
		totals.ModifiedProjects++
		totals.ChangedParams += changedCount
		status := "drafted"
		if firebase.IsDryRun(ctx) {
			status = "would draft"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📝 %s: %s\n", status, project.ProjectID)
	}
	return totals, nil
}

// RemoteMutationPlanner builds the per-project mutation from a freshly revalidated
// config. Returning a nil mutation leaves the project untouched.
type RemoteMutationPlanner func(project core.Project, cfg *ProjectConfig) (RemoteConfigMutation, error)

// RunRemotePublishLoop revalidates, mutates, and publishes each project in turn,
// retrying a project whenever the publish reports a stale ETag. It prints
// "<emoji> published: <id>" for every project that changes and returns the totals.
func RunRemotePublishLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, operation, publishedEmoji string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	var totals RemoteMutationTotals
	for _, project := range projects {
		hasDraft, err := svc.HasDraft(project.ProjectID)
		if err != nil {
			return totals, err
		}
		if hasDraft {
			return totals, fmt.Errorf("project %s has an unpublished draft; use --draft, publish it, or discard it first", project.ProjectID)
		}
	}
	for _, project := range projects {
		for {
			cfg, err := RevalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return totals, err
			}

			mutate, err := plan(project, cfg)
			if err != nil {
				return totals, err
			}
			if mutate == nil {
				break
			}

			changedCount, retry, err := PublishProjectConfigMutation(ctx, svc, cfg, operation, cmd.ErrOrStderr(), mutate)
			if err != nil {
				return totals, err
			}
			if changedCount == 0 {
				break
			}
			if retry {
				continue
			}

			totals.ModifiedProjects++
			totals.ChangedParams += changedCount
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s published: %s\n", publishedEmoji, project.ProjectID)
			break
		}
	}
	return totals, nil
}

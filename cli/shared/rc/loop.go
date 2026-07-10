package rc

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
)

// RemoteMutationTotals counts the projects and parameters changed by a publish loop.
type RemoteMutationTotals struct {
	ModifiedProjects int
	ChangedParams    int
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

package cache

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cached Remote Config snapshots",
	}
	cacheCmd.AddCommand(newPathCommand(), newClearCommand(), newListCommand())
	return cacheCmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print cached Remote Config snapshots directory path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetParametersCacheDirPath()
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

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cached Remote Config snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			entries, err := loadCacheEntries()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 Nothing to clear")
				return nil
			}
			var snapshotCount int
			var snapshotSize int64
			projects := map[string]struct{}{}
			for _, entry := range entries {
				snapshotCount++
				snapshotSize += entry.Size
				projects[entry.ProjectID] = struct{}{}
			}
			deleteCaches := snapshotCount > 0
			if deleteCaches && !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete %d cached Remote Config versions (%s) across %d projects?", snapshotCount, strings.TrimSpace(humanSize(snapshotSize)), len(projects)),
					shared.ConfirmationOptions{Destructive: true, Notes: []shared.ConfirmationNote{{Text: "Local versions no longer retained by Firebase may be permanently lost."}}},
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				deleteCaches = ok
			}
			if deleteCaches {
				if err := config.ClearParametersCache(); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 cleared caches: %s\n", config.GetParametersCacheDirPath())
			}

			return nil
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cached Remote Config snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			entries, err := loadCacheEntries()
			if err != nil {
				return err
			}

			if jsonOut {
				if err := shared.WriteJSON(cmd, entries); err != nil {
					return err
				}
				logCacheTotal(entries)
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderCacheTable(entries))
			logCacheTotal(entries)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print cache entries as JSON")
	return cmd
}

package cache

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func New() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cached Remote Config snapshots",
	}
	cacheCmd.AddCommand(newPathCommand(), newClearCommand(), newListCommand())
	contract.MustRegisterResponsePath(cacheCmd, "path", shared.PathResult{})
	contract.MustRegisterResponsePath(cacheCmd, "clear", cacheClearResult{})
	contract.MustRegisterResponsePath(cacheCmd, "list", []cacheEntry{})
	return cacheCmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print cached Remote Config snapshots directory path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetParametersCacheDirPath()
			if jsonOut {
				return shared.WriteJSON(cmd, shared.PathResult{Path: path})
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
		Args:  cobra.NoArgs,
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
				if contract.Enabled(cmd) {
					return shared.WriteJSON(cmd, cacheClearResult{Path: config.GetParametersCacheDirPath(), Status: "unchanged", SnapshotsDeleted: 0, TargetsAffected: 0, BytesDeleted: 0})
				}
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
				if err := shared.RequireYesInMachineMode(cmd, yes, "clearing cached Remote Config snapshots", true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(
					fmt.Sprintf(
						"Delete %s (%s) across %s?",
						rcdisplay.FormatCount(snapshotCount, "cached Remote Config version", "cached Remote Config versions"),
						strings.TrimSpace(humanSize(snapshotSize)),
						rcdisplay.FormatCount(len(projects), "template target", "template targets"),
					),
					shared.ConfirmationOptions{Destructive: true, Notes: []shared.ConfirmationNote{{Text: "Local versions no longer retained by Firebase may be permanently lost."}}},
				)
				confirm.Input = cmd.InOrStdin()
				confirm.Output = cmd.ErrOrStderr()
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
				if contract.Enabled(cmd) {
					return shared.WriteJSON(cmd, cacheClearResult{Path: config.GetParametersCacheDirPath(), Status: "cleared", SnapshotsDeleted: snapshotCount, TargetsAffected: len(projects), BytesDeleted: snapshotSize})
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 cleared caches: %s\n", config.GetParametersCacheDirPath())
			}

			return nil
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

type cacheClearResult struct {
	Path             string `json:"path"`
	Status           string `json:"status" contract:"enum=cleared|unchanged"`
	SnapshotsDeleted int    `json:"snapshots_deleted"`
	TargetsAffected  int    `json:"targets_affected"`
	BytesDeleted     int64  `json:"bytes_deleted"`
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cached Remote Config snapshots",
		Args:  cobra.NoArgs,
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

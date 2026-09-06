package cache

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

func New() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage local caches",
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
		Short: "Print a cache directory path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			kind, err := cacheKind(cmd)
			if err != nil {
				return err
			}
			path := config.GetCacheDirPath()
			switch kind {
			case "remote-config":
				path = config.GetParametersCacheDirPath()
			case "apps":
				path = config.GetAppsCacheDirPath()
			}
			if jsonOut {
				return shared.WriteJSON(cmd, shared.PathResult{Path: path})
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().String("kind", "remote-config", "Cache kind: all, remote-config, or apps")
	cmd.Flags().Bool("json", false, "Print path as JSON")
	return cmd
}

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear local cache entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			kind, err := cacheKind(cmd)
			if err != nil {
				return err
			}
			entries, err := loadCacheEntries(kind)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				if contract.Enabled(cmd) {
					return shared.WriteJSON(cmd, cacheClearResult{Path: cachePath(kind), Kind: kind, Status: "unchanged", EntriesDeleted: 0, TargetsAffected: 0, BytesDeleted: 0})
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 Nothing to clear")
				return nil
			}
			var entryCount int
			var entrySize int64
			projects := map[string]struct{}{}
			for _, entry := range entries {
				entryCount++
				entrySize += entry.Size
				projects[entry.ProjectID] = struct{}{}
			}
			deleteCaches := entryCount > 0
			if deleteCaches && !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "clearing local cache entries", true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(
					fmt.Sprintf(
						"Delete %s (%s) across %s?",
						rcdisplay.FormatCount(entryCount, "cache entry", "cache entries"),
						strings.TrimSpace(humanSize(entrySize)),
						rcdisplay.FormatCount(len(projects), "project", "projects"),
					),
					shared.ConfirmationOptions{Destructive: true},
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
				if err := clearCacheKind(kind); err != nil {
					return err
				}
				if contract.Enabled(cmd) {
					return shared.WriteJSON(cmd, cacheClearResult{Path: cachePath(kind), Kind: kind, Status: "cleared", EntriesDeleted: entryCount, TargetsAffected: len(projects), BytesDeleted: entrySize})
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 cleared caches: %s\n", cachePath(kind))
			}

			return nil
		},
	}
	cmd.Flags().String("kind", "all", "Cache kind: all, remote-config, or apps")
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

type cacheClearResult struct {
	Path            string `json:"path"`
	Kind            string `json:"kind" contract:"enum=all|remote-config|apps"`
	Status          string `json:"status" contract:"enum=cleared|unchanged"`
	EntriesDeleted  int    `json:"entries_deleted"`
	TargetsAffected int    `json:"targets_affected"`
	BytesDeleted    int64  `json:"bytes_deleted"`
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local cache entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			kind, err := cacheKind(cmd)
			if err != nil {
				return err
			}
			entries, err := loadCacheEntries(kind)
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
	cmd.Flags().String("kind", "all", "Cache kind: all, remote-config, or apps")
	cmd.Flags().Bool("json", false, "Print cache entries as JSON")
	return cmd
}

func cacheKind(cmd *cobra.Command) (string, error) {
	kind, err := cmd.Flags().GetString("kind")
	if err != nil {
		return "", err
	}
	switch kind {
	case "all", "remote-config", "apps":
		return kind, nil
	default:
		return "", shared.InvalidArgument(fmt.Errorf("--kind must be all, remote-config, or apps"))
	}
}

func cachePath(kind string) string {
	switch kind {
	case "remote-config":
		return config.GetParametersCacheDirPath()
	case "apps":
		return config.GetAppsCacheDirPath()
	default:
		return config.GetCacheDirPath()
	}
}

func clearCacheKind(kind string) error {
	if kind == "all" || kind == "remote-config" {
		if err := config.ClearParametersCache(); err != nil {
			return err
		}
	}
	if kind == "all" || kind == "apps" {
		return config.ClearAppsCache()
	}
	return nil
}

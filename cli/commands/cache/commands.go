package cache

import (
	"encoding/json"
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cached parameters files",
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print parameters cache directory path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetCacheDirPath()
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(map[string]string{"path": path})
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	pathCmd.Flags().Bool("json", false, "Print path as JSON")

	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete cached parameters files",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			deleteCaches := true
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete cached parameters files in %s?", config.GetParametersCacheDirPath()),
					confirmation.Yes,
					shared.ConfirmationOptions{Destructive: true},
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				deleteCaches = ok
			}
			if deleteCaches {
				if err := config.PurgeParametersCache(); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged caches: %s\n", config.GetParametersCacheDirPath())
			}

			draftIDs, err := config.ListDraftProjectIDs()
			if err != nil {
				return err
			}
			if len(draftIDs) > 0 {
				deleteDrafts := true
				if !yes {
					confirm := shared.NewConfirmation(
						fmt.Sprintf("Delete draft files in %s?", config.GetDraftsDirPath()),
						confirmation.No,
						shared.ConfirmationOptions{Destructive: true},
					)
					ok, err := confirm.RunPrompt()
					if err != nil {
						return err
					}
					deleteDrafts = ok
				}
				if deleteDrafts {
					if err := config.PurgeDrafts(); err != nil {
						return err
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged drafts: %s\n", config.GetDraftsDirPath())
				}
			}

			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List cached parameters files",
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
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(entries); err != nil {
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
	listCmd.Flags().Bool("json", false, "Print cache entries as JSON")

	cacheCmd.AddCommand(pathCmd, purgeCmd, listCmd)
	return cacheCmd
}

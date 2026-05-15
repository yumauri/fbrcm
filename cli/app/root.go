package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	addcmd "fbrcm/cli/commands/add"
	cachecmd "fbrcm/cli/commands/cache"
	deletecmd "fbrcm/cli/commands/delete"
	getcmd "fbrcm/cli/commands/get"
	logincmd "fbrcm/cli/commands/login"
	profilecmd "fbrcm/cli/commands/profile"
	projectcmd "fbrcm/cli/commands/project"
	projectscmd "fbrcm/cli/commands/projects"
	updatecmd "fbrcm/cli/commands/update"
	"fbrcm/core"
	"fbrcm/core/config"
	corelog "fbrcm/core/log"
)

var rootCmd = &cobra.Command{
	Use:   "fbrcm",
	Short: "Firebase project viewer",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if isProfileCommand(cmd) || cmd.Name() == "help" {
			return nil
		}
		if err := config.EnsureActiveProfile(); err != nil {
			return fmt.Errorf("ensure active profile: %w", err)
		}
		return nil
	},
}

const versionTemplate = `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s\n" .Version}}`

// isProfileCommand reports is profile command and returns the resulting value or error.
func isProfileCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "profile" || strings.HasPrefix(cmd.CommandPath(), "fbrcm profile")
}

// Execute handles execute and returns the resulting value or error.
func Execute(s *core.Core, version, commit, date string) {
	corelog.For("cli").Debug("register cli commands")
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	rootCmd.SetVersionTemplate(versionTemplate)

	rootCmd.AddCommand(addcmd.New(s))
	rootCmd.AddCommand(cachecmd.New())
	rootCmd.AddCommand(deletecmd.New(s))
	rootCmd.AddCommand(getcmd.New(s))
	rootCmd.AddCommand(logincmd.New(s))
	rootCmd.AddCommand(profilecmd.New())
	rootCmd.AddCommand(projectcmd.New(s))
	rootCmd.AddCommand(projectscmd.New(s))
	rootCmd.AddCommand(updatecmd.New(s))

	if err := rootCmd.Execute(); err != nil {
		corelog.For("cli").Error("cli command failed", "err", err)
		os.Exit(1)
	}
}

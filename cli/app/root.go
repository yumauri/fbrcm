package app

import (
	"os"

	"github.com/spf13/cobra"

	addcmd "fbrcm/cli/commands/add"
	cachecmd "fbrcm/cli/commands/cache"
	deletecmd "fbrcm/cli/commands/delete"
	getcmd "fbrcm/cli/commands/get"
	logincmd "fbrcm/cli/commands/login"
	projectcmd "fbrcm/cli/commands/project"
	projectscmd "fbrcm/cli/commands/projects"
	"fbrcm/core"
	corelog "fbrcm/core/log"
)

var rootCmd = &cobra.Command{
	Use:   "fbrcm",
	Short: "Firebase project viewer",
}

func Execute(s *core.Core) {
	corelog.For("cli").Debug("register cli commands")
	rootCmd.AddCommand(addcmd.New(s))
	rootCmd.AddCommand(cachecmd.New())
	rootCmd.AddCommand(deletecmd.New(s))
	rootCmd.AddCommand(getcmd.New(s))
	rootCmd.AddCommand(logincmd.New(s))
	rootCmd.AddCommand(projectcmd.New(s))
	rootCmd.AddCommand(projectscmd.New(s))

	if err := rootCmd.Execute(); err != nil {
		corelog.For("cli").Error("cli command failed", "err", err)
		os.Exit(1)
	}
}

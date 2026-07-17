package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	addcmd "github.com/yumauri/fbrcm/cli/commands/add"
	authcmd "github.com/yumauri/fbrcm/cli/commands/auth"
	cachecmd "github.com/yumauri/fbrcm/cli/commands/cache"
	conditionscmd "github.com/yumauri/fbrcm/cli/commands/conditions"
	configcmd "github.com/yumauri/fbrcm/cli/commands/config"
	deletecmd "github.com/yumauri/fbrcm/cli/commands/delete"
	doctorcmd "github.com/yumauri/fbrcm/cli/commands/doctor"
	draftcmd "github.com/yumauri/fbrcm/cli/commands/draft"
	getcmd "github.com/yumauri/fbrcm/cli/commands/get"
	groupscmd "github.com/yumauri/fbrcm/cli/commands/groups"
	profilecmd "github.com/yumauri/fbrcm/cli/commands/profile"
	projectcmd "github.com/yumauri/fbrcm/cli/commands/project"
	projectscmd "github.com/yumauri/fbrcm/cli/commands/projects"
	updatecmd "github.com/yumauri/fbrcm/cli/commands/update"
	versionscmd "github.com/yumauri/fbrcm/cli/commands/versions"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

const versionTemplate = `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s\n" .Version}}`

func isProfileCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "profile" || strings.HasPrefix(cmd.CommandPath(), "fbrcm profile")
}

func newRootCommand(s *core.Core, version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "fbrcm",
		Short: "Firebase Remote Config manager",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			profileName, err := cmd.Flags().GetString("profile")
			if err != nil {
				return err
			}
			if err := config.SetProfileOverride(profileName); err != nil {
				return fmt.Errorf("select profile: %w", err)
			}
			if isProfileCommand(cmd) || cmd.Name() == "doctor" || cmd.Name() == "help" {
				return nil
			}
			if err := config.EnsureActiveProfile(); err != nil {
				return fmt.Errorf("ensure active profile: %w", err)
			}
			return nil
		},
	}
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	rootCmd.SetVersionTemplate(versionTemplate)
	profileDefault, _ := env.LookupTrimmed(env.Profile)
	rootCmd.PersistentFlags().String("profile", profileDefault, "Use profile for this invocation without changing the active profile (env: FBRCM_PROFILE)")

	rootCmd.AddCommand(addcmd.New(s))
	rootCmd.AddCommand(authcmd.New(s))
	rootCmd.AddCommand(cachecmd.New())
	rootCmd.AddCommand(conditionscmd.New(s))
	rootCmd.AddCommand(configcmd.New())
	rootCmd.AddCommand(deletecmd.New(s))
	rootCmd.AddCommand(doctorcmd.New(s))
	rootCmd.AddCommand(draftcmd.New(s))
	rootCmd.AddCommand(getcmd.New(s))
	rootCmd.AddCommand(groupscmd.New(s))
	rootCmd.AddCommand(profilecmd.New())
	rootCmd.AddCommand(projectcmd.New(s))
	rootCmd.AddCommand(projectscmd.New(s))
	rootCmd.AddCommand(updatecmd.New(s))
	rootCmd.AddCommand(versionscmd.New(s))

	return rootCmd
}

func Execute(s *core.Core, version, commit, date string) {
	corelog.For("cli").Debug("register cli commands")
	rootCmd := newRootCommand(s, version, commit, date)
	if err := rootCmd.Execute(); err != nil {
		exitCode := 1
		var exitErr *shared.ExitError
		if errors.As(err, &exitErr) && exitErr.Code > 0 {
			exitCode = exitErr.Code
		}
		if err.Error() != "" {
			corelog.For("cli").Error("cli command failed", "err", err)
		}
		os.Exit(exitCode)
	}
}

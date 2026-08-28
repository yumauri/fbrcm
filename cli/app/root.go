package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	addcmd "github.com/yumauri/fbrcm/cli/commands/add"
	authcmd "github.com/yumauri/fbrcm/cli/commands/auth"
	cachecmd "github.com/yumauri/fbrcm/cli/commands/cache"
	conditionscmd "github.com/yumauri/fbrcm/cli/commands/conditions"
	configcmd "github.com/yumauri/fbrcm/cli/commands/config"
	deletecmd "github.com/yumauri/fbrcm/cli/commands/delete"
	doctorcmd "github.com/yumauri/fbrcm/cli/commands/doctor"
	draftcmd "github.com/yumauri/fbrcm/cli/commands/draft"
	duplicatecmd "github.com/yumauri/fbrcm/cli/commands/duplicate"
	getcmd "github.com/yumauri/fbrcm/cli/commands/get"
	groupscmd "github.com/yumauri/fbrcm/cli/commands/groups"
	hookscmd "github.com/yumauri/fbrcm/cli/commands/hooks"
	managedfeaturescmd "github.com/yumauri/fbrcm/cli/commands/managedfeatures"
	metacmd "github.com/yumauri/fbrcm/cli/commands/meta"
	profilecmd "github.com/yumauri/fbrcm/cli/commands/profile"
	projectcmd "github.com/yumauri/fbrcm/cli/commands/project"
	projectscmd "github.com/yumauri/fbrcm/cli/commands/projects"
	themecmd "github.com/yumauri/fbrcm/cli/commands/theme"
	updatecmd "github.com/yumauri/fbrcm/cli/commands/update"
	versionscmd "github.com/yumauri/fbrcm/cli/commands/versions"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/about"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func isProfileCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "profile" || strings.HasPrefix(cmd.CommandPath(), "fbrcm profile")
}

func isConfigCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "config" || strings.HasPrefix(cmd.CommandPath(), "fbrcm config")
}

func isThemeCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "theme" || strings.HasPrefix(cmd.CommandPath(), "fbrcm theme")
}

func isHooksCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "hooks" || strings.HasPrefix(cmd.CommandPath(), "fbrcm hooks")
}

func isProjectAliasesCommand(cmd *cobra.Command) bool {
	return strings.HasPrefix(cmd.CommandPath(), "fbrcm projects aliases")
}

func isContractMetadataCommand(cmd *cobra.Command) bool {
	return strings.HasPrefix(cmd.CommandPath(), "fbrcm capabilities") ||
		strings.HasPrefix(cmd.CommandPath(), "fbrcm schema") ||
		strings.HasPrefix(cmd.CommandPath(), "fbrcm completion")
}

func newRootCommand(s *core.Core, version, commit, date string) *cobra.Command {
	return newRootCommandWithOfflineInit(s, version, commit, date, firebase.InitOfflineModeContext)
}

// NewRootForContract returns a complete command tree for schema generation and
// contract conformance tests. It performs no command execution by itself.
func NewRootForContract(version string) *cobra.Command {
	return newRootCommandWithOfflineInit(nil, version, "schema", "schema", func(context.Context, bool) {})
}

func newRootCommandWithOfflineInit(s *core.Core, version, commit, date string, initOfflineMode func(context.Context, bool)) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "fbrcm",
		Short: "Firebase Remote Config manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, contract.Capabilities(cmd.Root()))
			}
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := shared.ValidateNonBlankInputs(cmd, args); err != nil {
				return err
			}
			if err := cmd.ValidateRequiredFlags(); err != nil {
				return shared.InvalidArgument(err)
			}
			if err := cmd.ValidateFlagGroups(); err != nil {
				return shared.InvalidArgument(err)
			}
			if err := shared.ValidateQueryFlags(cmd); err != nil {
				return err
			}
			if err := shared.CommandContext(cmd).Err(); err != nil {
				return err
			}
			ctx := shared.CommandContext(cmd)
			stateless, err := cmd.Flags().GetBool("stateless")
			if err != nil {
				return shared.InvalidArgument(err)
			}
			if stateless {
				ctx = core.WithExecutionPolicy(ctx, core.StatelessExecutionPolicy())
				ctx = machine.WithProfileless(ctx)
				cmd.SetContext(ctx)
				commandID := contract.CommandID(cmd)
				if !contract.SupportsStatelessCommand(commandID) {
					return shared.InvalidArgument(fmt.Errorf("--stateless is not supported by %s", strings.TrimPrefix(cmd.CommandPath(), "fbrcm ")))
				}
				if cmd.Flags().Changed("profile") {
					return shared.InvalidArgument(fmt.Errorf("--profile cannot be used with --stateless"))
				}
				requiresAccessToken := contract.StatelessCommandRequiresAccessToken(commandID)
				if commandID == "get" && shared.StdinAvailable(cmd.InOrStdin()) {
					requiresAccessToken = false
				}
				if requiresAccessToken {
					if _, ok := env.LookupNonEmpty(env.GoogleAccessToken); !ok {
						return &core.AuthError{
							Kind: "configuration",
							Err:  fmt.Errorf("%s is required with --stateless", env.GoogleAccessToken),
						}
					}
				}
				corelog.For("cli.stateless").Info(
					"stateless mode enabled",
					"command", commandID,
				)
			} else {
				ctx = core.WithExecutionPolicy(ctx, core.StatefulExecutionPolicy())
			}
			if contract.Enabled(cmd) {
				ctx = firebase.WithOAuthInteractionAllowed(ctx, false)
			}
			cmd.SetContext(ctx)
			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				return shared.InvalidArgument(err)
			}
			if cmd.Flags().Changed("timeout") && timeout <= 0 {
				return shared.InvalidArgument(fmt.Errorf("--timeout must be greater than zero"))
			}
			if s != nil {
				s.SetHookOutput(cmd.ErrOrStderr())
			}
			noLocalConfig, err := cmd.Flags().GetBool("no-local-config")
			if err != nil {
				return shared.InvalidArgument(err)
			}
			config.SetLocalConfigDisabled(noLocalConfig || stateless)
			if cmd.Name() == "help" || contract.IsCommandGroup(cmd) || isConfigCommand(cmd) || isThemeCommand(cmd) || isHooksCommand(cmd) || isProjectAliasesCommand(cmd) || isContractMetadataCommand(cmd) {
				initOfflineMode(shared.CommandContext(cmd), false)
				return nil
			}
			if s != nil {
				if stateless {
					s.ResetFirebaseRequestPolicy()
				} else if err := s.ConfigureFirebaseRequests(); err != nil {
					if cmd.Name() != "doctor" {
						return err
					}
					s.ResetFirebaseRequestPolicy()
				}
				ctx = s.WithFirebaseRequestController(shared.CommandContext(cmd))
				cmd.SetContext(ctx)
			}
			progress.Start(commandProgressMessage(cmd))
			if machine.Profileless(shared.CommandContext(cmd)) {
				initOfflineMode(shared.CommandContext(cmd), false)
				return nil
			}
			profileName, err := cmd.Flags().GetString("profile")
			if err != nil {
				return shared.InvalidArgument(err)
			}
			if err := config.SetProfileOverride(profileName); err != nil {
				return &shared.ValidationError{Code: "profile.invalid", Source: "profile", Stage: "selection", Err: fmt.Errorf("select profile: %w", err)}
			}
			if isProfileCommand(cmd) || cmd.Name() == "doctor" {
				initOfflineMode(shared.CommandContext(cmd), false)
				return nil
			}
			if err := config.EnsureActiveProfile(); err != nil {
				return &shared.ValidationError{Code: "profile.invalid", Source: "profile", Stage: "selection", Err: fmt.Errorf("ensure active profile: %w", err)}
			}
			initOfflineMode(shared.CommandContext(cmd), false)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			progress.Stop()
		},
	}
	rootCmd.SetContext(shared.WithMachineState(context.Background()))
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return shared.InvalidArgument(err) })
	rootCmd.SetOut(progress.StopWriter(os.Stdout))
	rootCmd.SetErr(progress.StopWriter(os.Stderr))
	rootCmd.Version = (about.BuildInfo{Version: version, Commit: commit, Date: date}).Metadata()
	rootCmd.SetVersionTemplate(buildVersionTemplate())
	profileDefault, _ := env.LookupTrimmed(env.Profile)
	rootCmd.PersistentFlags().String("profile", profileDefault, "Use profile for this invocation without changing the active profile (env: FBRCM_PROFILE)")
	rootCmd.PersistentFlags().Bool("stateless", false, "Run a supported command without profiles or application-managed local state (Firebase API token env: FBRCM_GOOGLE_ACCESS_TOKEN)")
	rootCmd.PersistentFlags().Bool("no-local-config", false, "Ignore .fbrcm.toml repository configuration (env: FBRCM_NO_LOCAL_CONFIG)")
	rootCmd.PersistentFlags().Bool("json", false, "Emit one versioned machine-readable JSON envelope")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Maximum duration for the complete command")

	rootCmd.AddCommand(addcmd.New(s))
	rootCmd.AddCommand(authcmd.New(s))
	rootCmd.AddCommand(cachecmd.New())
	rootCmd.AddCommand(conditionscmd.New(s))
	rootCmd.AddCommand(configcmd.New())
	rootCmd.AddCommand(deletecmd.New(s))
	rootCmd.AddCommand(doctorcmd.New(s))
	rootCmd.AddCommand(draftcmd.New(s))
	rootCmd.AddCommand(duplicatecmd.New(s))
	rootCmd.AddCommand(managedfeaturescmd.NewExperiments(s))
	rootCmd.AddCommand(getcmd.New(s))
	rootCmd.AddCommand(groupscmd.New(s))
	rootCmd.AddCommand(hookscmd.New())
	rootCmd.AddCommand(managedfeaturescmd.NewPersonalizations(s))
	rootCmd.AddCommand(profilecmd.New())
	rootCmd.AddCommand(projectcmd.New(s))
	rootCmd.AddCommand(projectscmd.New(s))
	rootCmd.AddCommand(themecmd.New())
	rootCmd.AddCommand(managedfeaturescmd.NewRollouts(s))
	rootCmd.AddCommand(metacmd.NewCapabilities())
	rootCmd.AddCommand(metacmd.NewSchema())
	rootCmd.AddCommand(updatecmd.New(s))
	rootCmd.AddCommand(versionscmd.New(s))
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	installCompletionMachineOutput(rootCmd)
	installCommandErrorBoundaries(rootCmd)
	contract.RegisterResponse(rootCmd, contract.CapabilityIndex{}, contract.TextData{})
	contract.MustRegisterResponsePath(rootCmd, "help", contract.TextData{})
	for _, shell := range []string{"bash", "fish", "powershell", "zsh"} {
		contract.MustRegisterResponsePath(rootCmd, "completion "+shell, contract.TextData{})
	}

	return rootCmd
}

func installCompletionMachineOutput(root *cobra.Command) {
	completion, _, err := root.Find([]string{"completion"})
	if err != nil {
		panic(fmt.Sprintf("find completion command: %v", err))
	}
	for _, shell := range []string{"bash", "fish", "powershell", "zsh"} {
		command, _, findErr := completion.Find([]string{shell})
		if findErr != nil {
			panic(fmt.Sprintf("find completion %s command: %v", shell, findErr))
		}
		original := command.RunE
		command.RunE = func(cmd *cobra.Command, args []string) error {
			if !contract.Enabled(cmd) {
				return original(cmd, args)
			}
			noDescriptions, flagErr := cmd.Flags().GetBool("no-descriptions")
			if flagErr != nil {
				return shared.InvalidArgument(flagErr)
			}
			var generated bytes.Buffer
			var generateErr error
			switch cmd.Name() {
			case "bash":
				generateErr = cmd.Root().GenBashCompletionV2(&generated, !noDescriptions)
			case "fish":
				generateErr = cmd.Root().GenFishCompletion(&generated, !noDescriptions)
			case "powershell":
				if noDescriptions {
					generateErr = cmd.Root().GenPowerShellCompletion(&generated)
				} else {
					generateErr = cmd.Root().GenPowerShellCompletionWithDesc(&generated)
				}
			case "zsh":
				if noDescriptions {
					generateErr = cmd.Root().GenZshCompletionNoDesc(&generated)
				} else {
					generateErr = cmd.Root().GenZshCompletion(&generated)
				}
			default:
				return &shared.ValidationError{Code: "command.not_found", Source: "command", Stage: "selection", Target: cmd.Name(), Err: fmt.Errorf("unsupported completion shell %q", cmd.Name())}
			}
			if generateErr != nil {
				return generateErr
			}
			return shared.WriteJSON(cmd, contract.TextData{Text: generated.String()})
		}
	}
}

func installCommandErrorBoundaries(root *cobra.Command) {
	var visit func(*cobra.Command)
	visit = func(cmd *cobra.Command) {
		if cmd != root && cmd.RunE == nil && cmd.Run == nil && cmd.HasAvailableSubCommands() {
			contract.MarkCommandGroup(cmd)
			cmd.Args = cobra.NoArgs
			cmd.RunE = func(cmd *cobra.Command, _ []string) error {
				return cmd.Help()
			}
		}
		if cmd.Args != nil {
			validate := cmd.Args
			cmd.Args = func(cmd *cobra.Command, args []string) error {
				return shared.InvalidArgument(validate(cmd, args))
			}
		}
		if cmd.RunE != nil {
			run := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				shared.MarkCommandRun(cmd)
				return run(cmd, args)
			}
		} else if cmd.Run != nil {
			run := cmd.Run
			cmd.Run = func(cmd *cobra.Command, args []string) {
				shared.MarkCommandRun(cmd)
				run(cmd, args)
			}
		}
		for _, child := range cmd.Commands() {
			visit(child)
		}
	}
	visit(root)
}

func Execute(s *core.Core, version, commit, date string) {
	corelog.For("cli").Debug("register cli commands")
	rootCmd := newRootCommand(s, version, commit, date)
	jsonMode := contract.JSONRequested(os.Args[1:])
	var captured bytes.Buffer
	if jsonMode {
		shared.SetMachineMode(true)
		defer shared.SetMachineMode(false)
		rootCmd.SetOut(&captured)
		if !shared.StdinAvailable(os.Stdin) {
			rootCmd.SetIn(shared.NonInteractiveInput())
		}
		rootCmd.SilenceErrors = true
		rootCmd.SilenceUsage = true
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if timeout := requestedTimeout(os.Args[1:]); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	if jsonMode {
		ctx = firebase.WithOAuthInteractionAllowed(ctx, false)
	}
	rootCmd.SetContext(shared.WithMachineState(ctx))
	executedCmd, err := rootCmd.ExecuteC()
	err = completedCommandError(ctx, err)
	progress.Stop()
	if jsonMode {
		envelope := contract.BuildEnvelope(executedCmd, version, captured.Bytes(), err)
		if writeErr := contract.Write(os.Stdout, envelope); writeErr != nil {
			corelog.For("cli").Error("write CLI contract envelope", "err", writeErr)
			os.Exit(13)
		}
		if envelope.ExitCode != 0 {
			os.Exit(envelope.ExitCode)
		}
		return
	}
	if err != nil {
		exitCode := commandExitCode(executedCmd, err)
		if err.Error() != "" {
			corelog.For("cli").Error("cli command failed", "err", err)
		}
		os.Exit(exitCode)
	}
}

func completedCommandError(ctx context.Context, commandErr error) error {
	if commandErr != nil {
		return commandErr
	}
	return ctx.Err()
}

func requestedTimeout(args []string) time.Duration {
	var requested time.Duration
	for index, arg := range args {
		if arg == "--" {
			break
		}
		if value, ok := strings.CutPrefix(arg, "--timeout="); ok {
			if timeout, err := time.ParseDuration(value); err == nil {
				requested = timeout
			}
			continue
		}
		if arg == "--timeout" && index+1 < len(args) {
			if timeout, err := time.ParseDuration(args[index+1]); err == nil {
				requested = timeout
			}
		}
	}
	return requested
}

func commandProgressMessage(cmd *cobra.Command) string {
	path := strings.TrimPrefix(cmd.CommandPath(), "fbrcm ")
	switch {
	case path == "get":
		return "Loading Remote Config…"
	case path == "projects list":
		return "Loading projects…"
	case path == "projects update":
		return "Syncing projects…"
	case path == "doctor":
		return "Running diagnostics…"
	case path == "auth login":
		return "Authenticating…"
	case strings.HasPrefix(path, "versions "):
		return "Loading Remote Config versions…"
	case path == "draft publish":
		return "Preparing Remote Config drafts…"
	case path == "project import":
		return "Preparing Remote Config import…"
	case strings.HasPrefix(path, "experiments"),
		strings.HasPrefix(path, "rollouts"),
		strings.HasPrefix(path, "personalizations"):
		return "Loading managed features…"
	case path == "add", path == "update", path == "delete", path == "duplicate",
		strings.HasPrefix(path, "groups "), strings.HasPrefix(path, "conditions "),
		path == "projects promote":
		return "Preparing Remote Config changes…"
	default:
		return "Working…"
	}
}

func commandExitCode(cmd *cobra.Command, err error) int {
	return contract.ExitCode(cmd, err)
}

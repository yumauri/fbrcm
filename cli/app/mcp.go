package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/mcpserver"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

type hostedExecutionKey struct{}
type hostedExecution struct{ oauth, hooks bool }

func newMCPCommand(service *core.Core, version, commit, date string) *cobra.Command {
	var options mcpserver.Options
	command := &cobra.Command{Use: "mcp", Short: "Serve Firebase Remote Config tools over MCP stdio", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if contract.Enabled(cmd) {
				return shared.InvalidArgument(fmt.Errorf("mcp uses JSON-RPC framing; --json cannot be combined with mcp"))
			}
			options.Stateless, _ = cmd.Flags().GetBool("stateless")
			options.NoLocalConfig, _ = cmd.Flags().GetBool("no-local-config")
			options.Profile, _ = cmd.Flags().GetString("profile")
			if options.Stateless && !cmd.Flags().Changed("profile") {
				options.Profile = ""
			}
			if err := options.Validate(); err != nil {
				return shared.InvalidArgument(err)
			}
			config.SetLocalConfigDisabled(options.NoLocalConfig || options.Stateless)
			if !options.Stateless {
				if err := config.SetProfileOverride(options.Profile); err != nil {
					return &shared.ValidationError{Code: "profile.invalid", Source: "profile", Stage: "selection", Err: err}
				}
				if err := config.EnsureActiveProfile(); err != nil {
					return &shared.ValidationError{Code: "profile.invalid", Source: "profile", Stage: "selection", Err: err}
				}
				options.Profile = config.GetActiveProfileName()
			}
			progress.Configure(cmd.ErrOrStderr(), false)
			shared.SetMachineMode(true)
			defer shared.SetMachineMode(false)
			execute := func(ctx context.Context, capability contract.Capability, input mcpserver.Invocation, confirmed, oauth bool, observer func(core.OAuthAuthorizationEvent)) contract.Envelope {
				return runHostedMachine(ctx, service, version, commit, date, capability, input, options, confirmed, oauth, observer, cmd.ErrOrStderr())
			}
			root := NewRootForContract(version)
			defer contract.UnregisterResponses(root)
			server, err := mcpserver.New(cmd.Context(), root, version, options, execute)
			if err != nil {
				return err
			}
			defer server.Close()
			cmd.Root().SilenceUsage = true
			cmd.Root().SilenceErrors = true
			reader, ok := cmd.InOrStdin().(io.ReadCloser)
			if !ok {
				reader = io.NopCloser(cmd.InOrStdin())
			}
			return server.Protocol.Run(cmd.Context(), &mcp.IOTransport{Reader: reader, Writer: writeCloser{cmd.OutOrStdout()}})
		},
	}
	command.Flags().StringSliceVar(&options.Toolsets, "toolsets", []string{"inspect", "edit", "drafts", "plans", "publish"}, "Enabled toolsets: inspect, edit, drafts, plans, publish, diagnostics")
	command.Flags().BoolVar(&options.AllowWrites, "allow-writes", false, "Enable mutation tools and explicit artifact writes (inspection may still refresh caches and credentials)")
	command.Flags().BoolVar(&options.AllowHooks, "allow-hooks", false, "Allow already-trusted configured hooks during tool execution")
	command.Flags().StringVar(&options.Confirmation, "confirmation", "host", "Mutation approval: host (explicit elicitation) or none (operator-authorized unattended execution)")
	command.Flags().StringVar(&options.BrowserAuth, "browser-auth", "auto", "Browser authorization: auto (ask supporting hosts) or never")
	command.Flags().DurationVar(&options.RequestTimeout, "request-timeout", 5*time.Minute, "Maximum duration per tool operation, including queueing and user interaction")
	command.Flags().DurationVar(&options.AuthTimeout, "auth-timeout", 2*time.Minute, "Maximum duration of a browser authorization attempt")
	contract.RegisterNoData(command)
	return command
}

// IsMCPInvocation resolves the command without executing it or loading profile
// state. Presentation setup must not initialize a theme for a protocol server.
func IsMCPInvocation(args []string) bool {
	root := NewRootForContract("discovery")
	defer contract.UnregisterResponses(root)
	command, _, err := root.Find(args)
	return err == nil && command.Name() == "mcp"
}

type writeCloser struct{ io.Writer }

func (writeCloser) Close() error { return nil }

func runHostedMachine(ctx context.Context, service *core.Core, version, commit, date string, capability contract.Capability, input mcpserver.Invocation, options mcpserver.Options, confirmed, oauth bool, observer func(core.OAuthAuthorizationEvent), stderr io.Writer) contract.Envelope {
	root := newRootCommand(service, version, commit, date)
	defer contract.UnregisterResponses(root)
	ctx = context.WithValue(ctx, hostedExecutionKey{}, hostedExecution{oauth: oauth, hooks: options.AllowHooks})
	ctx = firebase.WithOAuthTimeout(ctx, options.AuthTimeout)
	root.SetContext(shared.WithMachineState(ctx))
	root.SilenceUsage, root.SilenceErrors = true, true
	var captured bytes.Buffer
	root.SetOut(&captured)
	root.SetErr(stderr)
	root.SetIn(shared.NonInteractiveInput())
	if len(input.Stdin) != 0 && !bytes.Equal(bytes.TrimSpace(input.Stdin), []byte("null")) {
		root.SetIn(shared.DocumentInput(bytes.NewReader(input.Stdin)))
	}
	argv, err := input.Argv(capability, options, confirmed)
	if err != nil {
		command, _, _ := root.Find(capability.Path)
		return contract.BuildEnvelope(command, version, nil, shared.InvalidArgument(err))
	}
	root.SetArgs(argv)
	if service != nil {
		service.ResetFirebaseClients()
		service.ConfigureOAuthAuthorization(false, observer)
	}
	executed, err := root.ExecuteC()
	progress.Stop()
	return contract.BuildEnvelope(executed, version, captured.Bytes(), completedCommandError(ctx, err))
}

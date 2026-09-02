// Package mcp is the MCP frontend. Command-line arguments select and configure
// this process mode; tool execution never enters the CLI frontend.
package mcp

import (
	"context"
	"fmt"
	"io"
	"time"

	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/about"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/mcp/server"
	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/launchflags"
	"github.com/yumauri/fbrcm/ops/shared"
)

func NewCommand(service *core.Core, version, commit, date string) *cobra.Command {
	var options mcpserver.Options
	command := &cobra.Command{Use: "mcp", Short: "Serve Firebase Remote Config tools over MCP stdio", Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return shared.ValidateNonBlankInputs(cmd, args) },
		RunE: func(cmd *cobra.Command, _ []string) error {
			if contract.Enabled(cmd) {
				return shared.InvalidArgument(fmt.Errorf("mcp uses JSON-RPC framing; --json cannot be combined with mcp"))
			}
			timeout, _ := cmd.Flags().GetDuration("timeout")
			if cmd.Flags().Changed("timeout") && timeout <= 0 {
				return shared.InvalidArgument(fmt.Errorf("--timeout must be greater than zero"))
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
			ctx := shared.CommandContext(cmd)
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			registry, err := ops.NewRegistry(service)
			if err != nil {
				return err
			}
			execute := func(ctx context.Context, c contract.Capability, input mcpserver.Invocation, confirmed, oauth bool, observer func(core.OAuthAuthorizationEvent)) contract.Envelope {
				return registry.Execute(ctx, c, input, ops.Execution{Version: version, BuildVersion: (about.BuildInfo{Version: version, Commit: commit, Date: date}).Metadata(), Profile: options.Profile, Stateless: options.Stateless, NoLocalConfig: options.NoLocalConfig, AllowHooks: options.AllowHooks, AllowOAuth: oauth, Confirmed: confirmed, AuthTimeout: options.AuthTimeout, OAuthObserver: observer, Stderr: cmd.ErrOrStderr()})
			}
			server, err := mcpserver.New(ctx, registry.Capabilities(), version, options, execute)
			if err != nil {
				return err
			}
			defer server.Close()
			cmd.Root().SilenceErrors = true
			cmd.Root().SilenceUsage = true
			reader, ok := cmd.InOrStdin().(io.ReadCloser)
			if !ok {
				reader = io.NopCloser(cmd.InOrStdin())
			}
			return server.Protocol.Run(ctx, &protocol.IOTransport{Reader: reader, Writer: writeCloser{cmd.OutOrStdout()}})
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

type writeCloser struct{ io.Writer }

func (writeCloser) Close() error { return nil }

func rootCommand(service *core.Core, version, commit, date string) *cobra.Command {
	root := &cobra.Command{Use: "fbrcm", Short: "Firebase Remote Config manager", SilenceErrors: true, SilenceUsage: true}
	launchflags.Add(root.PersistentFlags())
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return shared.InvalidArgument(err) })
	root.AddCommand(NewCommand(service, version, commit, date))
	return root
}

func IsInvocation(args []string) bool {
	root := rootCommand(nil, "discovery", "", "")
	defer contract.UnregisterResponses(root)
	cmd, _, err := root.Find(args)
	return err == nil && cmd.Name() == "mcp"
}

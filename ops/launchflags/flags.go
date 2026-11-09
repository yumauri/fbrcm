// Package launchflags owns shared process-launch options. It is not involved in
// executing workflows or decoding MCP tool inputs.
package launchflags

import (
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/core/env"
)

type Options struct {
	JSON          bool
	Timeout       time.Duration
	Stateless     bool
	NoLocalConfig bool
}

// Inspect reads process-launch flags without setting any command flag values.
// Cobra remains responsible for reporting invalid arguments. The preliminary
// parse tolerates invalid values and unknown flags so those errors can still
// be reported in the requested output format.
func Inspect(root *cobra.Command, args []string) Options {
	cmd, args, _ := root.Find(args)
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultVersionFlag()

	parser := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	parser.SetOutput(io.Discard)
	parser.SetNormalizeFunc(cmd.Flags().GetNormalizeFunc())
	parser.ParseErrorsAllowlist.UnknownFlagsHandling = pflag.IgnoreUnknownFlag
	parser.AddFlagSet(cmd.LocalFlags())
	parser.AddFlagSet(cmd.InheritedFlags())

	var options Options
	for {
		options = Options{}
		var invalidJSON bool
		err := parser.ParseAll(args, func(flag *pflag.Flag, value string) error {
			switch flag.Name {
			case "json":
				enabled, err := strconv.ParseBool(value)
				invalidJSON = invalidJSON || err != nil
				options.JSON = enabled
			case "timeout":
				options.Timeout, _ = time.ParseDuration(value)
			case "stateless":
				options.Stateless, _ = strconv.ParseBool(value)
			case "no-local-config":
				options.NoLocalConfig, _ = strconv.ParseBool(value)
			}
			return nil
		})
		if syntax, ok := errors.AsType[*pflag.InvalidSyntaxError](err); ok {
			// Treat malformed flags as ignorable unknown flags without
			// changing the number of tokens consumed by other options.
			args = slices.Clone(args)
			for i, arg := range args {
				if arg == syntax.GetSpecifiedFlag() {
					args[i] = "--fbrcm-invalid-launch-syntax=" + arg
				}
			}
			continue
		}
		options.JSON = options.JSON || invalidJSON
		break
	}

	// Cobra's private shell protocol leaves the command's completion tokens
	// unparsed. An explicit JSON flag in those tokens requests a rejection.
	if cmd == root {
		name := parser.Arg(0)
		if name == cobra.ShellCompRequestCmd || name == cobra.ShellCompNoDescRequestCmd {
			if i := slices.Index(args, name); i >= 0 {
				options.JSON = shellCompletionJSON(args[i+1:], options.JSON)
			}
		}
	}
	return options
}

func shellCompletionJSON(args []string, enabled bool) bool {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--json" {
			enabled = true
		}
		if value, ok := strings.CutPrefix(arg, "--json="); ok {
			parsed, err := strconv.ParseBool(value)
			enabled = parsed || err != nil
		}
	}
	return enabled
}

func Add(flags *pflag.FlagSet) {
	profileDefault, _ := env.LookupTrimmed(env.Profile)
	flags.String("profile", profileDefault, "Use profile for this invocation without changing the active profile (env: FBRCM_PROFILE)")
	flags.Bool("stateless", false, "Run a supported command without profiles or application-managed local state (Firebase API token env: FBRCM_GOOGLE_ACCESS_TOKEN)")
	flags.Bool("include-hidden", false, "Include projects hidden by the profile blacklist or whitelist for this invocation")
	flags.Bool("no-local-config", false, "Ignore .fbrcm.toml repository configuration (env: FBRCM_NO_LOCAL_CONFIG)")
	flags.Bool("json", false, "Emit one versioned machine-readable JSON envelope")
	flags.Duration("timeout", 0, "Maximum duration for the complete command")
}

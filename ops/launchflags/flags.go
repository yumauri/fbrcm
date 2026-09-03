// Package launchflags owns shared process-launch options. It is not involved in
// executing workflows or decoding MCP tool inputs.
package launchflags

import (
	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/core/env"
)

func Add(flags *pflag.FlagSet) {
	profileDefault, _ := env.LookupTrimmed(env.Profile)
	flags.String("profile", profileDefault, "Use profile for this invocation without changing the active profile (env: FBRCM_PROFILE)")
	flags.Bool("stateless", false, "Run a supported command without profiles or application-managed local state (Firebase API token env: FBRCM_GOOGLE_ACCESS_TOKEN)")
	flags.Bool("no-local-config", false, "Ignore .fbrcm.toml repository configuration (env: FBRCM_NO_LOCAL_CONFIG)")
	flags.Bool("json", false, "Emit one versioned machine-readable JSON envelope")
	flags.Duration("timeout", 0, "Maximum duration for the complete command")
}

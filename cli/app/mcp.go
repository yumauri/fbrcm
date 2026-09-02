package app

import (
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/mcp"
)

// The CLI retains launch metadata for help, completion, and discovery. main
// dispatches real MCP invocations directly to the independent frontend.
func newMCPCommand(service *core.Core, version, commit, date string) *cobra.Command {
	return mcp.NewCommand(service, version, commit, date)
}

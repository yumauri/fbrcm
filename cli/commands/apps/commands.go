package apps

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/workflows/apps"
)

func New(svc *core.Core) *cobra.Command {
	return cliadapter.Command(apps.NewDefinition(svc))
}

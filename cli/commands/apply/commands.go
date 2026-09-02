package applycmd

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
	workflows "github.com/yumauri/fbrcm/ops/workflows/apply"
)

func New(svc *core.Core) *cobra.Command { return cliadapter.Command(workflows.NewDefinition(svc)) }

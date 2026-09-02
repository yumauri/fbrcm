package plancmd

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	workflows "github.com/yumauri/fbrcm/ops/workflows/plan"
)

func New() *cobra.Command { return cliadapter.Command(workflows.NewDefinition()) }

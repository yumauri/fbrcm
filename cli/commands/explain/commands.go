package explain

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	workflow "github.com/yumauri/fbrcm/ops/workflows/explain"
)

func New(svc *core.Core) *cobra.Command { return cliadapter.Command(workflow.NewDefinition(svc)) }

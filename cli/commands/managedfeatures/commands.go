package managedfeatures

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
	workflows "github.com/yumauri/fbrcm/ops/workflows/managedfeatures"
)

func NewExperiments(svc *core.Core) *cobra.Command {
	return cliadapter.Command(workflows.NewExperimentsDefinition(svc))
}
func NewRollouts(svc *core.Core) *cobra.Command {
	return cliadapter.Command(workflows.NewRolloutsDefinition(svc))
}
func NewPersonalizations(svc *core.Core) *cobra.Command {
	return cliadapter.Command(workflows.NewPersonalizationsDefinition(svc))
}

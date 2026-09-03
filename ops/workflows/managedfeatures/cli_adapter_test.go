package managedfeatures

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
)

func NewExperiments(svc *core.Core) *cobra.Command {
	return cliadapter.Command(NewExperimentsDefinition(svc))
}
func NewRollouts(svc *core.Core) *cobra.Command {
	return cliadapter.Command(NewRolloutsDefinition(svc))
}
func NewPersonalizations(svc *core.Core) *cobra.Command {
	return cliadapter.Command(NewPersonalizationsDefinition(svc))
}

func newExperimentsShowCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newExperimentsShowCommandDefinition(svc))
}

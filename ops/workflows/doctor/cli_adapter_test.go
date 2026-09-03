package doctor

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
)

func New(svc *core.Core) *cobra.Command { return cliadapter.Command(NewDefinition(svc)) }
func newCommand(runDoctor doctorFunc, notifyContext notifyContextFunc) *cobra.Command {
	return cliadapter.Command(newCommandDefinition(runDoctor, notifyContext))
}

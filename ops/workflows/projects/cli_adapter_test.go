package projects

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
)

func newAliasesListCommand() *cobra.Command {
	return cliadapter.Command(newAliasesListCommandDefinition())
}
func newAliasesSetCommand() *cobra.Command {
	return cliadapter.Command(newAliasesSetCommandDefinition())
}
func newAliasesRemoveCommand() *cobra.Command {
	return cliadapter.Command(newAliasesRemoveCommandDefinition())
}
func newAliasesImportCommand() *cobra.Command {
	return cliadapter.Command(newAliasesImportCommandDefinition())
}
func New(svc *core.Core) *cobra.Command { return cliadapter.Command(NewDefinition(svc)) }

func newListCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newListCommandDefinition(svc))
}

func newResetCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newResetCommandDefinition(svc))
}

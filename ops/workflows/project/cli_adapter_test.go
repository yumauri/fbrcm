package project

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	core "github.com/yumauri/fbrcm/core"
)

func New(svc *core.Core) *cobra.Command { return cliadapter.Command(NewDefinition(svc)) }
func newOpenCommand(svc *core.Core, openURL func(string) error) *cobra.Command {
	return cliadapter.Command(newOpenCommandDefinition(svc, openURL))
}

func newDefaultsCommandWithDownloader(svc *core.Core, download defaultsDownloader) *cobra.Command {
	return cliadapter.Command(newDefaultsCommandWithDownloaderDefinition(svc, download))
}

func newProjectQuotaProjectShowCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newProjectQuotaProjectShowCommandDefinition(svc))
}
func newProjectQuotaProjectSetCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newProjectQuotaProjectSetCommandDefinition(svc))
}
func newProjectQuotaProjectUnsetCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newProjectQuotaProjectUnsetCommandDefinition(svc))
}
func newShowCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newShowCommandDefinition(svc))
}

func newTemplatesShowCommand() *cobra.Command {
	return cliadapter.Command(newTemplatesShowCommandDefinition())
}
func newTemplatesSetCommand(svc *core.Core) *cobra.Command {
	return cliadapter.Command(newTemplatesSetCommandDefinition(svc))
}

package plancmd

import (
	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
)

func New() *cobra.Command { return cliadapter.Command(NewDefinition()) }

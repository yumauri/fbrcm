package cli

import (
	"fbrcm/cli/app"
	"fbrcm/core"
	corelog "fbrcm/core/log"
)

// Init initializes init and returns the resulting value or error.
func Init(s *core.Core) {
	corelog.For("cli").Debug("start cli")
	app.Execute(s)
}

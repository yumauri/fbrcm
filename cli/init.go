package cli

import (
	"fbrcm/cli/app"
	"fbrcm/core"
	corelog "fbrcm/core/log"
)

// Init initializes init and returns the resulting value or error.
func Init(s *core.Core, version, commit, date string) {
	corelog.For("cli").Debug("start cli")
	app.Execute(s, version, commit, date)
}

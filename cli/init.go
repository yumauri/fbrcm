package cli

import (
	"fbrcm/cli/app"
	"fbrcm/core"
	corelog "fbrcm/core/log"
)

func Init(s *core.Core) {
	corelog.For("cli").Debug("start cli")
	app.Execute(s)
}

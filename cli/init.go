package cli

import (
	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func Init(s *core.Core, version, commit, date string) {
	corelog.For("cli").Debug("start cli")
	app.Execute(s, version, commit, date)
}

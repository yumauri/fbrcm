package cli

import (
	"os"

	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func Init(s *core.Core, version, commit, date string) {
	progress.Configure(os.Stderr, term.IsTerminal(int(os.Stderr.Fd())))
	corelog.ConfigureCLIOutput(progress.LogWriter(os.Stderr), progress.StopWriter(os.Stderr))
	corelog.For("cli").Debug("start cli")
	app.Execute(s, version, commit, date)
}

package cli

import (
	"os"

	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/launchflags"
)

func Init(s *core.Core, version, commit, date string, launch launchflags.Options) {
	progress.Configure(os.Stderr, term.IsTerminal(int(os.Stderr.Fd())))
	corelog.ConfigureCLIOutput(progress.LogWriter(os.Stderr), progress.StopWriter(os.Stderr))
	configureTheme(launch)
	corelog.For("cli").Debug("start cli")
	app.Execute(s, version, commit, date, launch)
}

func configureTheme(launch launchflags.Options) {
	if env.NoColorEnabled() || launch.JSON || launch.Stateless {
		return
	}
	config.SetLocalConfigDisabled(launch.NoLocalConfig)
	resolved, err := config.ApplyConfiguredTheme()
	corelog.RefreshStyles()
	if err != nil {
		corelog.For("theme").Warn("theme unavailable; using built-in colors", "err", err)
		return
	}
	if resolved.Name != "" {
		corelog.For("theme").Debug("theme loaded", "theme", resolved.Name, "path", resolved.Path)
	}
}

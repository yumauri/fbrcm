package cli

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/contract"
)

func Init(s *core.Core, version, commit, date string) {
	progress.Configure(os.Stderr, term.IsTerminal(int(os.Stderr.Fd())))
	corelog.ConfigureCLIOutput(progress.LogWriter(os.Stderr), progress.StopWriter(os.Stderr))
	configureTheme(os.Args[1:])
	corelog.For("cli").Debug("start cli")
	app.Execute(s, version, commit, date)
}

func configureTheme(args []string) {
	if env.NoColorEnabled() || contract.JSONRequested(args) || booleanFlagEnabled(args, "stateless") {
		return
	}
	config.SetLocalConfigDisabled(booleanFlagEnabled(args, "no-local-config"))
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

func booleanFlagEnabled(args []string, name string) bool {
	flag := "--" + name
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == flag {
			return true
		}
		if value, ok := strings.CutPrefix(arg, flag+"="); ok {
			enabled, err := strconv.ParseBool(value)
			return err == nil && enabled
		}
	}
	return false
}

package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/about"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/app"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func Init(s *core.Core, version, commit, date string) {
	corelog.For("tui").Debug("start tui")
	if !env.NoColorEnabled() {
		resolved, err := config.ApplyConfiguredTheme()
		corelog.RefreshStyles()
		if err != nil {
			corelog.For("theme").Warn("theme unavailable; using built-in colors", "err", err)
		} else if resolved.Name != "" {
			corelog.For("theme").Debug("theme loaded", "theme", resolved.Name, "path", resolved.Path)
		}
	}
	if err := s.ConfigureFirebaseRequests(); err != nil {
		corelog.For("tui").Error("network config load failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if _, err := tuiconfig.Load(); err != nil {
		corelog.For("tui").Error("tui config load failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	m := app.NewWithBuildInfo(s, about.BuildInfo{Version: version, Commit: commit, Date: date})
	p := tea.NewProgram(m, programOptions()...)
	if _, err := p.Run(); err != nil {
		corelog.For("tui").Error("tui exited with error", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func programOptions() []tea.ProgramOption {
	if !env.NoColorEnabled() {
		return nil
	}
	return []tea.ProgramOption{tea.WithColorProfile(colorprofile.ASCII)}
}

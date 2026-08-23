package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/about"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestAboutOpensOnlyFromActionsAndShowsBuildInformation(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.buildInfo = about.BuildInfo{Version: "1.2.3", Commit: "abc123", Date: "2026-06-14"}
	m.helpPalette, _ = m.helpPalette.Open()
	m.helpPalette.input.SetValue("about")
	filtered := m.helpPalette.filtered(m.helpPaletteActions())
	if len(filtered) == 0 || filtered[0].action != tuiconfig.ActionAbout || !filtered[0].enabled || len(filtered[0].keys) != 0 {
		t.Fatalf("first About search result = %+v, want enabled palette-only action", filtered)
	}
	actions := filtered[:1]

	next, cmd, handled := m.runHelpPaletteAction(actions, helpPaletteListHeight(m.height))
	if !handled || cmd != nil || !next.aboutOpen || next.helpPalette.IsOpen() {
		t.Fatalf("run About = handled:%v cmd:%v about:%v actions:%v", handled, cmd != nil, next.aboutOpen, next.helpPalette.IsOpen())
	}
	if next.mouseMode() != tea.MouseModeAllMotion {
		t.Fatalf("About mouse mode = %v, want all motion", next.mouseMode())
	}

	view := ansi.Strip(next.aboutView())
	wants := []string{
		"fbrcm 1.2.3 (commit abc123, built 2026-06-14)",
		about.Author,
	}
	wants = append(strings.Split(about.Logo, "\n"), wants...)
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Fatalf("About popup missing %q:\n%s", want, view)
		}
	}
	lines := strings.Split(view, "\n")
	if len(lines) < 3 || strings.Trim(lines[len(lines)-2], "│ ") != "" {
		t.Fatalf("About popup has no empty row below author:\n%s", view)
	}
	authorLine := ""
	for _, line := range lines {
		if strings.Contains(line, about.Author) {
			authorLine = line
			break
		}
	}
	if authorLine != "│  "+about.Author+"  │" {
		t.Fatalf("About author padding = %q, want two spaces on each side", authorLine)
	}
}

func TestAboutClosesOnNonQuitKeysWithoutPropagatingThem(t *testing.T) {
	for _, key := range []tea.Key{{Code: 'x', Text: "x"}, {Code: '?', Text: "?"}, {Code: tea.KeyEnter}} {
		m := viewTestModel(90, 24, panels.Projects)
		m.aboutOpen = true
		next, cmd := m.Update(tea.KeyPressMsg(key))
		m = next.(Model)
		if m.aboutOpen || m.helpPalette.IsOpen() || cmd != nil {
			t.Fatalf("key %q = about:%v actions:%v cmd:%v, want consumed close", key.String(), m.aboutOpen, m.helpPalette.IsOpen(), cmd != nil)
		}
	}
}

func TestAboutPreservesQuitKeys(t *testing.T) {
	for _, key := range []tea.Key{{Code: 'q', Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}} {
		m := viewTestModel(90, 24, panels.Projects)
		m.aboutOpen = true
		next, cmd := m.Update(tea.KeyPressMsg(key))
		m = next.(Model)
		if m.aboutOpen || cmd == nil {
			t.Fatalf("quit key %q = open:%v cmd:%v, want application quit", key.String(), m.aboutOpen, cmd != nil)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("quit key %q command did not return tea.QuitMsg", key.String())
		}
	}
}

func TestAboutClosesOnMouseClickInsideOrOutside(t *testing.T) {
	for _, click := range []tea.MouseClickMsg{
		{X: 0, Y: 0, Button: tea.MouseLeft},
		{X: 45, Y: 12, Button: tea.MouseLeft},
	} {
		m := viewTestModel(90, 24, panels.Projects)
		m.aboutOpen = true
		next, cmd := m.Update(click)
		m = next.(Model)
		if m.aboutOpen || cmd != nil {
			t.Fatalf("click at %d,%d = open:%v cmd:%v, want close", click.X, click.Y, m.aboutOpen, cmd != nil)
		}
	}
}

func TestAboutKeepsSizeAndRecentersOnTerminalResize(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.aboutOpen = true
	before := m.aboutView()
	beforeX, beforeY := m.aboutPosition()

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	m = next.(Model)
	after := m.aboutView()
	afterX, afterY := m.aboutPosition()
	if !m.aboutOpen {
		t.Fatal("About closed on terminal resize")
	}
	if lipgloss.Width(after) != lipgloss.Width(before) || lipgloss.Height(after) != lipgloss.Height(before) {
		t.Fatalf("About resized from %dx%d to %dx%d", lipgloss.Width(before), lipgloss.Height(before), lipgloss.Width(after), lipgloss.Height(after))
	}
	if afterX <= beforeX || afterY <= beforeY {
		t.Fatalf("About position changed from %d,%d to %d,%d, want recentered in larger terminal", beforeX, beforeY, afterX, afterY)
	}
}

func TestNewWithBuildInfoPreservesReleaseMetadataAcrossProfileReset(t *testing.T) {
	info := about.BuildInfo{Version: "1.2.3", Commit: "abc123", Date: "2026-06-14"}
	m := NewWithBuildInfo(nil, info)
	m.resetWorkspaceForProfile()
	if m.buildInfo != info {
		t.Fatalf("build info after profile reset = %#v, want %#v", m.buildInfo, info)
	}
}

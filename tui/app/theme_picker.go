package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/themepicker"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type themeSavedMsg struct {
	palette corestyles.Palette
	err     error
}

func (m Model) openThemePicker() (Model, tea.Cmd, bool) {
	names, listErr := coreconfig.ListThemes()
	resolved, resolveErr := coreconfig.ResolveAppConfig()
	if resolveErr != nil {
		return m, nil, true
	}

	options := make([]themepicker.Option, 0, len(names)+1)
	options = append(options, themepicker.Option{Name: coreconfig.BuiltInThemeName, Palette: corestyles.DefaultPalette()})
	selected := 0
	for _, name := range names {
		resolution, err := coreconfig.LoadTheme(name)
		options = append(options, themepicker.Option{Name: name, Palette: resolution.Palette, Err: err})
		if name == resolved.Effective.Theme && err == nil {
			selected = len(options) - 1
		}
	}

	m.closeOverlays()
	m.themeOriginal = corestyles.CurrentPalette()
	m.themePicker = m.themePicker.SetBounds(0, 0, m.width, m.height).Open(options, selected)
	watchCmd, watchErr := m.startThemeWatcher()
	if openErr := errors.Join(listErr, watchErr); openErr != nil {
		m.themePicker = m.themePicker.SetError(openErr)
	}
	return m, watchCmd, true
}

func (m Model) updateThemePicker(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch watched := msg.(type) {
	case themeWatchChangedMsg:
		if watched.session != m.themeWatcher {
			return m, nil, true
		}
		m.themeWatchRevision++
		revision := m.themeWatchRevision
		return m, tea.Batch(
			waitThemeWatchEvent(watched.session),
			debounceThemeReload(watched.session, revision),
		), true
	case themeWatchReloadMsg:
		if watched.session != m.themeWatcher || watched.revision != m.themeWatchRevision {
			return m, nil, true
		}
		m.reloadCurrentThemePreview()
		return m, nil, true
	case themeWatchErrorMsg:
		if watched.session != m.themeWatcher {
			return m, nil, true
		}
		m.themePicker = m.themePicker.SetError(fmt.Errorf("watch themes: %w", watched.err))
		return m, waitThemeWatchEvent(watched.session), true
	case themeWatchStoppedMsg:
		if watched.session != m.themeWatcher {
			return m, nil, true
		}
		m.themeWatcher = nil
		m.themePicker = m.themePicker.SetError(errors.New("theme watcher stopped"))
		return m, nil, true
	}

	if saved, ok := msg.(themeSavedMsg); ok {
		if saved.err != nil {
			m.themePicker = m.themePicker.SetError(saved.err)
			return m, nil, true
		}
		m.applyThemePreview(saved.palette)
		m.stopThemeWatcher()
		m.themePicker = m.themePicker.Close()
		m.themeOriginal = nil
		return m, nil, true
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
		return m, nil, true
	case tea.KeyMsg:
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
			m.cancelThemePicker()
			return m, m.requestQuit(), true
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionCancel, k),
			tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionThemes, k):
			m.cancelThemePicker()
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionSubmit, k):
			return m, m.saveCurrentTheme(), true
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionUp, k):
			if m.themePicker.Move(-1) {
				m.previewCurrentTheme()
			}
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionDown, k):
			if m.themePicker.Move(1) {
				m.previewCurrentTheme()
			}
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionHome, k):
			if m.themePicker.MoveFirst() {
				m.previewCurrentTheme()
			}
		case tuiconfig.Matches(tuiconfig.BlockThemePicker, tuiconfig.ActionEnd, k):
			if m.themePicker.MoveLast() {
				m.previewCurrentTheme()
			}
		}
		return m, nil, true
	case tea.MouseClickMsg:
		if msg.Mouse().Button != tea.MouseLeft {
			return m, nil, true
		}
		changed, double, hit := m.themePicker.SelectOptionAt(msg.Mouse().X, msg.Mouse().Y, time.Now())
		if !hit {
			return m, nil, true
		}
		if changed {
			m.previewCurrentTheme()
		}
		if double {
			return m, m.saveCurrentTheme(), true
		}
		return m, nil, true
	case tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m *Model) previewCurrentTheme() {
	option, ok := m.themePicker.Current()
	if !ok || option.Err != nil || option.Palette == nil {
		return
	}
	m.applyThemePreview(option.Palette)
}

func (m *Model) reloadCurrentThemePreview() {
	option, ok := m.themePicker.Current()
	if !ok || option.Name == coreconfig.BuiltInThemeName {
		return
	}
	resolution, err := coreconfig.LoadTheme(option.Name)
	if err != nil {
		m.themePicker = m.themePicker.SetError(fmt.Errorf("reload theme %q: %w", option.Name, err))
		return
	}
	m.themePicker = m.themePicker.SetCurrentPalette(resolution.Palette)
	m.applyThemePreview(resolution.Palette)
}

func (m *Model) applyThemePreview(palette corestyles.Palette) {
	corestyles.ApplyPalette(palette)
	corelog.RefreshStyles()
	m.projects = m.projects.RefreshTheme()
	m.details = m.details.RefreshTheme()
	m.logs = m.logs.RefreshTheme()
}

func (m *Model) cancelThemePicker() {
	if !m.themePicker.IsOpen() {
		m.stopThemeWatcher()
		return
	}
	if m.themeOriginal != nil {
		m.applyThemePreview(m.themeOriginal)
	}
	m.themePicker = m.themePicker.Close()
	m.themeOriginal = nil
	m.stopThemeWatcher()
}

func (m *Model) saveCurrentTheme() tea.Cmd {
	if m.themePicker.Saving() {
		return nil
	}
	option, ok := m.themePicker.Current()
	if !ok {
		return nil
	}
	if option.Err != nil {
		m.themePicker = m.themePicker.SetError(option.Err)
		return nil
	}
	m.themePicker = m.themePicker.SetSaving(true)
	name := option.Name
	return func() tea.Msg {
		if err := persistEffectiveTheme(name); err != nil {
			return themeSavedMsg{err: err}
		}
		resolved, err := coreconfig.ResolveConfiguredTheme()
		if err != nil {
			return themeSavedMsg{err: fmt.Errorf("reload selected theme: %w", err)}
		}
		return themeSavedMsg{palette: resolved.Palette}
	}
}

// persistEffectiveTheme stores custom themes in the layer currently providing
// the effective selection. Choosing built-in clears both configured layers so
// the committed result matches the palette that was previewed.
func persistEffectiveTheme(name string) error {
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return err
	}
	if name == coreconfig.BuiltInThemeName {
		if resolved.Local.Config.Theme != "" {
			if err := coreconfig.ResetConfiguredTheme(coreconfig.ThemeScopeLocal); err != nil {
				return err
			}
		}
		if resolved.Global.Config.Theme != "" {
			if err := coreconfig.ResetConfiguredTheme(coreconfig.ThemeScopeGlobal); err != nil {
				return err
			}
		}
		return nil
	}
	scope := coreconfig.ThemeScopeGlobal
	if resolved.Local.Config.Theme != "" {
		scope = coreconfig.ThemeScopeLocal
	}
	return coreconfig.SetConfiguredTheme(name, scope)
}

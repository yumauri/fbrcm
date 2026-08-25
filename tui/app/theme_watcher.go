package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

const themeWatchDebounce = 200 * time.Millisecond

type themeWatchSession struct {
	directory string
	events    <-chan fsnotify.Event
	errors    <-chan error
	closeFn   func() error
	closeOnce sync.Once
}

func (s *themeWatchSession) close() {
	if s == nil || s.closeFn == nil {
		return
	}
	s.closeOnce.Do(func() { _ = s.closeFn() })
}

type themeWatchChangedMsg struct{ session *themeWatchSession }

type themeWatchErrorMsg struct {
	session *themeWatchSession
	err     error
}

type themeWatchStoppedMsg struct{ session *themeWatchSession }

type themeWatchReloadMsg struct {
	session  *themeWatchSession
	revision uint64
}

func (m *Model) startThemeWatcher() (tea.Cmd, error) {
	m.stopThemeWatcher()
	directory := coreconfig.GetThemesDirPath()
	info, err := os.Stat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect themes directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("themes path %q is not a directory", directory)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create theme watcher: %w", err)
	}
	if err := watcher.Add(directory); err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("watch themes directory: %w", err)
	}

	session := &themeWatchSession{
		directory: filepath.Clean(directory),
		events:    watcher.Events,
		errors:    watcher.Errors,
		closeFn:   watcher.Close,
	}
	m.themeWatcher = session
	m.themeWatchRevision = 0
	return waitThemeWatchEvent(session), nil
}

func (m *Model) stopThemeWatcher() {
	if m.themeWatcher != nil {
		m.themeWatcher.close()
	}
	m.themeWatcher = nil
	m.themeWatchRevision++
}

func waitThemeWatchEvent(session *themeWatchSession) tea.Cmd {
	return func() tea.Msg {
		events := session.events
		errorsCh := session.errors
		for events != nil || errorsCh != nil {
			select {
			case event, ok := <-events:
				if !ok {
					events = nil
					continue
				}
				if relevantThemeWatchEvent(session.directory, event) {
					return themeWatchChangedMsg{session: session}
				}
			case err, ok := <-errorsCh:
				if !ok {
					errorsCh = nil
					continue
				}
				return themeWatchErrorMsg{session: session, err: err}
			}
		}
		return themeWatchStoppedMsg{session: session}
	}
}

func relevantThemeWatchEvent(directory string, event fsnotify.Event) bool {
	if filepath.Clean(filepath.Dir(event.Name)) != directory || filepath.Ext(event.Name) != ".toml" {
		return false
	}
	return event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove)
}

func debounceThemeReload(session *themeWatchSession, revision uint64) tea.Cmd {
	return tea.Tick(themeWatchDebounce, func(time.Time) tea.Msg {
		return themeWatchReloadMsg{session: session, revision: revision}
	})
}

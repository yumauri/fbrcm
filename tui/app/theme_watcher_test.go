package app

import (
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestRelevantThemeWatchEvent(t *testing.T) {
	directory := filepath.Clean(t.TempDir())
	tests := []struct {
		name  string
		event fsnotify.Event
		want  bool
	}{
		{name: "write", event: fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Write}, want: true},
		{name: "create", event: fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Create}, want: true},
		{name: "rename", event: fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Rename}, want: true},
		{name: "remove", event: fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Remove}, want: true},
		{name: "uppercase extension", event: fsnotify.Event{Name: filepath.Join(directory, "theme.TOML"), Op: fsnotify.Write}},
		{name: "chmod", event: fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Chmod}},
		{name: "other file", event: fsnotify.Event{Name: filepath.Join(directory, "notes.txt"), Op: fsnotify.Write}},
		{name: "nested file", event: fsnotify.Event{Name: filepath.Join(directory, "pack", "theme.toml"), Op: fsnotify.Write}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relevantThemeWatchEvent(directory, tt.event); got != tt.want {
				t.Fatalf("relevantThemeWatchEvent() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestWaitThemeWatchEventIgnoresUnrelatedFiles(t *testing.T) {
	directory := filepath.Clean(t.TempDir())
	events := make(chan fsnotify.Event, 2)
	errorsCh := make(chan error)
	session := &themeWatchSession{directory: directory, events: events, errors: errorsCh}
	events <- fsnotify.Event{Name: filepath.Join(directory, "notes.txt"), Op: fsnotify.Write}
	events <- fsnotify.Event{Name: filepath.Join(directory, "theme.toml"), Op: fsnotify.Rename}

	msg := waitThemeWatchEvent(session)()
	changed, ok := msg.(themeWatchChangedMsg)
	if !ok || changed.session != session {
		t.Fatalf("watch message = %#v, want changed message for session", msg)
	}
}

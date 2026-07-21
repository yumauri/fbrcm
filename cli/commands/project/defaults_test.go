package project

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestDefaultsCommandWritesSelectedFormatToStdout(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{Name: "Demo", ProjectID: "demo-project", AuthID: "main"}})
	var gotProject string
	var gotFormat firebase.DefaultsFormat
	cmd := newDefaultsCommandWithDownloader(svc, func(_ context.Context, projectID string, format firebase.DefaultsFormat) ([]byte, error) {
		gotProject, gotFormat = projectID, format
		return []byte("<defaults/>\n"), nil
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"demo", "--format", "xml"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if gotProject != "demo-project" || gotFormat != firebase.DefaultsFormatXML {
		t.Fatalf("download args = %q, %q", gotProject, gotFormat)
	}
	if out.String() != "<defaults/>\n" {
		t.Fatalf("stdout = %q", out.String())
	}
}

func TestDefaultsCommandCreatesPrivateFileAndOverwritesWithYes(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{Name: "Demo", ProjectID: "demo-project", AuthID: "main"}})
	downloads := 0
	download := func(_ context.Context, _ string, _ firebase.DefaultsFormat) ([]byte, error) {
		downloads++
		return []byte(`{"flag":"on"}`), nil
	}
	destination := filepath.Join(t.TempDir(), "nested", "defaults.json")

	cmd := newDefaultsCommandWithDownloader(svc, download)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"demo", "--to", destination})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"flag":"on"}` {
		t.Fatalf("file = %q", data)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != config.PrivateFileMode {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}

	cmd = newDefaultsCommandWithDownloader(svc, download)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"demo", "--to", destination, "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if downloads != 2 {
		t.Fatalf("downloads = %d, want 2", downloads)
	}
}

func TestDefaultsCommandRejectsFormatBeforeDownload(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{Name: "Demo", ProjectID: "demo-project", AuthID: "main"}})
	cmd := newDefaultsCommandWithDownloader(svc, func(context.Context, string, firebase.DefaultsFormat) ([]byte, error) {
		return nil, errors.New("unexpected download")
	})
	cmd.SetArgs([]string{"demo", "--format", "yaml"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "allowed: json, xml, plist") {
		t.Fatalf("format error = %v", err)
	}
}

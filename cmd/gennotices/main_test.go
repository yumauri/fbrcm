package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderNoticesIsDeterministic(t *testing.T) {
	components := []component{
		{
			name:    "example.com/module",
			version: "v1.2.3",
			source:  "example.com/module",
			license: "MIT",
			licenseFiles: []licenseFile{
				{name: "LICENSE", content: []byte("Example license")},
			},
		},
	}

	first := renderNotices(components)
	second := renderNotices(components)
	if !bytes.Equal(first, second) {
		t.Fatal("rendered notices are not deterministic")
	}
	for _, expected := range []string{
		"Component: example.com/module",
		"Version: v1.2.3",
		"License: MIT",
		"--- LICENSE ---",
		"Example license",
	} {
		if !bytes.Contains(first, []byte(expected)) {
			t.Errorf("rendered notices do not contain %q", expected)
		}
	}
}

func TestIdentifyCompatibleLicenses(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "MIT", text: "Permission is hereby granted, free of charge", want: "MIT"},
		{name: "MIT-0", text: "MIT No Attribution\nPermission is hereby granted, free of charge", want: "MIT-0"},
		{name: "Apache", text: "Apache License\nVersion 2.0", want: "Apache-2.0"},
		{
			name: "MCP transition",
			text: "licensing transition from the MIT License to the Apache License",
			want: "Apache-2.0 AND MIT",
		},
		{
			name: "BSD-3-Clause",
			text: "Redistribution and use in source and binary forms\nNeither the name may be used",
			want: "BSD-3-Clause",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := identifyLicense([]licenseFile{{name: "LICENSE", content: []byte(test.text)}})
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("license = %q, want %q", got, test.want)
			}
		})
	}
}

func TestIdentifyLicenseRejectsUnreviewedLicense(t *testing.T) {
	_, err := identifyLicense([]licenseFile{{name: "LICENSE", content: []byte("unknown terms")}})
	if err == nil || !strings.Contains(err.Error(), "unrecognized license") {
		t.Fatalf("unreviewed-license error = %v", err)
	}
}

func TestReadLicenseFilesFindsSupportedNamesInStableOrder(t *testing.T) {
	directory := t.TempDir()
	for name, content := range map[string]string{
		"NOTICE.txt":  "notice",
		"LICENSE.md":  "license",
		"PATENTS":     "patents",
		"README.md":   "ignored",
		"COPYING.txt": "copying",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	files, err := readLicenseFiles(directory)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(files))
	for _, file := range files {
		got = append(got, file.name)
	}
	want := []string{"COPYING.txt", "LICENSE.md", "NOTICE.txt", "PATENTS"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("license files = %v, want %v", got, want)
	}
}

func TestReadLicenseFilesRejectsMissingLicense(t *testing.T) {
	_, err := readLicenseFiles(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no LICENSE") {
		t.Fatalf("missing-license error = %v", err)
	}
}

func TestCheckNoticesDetectsStaleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "THIRD_PARTY_NOTICES")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkNotices(path, []byte("new")); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale-notice error = %v", err)
	}
}

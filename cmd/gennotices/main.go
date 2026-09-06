// Command gennotices generates the third-party notices distributed with fbrcm.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const defaultOutputPath = "THIRD_PARTY_NOTICES"

var releaseTargets = []target{
	{goos: "darwin", goarch: "amd64"},
	{goos: "darwin", goarch: "arm64"},
	{goos: "linux", goarch: "amd64"},
	{goos: "linux", goarch: "arm64"},
	{goos: "windows", goarch: "amd64"},
	{goos: "windows", goarch: "arm64"},
}

type target struct {
	goos   string
	goarch string
}

type listedPackage struct {
	Module *listedModule
}

type listedModule struct {
	Path    string
	Version string
	Dir     string
	Main    bool
	Replace *listedModule
}

type component struct {
	name         string
	version      string
	source       string
	license      string
	licenseFiles []licenseFile
}

type licenseFile struct {
	name    string
	content []byte
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "gennotices: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("gennotices", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	check := flags.Bool("check", false, "verify that the notice file is current")
	output := flags.String("output", defaultOutputPath, "generated notice path")
	tags := flags.String("tags", "", "comma-separated release build tags")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}

	components, err := collectComponents(*tags)
	if err != nil {
		return err
	}
	notices := renderNotices(components)
	if *check {
		return checkNotices(*output, notices)
	}
	if err := writeFileAtomically(*output, notices); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "generated notices for %d components at %s\n", len(components), *output)
	return nil
}

func collectComponents(tags string) ([]component, error) {
	modules := make(map[string]listedModule)
	for _, buildTarget := range releaseTargets {
		listed, err := listModules(buildTarget, tags)
		if err != nil {
			return nil, err
		}
		for _, module := range listed {
			modules[module.Path+"@"+module.Version] = module
		}
	}

	goComponent, err := goRuntimeComponent()
	if err != nil {
		return nil, err
	}
	components := []component{goComponent}
	keys := make([]string, 0, len(modules))
	for key := range modules {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		module := modules[key]
		files, err := readLicenseFiles(module.Dir)
		if err != nil {
			return nil, fmt.Errorf("reading licenses for %s@%s: %w", module.Path, module.Version, err)
		}
		license, err := identifyLicense(files)
		if err != nil {
			return nil, fmt.Errorf("classifying license for %s@%s: %w", module.Path, module.Version, err)
		}
		components = append(components, component{
			name:         module.Path,
			version:      module.Version,
			source:       module.Path,
			license:      license,
			licenseFiles: files,
		})
	}
	return components, nil
}

func listModules(buildTarget target, tags string) ([]listedModule, error) {
	arguments := []string{"list", "-deps", "-json"}
	if tags != "" {
		arguments = append(arguments, "-tags", tags)
	}
	arguments = append(arguments, ".")
	command := exec.Command("go", arguments...)
	command.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+buildTarget.goos,
		"GOARCH="+buildTarget.goarch,
	)
	output, err := command.Output()
	if err != nil {
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			return nil, fmt.Errorf("listing packages for %s/%s: %s", buildTarget.goos, buildTarget.goarch, strings.TrimSpace(string(exitError.Stderr)))
		}
		return nil, fmt.Errorf("listing packages for %s/%s: %w", buildTarget.goos, buildTarget.goarch, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	modules := make(map[string]listedModule)
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding package list for %s/%s: %w", buildTarget.goos, buildTarget.goarch, err)
		}
		if pkg.Module == nil || pkg.Module.Main {
			continue
		}
		module := *pkg.Module
		if module.Replace != nil {
			replacement := *module.Replace
			if replacement.Path == "" {
				replacement.Path = module.Path
			}
			if replacement.Version == "" {
				replacement.Version = module.Version
			}
			module = replacement
		}
		if module.Dir == "" {
			return nil, fmt.Errorf("module %s@%s has no source directory", module.Path, module.Version)
		}
		modules[module.Path+"@"+module.Version] = module
	}

	listed := make([]listedModule, 0, len(modules))
	for _, module := range modules {
		listed = append(listed, module)
	}
	return listed, nil
}

func goRuntimeComponent() (component, error) {
	gorootCommand := exec.Command("go", "env", "GOROOT")
	output, err := gorootCommand.Output()
	if err != nil {
		return component{}, fmt.Errorf("locating GOROOT: %w", err)
	}
	goroot := strings.TrimSpace(string(output))
	files, err := readNamedFiles(goroot, []string{"LICENSE", "PATENTS"})
	if err != nil {
		return component{}, fmt.Errorf("reading Go runtime notices: %w", err)
	}
	return component{
		name:         "The Go Programming Language runtime and standard library",
		version:      runtime.Version(),
		source:       "https://go.dev/",
		license:      "BSD-3-Clause",
		licenseFiles: files,
	}, nil
}

func readLicenseFiles(directory string) ([]licenseFile, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !isLicenseFile(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	slices.SortFunc(names, func(left, right string) int {
		if comparison := strings.Compare(strings.ToUpper(left), strings.ToUpper(right)); comparison != 0 {
			return comparison
		}
		return strings.Compare(left, right)
	})
	if len(names) == 0 {
		return nil, fmt.Errorf("no LICENSE, COPYING, NOTICE, or PATENTS file found in %s", directory)
	}
	return readNamedFiles(directory, names)
}

func isLicenseFile(name string) bool {
	upper := strings.ToUpper(name)
	for _, prefix := range []string{"LICENSE", "COPYING", "NOTICE", "PATENTS"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func readNamedFiles(directory string, names []string) ([]licenseFile, error) {
	files := make([]licenseFile, 0, len(names))
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		files = append(files, licenseFile{name: name, content: bytes.TrimSpace(content)})
	}
	return files, nil
}

func identifyLicense(files []licenseFile) (string, error) {
	var licenseText strings.Builder
	for _, file := range files {
		upperName := strings.ToUpper(file.name)
		if strings.HasPrefix(upperName, "LICENSE") || strings.HasPrefix(upperName, "COPYING") {
			licenseText.Write(file.content)
			licenseText.WriteByte('\n')
		}
	}
	text := licenseText.String()
	switch {
	case strings.Contains(text, "licensing transition from the MIT License to the Apache License"):
		return "Apache-2.0 AND MIT", nil
	case strings.Contains(text, "MIT No Attribution"):
		return "MIT-0", nil
	case strings.Contains(text, "Apache License") && strings.Contains(text, "Version 2.0"):
		return "Apache-2.0", nil
	case strings.Contains(text, "Permission is hereby granted, free of charge"):
		return "MIT", nil
	case strings.Contains(text, "Redistribution and use in source and binary forms") &&
		(strings.Contains(strings.ToLower(text), "neither the name") ||
			strings.Contains(text, "may not be used to endorse or promote")):
		return "BSD-3-Clause", nil
	default:
		return "", errors.New("unrecognized license; review it and explicitly extend the compatible-license policy")
	}
}

func renderNotices(components []component) []byte {
	var output strings.Builder
	output.WriteString("fbrcm third-party notices\n")
	output.WriteString("===========================\n\n")
	output.WriteString("This distribution includes third-party software. fbrcm's own source code is\n")
	output.WriteString("licensed under the MIT License in the accompanying LICENSE file. The components\n")
	output.WriteString("below remain subject to their respective terms.\n")
	for _, item := range components {
		output.WriteString("\n\n======================================================================\n")
		fmt.Fprintf(&output, "Component: %s\n", item.name)
		fmt.Fprintf(&output, "Version: %s\n", item.version)
		fmt.Fprintf(&output, "License: %s\n", item.license)
		fmt.Fprintf(&output, "Source: %s\n", item.source)
		for _, file := range item.licenseFiles {
			fmt.Fprintf(&output, "\n--- %s ---\n\n", file.name)
			output.Write(file.content)
			output.WriteByte('\n')
		}
	}
	return []byte(output.String())
}

func checkNotices(path string, expected []byte) error {
	actual, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s does not exist; run go run ./cmd/gennotices", path)
		}
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("%s is stale; run go run ./cmd/gennotices", path)
	}
	fmt.Fprintf(os.Stderr, "verified current notices at %s\n", path)
	return nil
}

func writeFileAtomically(path string, content []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("creating notice directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".third-party-notices-*")
	if err != nil {
		return fmt.Errorf("creating temporary notice: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("writing temporary notice: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("setting notice permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing temporary notice: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("replacing notice: %w", removeErr)
		}
		if renameErr := os.Rename(temporaryPath, path); renameErr != nil {
			return fmt.Errorf("installing notice: %w", renameErr)
		}
	}
	return nil
}

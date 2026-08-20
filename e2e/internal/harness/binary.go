package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Command struct {
	Path                string
	PrefixArgs          []string
	Directory           string
	EnvironmentOverride map[string]string
}

type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

func ResolveCLI(ctx context.Context, e2eRoot, binaryPath, goRunPath, outputDirectory string) (Command, error) {
	if binaryPath != "" && goRunPath != "" {
		return Command{}, fmt.Errorf("-binary and -go-run are mutually exclusive")
	}
	if binaryPath != "" {
		absolute, err := filepath.Abs(binaryPath)
		if err != nil {
			return Command{}, fmt.Errorf("resolve binary path: %w", err)
		}
		if _, err := os.Stat(absolute); err != nil {
			return Command{}, fmt.Errorf("stat binary: %w", err)
		}
		return Command{Path: absolute}, nil
	}
	if goRunPath != "" {
		goEnvironment, err := resolvedGoEnvironment(ctx)
		if err != nil {
			return Command{}, err
		}
		absolute, err := filepath.Abs(goRunPath)
		if err != nil {
			return Command{}, fmt.Errorf("resolve go-run path: %w", err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return Command{}, fmt.Errorf("stat go-run path: %w", err)
		}
		if !info.IsDir() {
			return Command{Path: "go", PrefixArgs: []string{"run", filepath.Base(absolute)}, Directory: filepath.Dir(absolute), EnvironmentOverride: goEnvironment}, nil
		}
		return Command{Path: "go", PrefixArgs: []string{"run", "."}, Directory: absolute, EnvironmentOverride: goEnvironment}, nil
	}

	repoRoot := filepath.Dir(e2eRoot)
	name := "fbrcm"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	output := filepath.Join(outputDirectory, name)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, ".")
	cmd.Dir = repoRoot
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return Command{}, fmt.Errorf("build fbrcm: %w\n%s", err, combined.String())
	}
	return Command{Path: output}, nil
}

func BuildReadGuard(ctx context.Context, e2eRoot, outputDirectory string) (string, error) {
	name := "readguard"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	output := filepath.Join(outputDirectory, name)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/readguard")
	cmd.Dir = e2eRoot
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build read-only middleware: %w\n%s", err, combined.String())
	}
	return output, nil
}

func ResolveHoverfly(ctx context.Context, e2eRoot, outputDirectory string) (string, error) {
	if configured := os.Getenv("FBRCM_E2E_HOVERFLY"); configured != "" {
		absolute, err := filepath.Abs(configured)
		if err != nil {
			return "", fmt.Errorf("resolve Hoverfly path: %w", err)
		}
		if _, err := os.Stat(absolute); err != nil {
			return "", fmt.Errorf("stat Hoverfly binary: %w", err)
		}
		return absolute, nil
	}
	if found, err := exec.LookPath("hoverfly"); err == nil {
		return found, nil
	}
	name := "hoverfly"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	output := filepath.Join(outputDirectory, name)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, "github.com/SpectoLabs/hoverfly/core/cmd/hoverfly")
	cmd.Dir = e2eRoot
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build pinned Hoverfly: %w\n%s", err, combined.String())
	}
	return output, nil
}

func Run(ctx context.Context, command Command, args, environment []string, directory string, stdin []byte) Result {
	allArgs := append(append([]string(nil), command.PrefixArgs...), args...)
	cmd := exec.CommandContext(ctx, command.Path, allArgs...)
	cmd.Dir = directory
	if command.Directory != "" {
		cmd.Dir = command.Directory
	}
	values := environmentMap(environment)
	maps.Copy(values, command.EnvironmentOverride)
	cmd.Env = flattenEnvironment(values)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
			_, _ = fmt.Fprintf(&stderr, "e2e runner: %v\n", err)
		}
	}
	return Result{ExitCode: exitCode, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
}

func resolvedGoEnvironment(ctx context.Context) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "-json", "GOCACHE", "GOMODCACHE", "GOPATH")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("resolve go-run environment: %w\n%s", err, output.String())
	}
	values := make(map[string]string)
	if err := json.Unmarshal(output.Bytes(), &values); err != nil {
		return nil, fmt.Errorf("decode go-run environment: %w", err)
	}
	return values, nil
}

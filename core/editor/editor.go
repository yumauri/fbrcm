// Package editor resolves and builds commands for interactive text editors.
package editor

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/yumauri/fbrcm/core/env"
)

// Resolve returns the first configured editor command. Explicit takes
// precedence over the environment and may be empty.
func Resolve(explicit string) string {
	for _, value := range []string{explicit, os.Getenv(env.Editor), os.Getenv("VISUAL"), os.Getenv("EDITOR")} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	if runtime.GOOS == "windows" {
		return "notepad.exe"
	}
	return "vi"
}

// Command builds a shell command that opens path with editorCommand. Editor
// commands may include arguments such as "code --wait".
func Command(editorCommand, path string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/S", "/C", editorCommand+" "+strconv.Quote(path))
	}
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "/bin/sh"
	}
	return exec.Command(shell, "-c", `exec `+editorCommand+` "$1"`, "fbrcm-editor", path)
}

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

const editorFileEnv = "FBRCM_EDITOR_FILE"

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
	process := exec.Command(shell, "-c", `exec `+editorCommand+` "$`+editorFileEnv+`"`)
	// POSIX shells expose arguments following -c as $0/$1, while fish exposes
	// them through $argv. An environment variable gives both shell families the
	// same safely quoted path expression.
	process.Env = append(os.Environ(), editorFileEnv+"="+path)
	return process
}

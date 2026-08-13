package main

import (
	"errors"
	"testing"

	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestCLIInitializationUsesSemanticExitStatus(t *testing.T) {
	err := &config.InvalidConfigurationError{Path: "config.toml", Stage: "decoding", Err: errors.New("invalid configuration")}
	if got := applicationInitializationExitCode(corelog.ModeCLI, err); got != 3 {
		t.Fatalf("CLI initialization exit code = %d, want 3", got)
	}
	if got := applicationInitializationExitCode(corelog.ModeTUI, err); got != 1 {
		t.Fatalf("TUI initialization exit code = %d, want 1", got)
	}
}

func TestJSONDefaultLogLevelIsSilentUnlessEnvironmentOverridesIt(t *testing.T) {
	t.Setenv(env.LogLevel, "")
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities", "--json"}); got != corelog.SilentLevel {
		t.Fatalf("JSON default log level = %v, want silent", got)
	}
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities"}); got != charmlog.InfoLevel {
		t.Fatalf("human default log level = %v, want info", got)
	}
	t.Setenv(env.LogLevel, "debug")
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities", "--json"}); got != charmlog.InfoLevel {
		t.Fatalf("environment-overridden default = %v, want initialization default info", got)
	}
}

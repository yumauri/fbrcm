package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestDoctorJSONPrintsCompleteReportBeforeFailureExit(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	firebase.SetOfflineMode(true)
	t.Cleanup(func() { firebase.SetOfflineMode(false) })

	cmd := New(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--json"})
	err = cmd.Execute()
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("doctor error = %#v, want exit code 1", err)
	}
	var items []doctorListItem
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if len(items) == 0 || items[0].Profile != config.DefaultProfileName {
		t.Fatalf("doctor items = %+v", items)
	}
	foundFailure := false
	for _, item := range items {
		foundFailure = foundFailure || item.Status == core.DoctorFail
	}
	if !foundFailure {
		t.Fatalf("doctor items have no failed check: %+v", items)
	}
}

func TestDoctorEmptyJSONIsArray(t *testing.T) {
	cmd := newCommand(func(context.Context) core.DoctorReport {
		return core.DoctorReport{Profile: "default"}
	}, passthroughNotifyContext)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "[]" {
		t.Fatalf("empty doctor JSON = %q, want []", got)
	}
}

func TestDoctorHasNoDeadlineByDefault(t *testing.T) {
	var hadDeadline bool
	cmd := newCommand(func(ctx context.Context) core.DoctorReport {
		_, hadDeadline = ctx.Deadline()
		return core.DoctorReport{}
	}, passthroughNotifyContext)
	cmd.SetOut(io.Discard)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor = %v", err)
	}
	if hadDeadline {
		t.Fatal("default doctor context has a deadline")
	}
}

func TestDoctorAppliesExplicitTimeout(t *testing.T) {
	var deadline time.Time
	cmd := newCommand(func(ctx context.Context) core.DoctorReport {
		deadline, _ = ctx.Deadline()
		return core.DoctorReport{}
	}, passthroughNotifyContext)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"--timeout", "1h"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor = %v", err)
	}
	remaining := time.Until(deadline)
	if remaining < 59*time.Minute || remaining > time.Hour {
		t.Fatalf("explicit deadline remaining = %v, want about 1h", remaining)
	}
}

func TestDoctorRejectsNonPositiveExplicitTimeout(t *testing.T) {
	called := false
	cmd := newCommand(func(context.Context) core.DoctorReport {
		called = true
		return core.DoctorReport{}
	}, passthroughNotifyContext)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"--timeout", "0"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("doctor error = %v", err)
	}
	if called {
		t.Fatal("doctor ran with an invalid timeout")
	}
}

func TestDoctorPrintsReportAfterInterruptCancellation(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	t.Setenv("COLUMNS", "60")
	cmd := newCommand(func(ctx context.Context) core.DoctorReport {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("doctor context error = %v, want context canceled", ctx.Err())
		}
		return core.DoctorReport{Checks: []core.DoctorCheck{{
			ID: "profile", Status: core.DoctorPass, Check: "Profile", Detail: "default",
		}}}
	}, func(parent context.Context) (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(parent)
		cancel()
		return ctx, func() {}
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("doctor error = %#v, want exit code 1", err)
	}
	for _, want := range []string{"Status", "Profile", "default", "PASS"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("interrupted output missing %q:\n%s", want, out.String())
		}
	}
}

func passthroughNotifyContext(parent context.Context) (context.Context, context.CancelFunc) {
	return parent, func() {}
}

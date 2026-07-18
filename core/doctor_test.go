package core

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestDoctorReportsAllLocalChecksWhileOffline(t *testing.T) {
	svc := setupCoreTestEnv(t)
	firebase.SetOfflineMode(true)
	t.Cleanup(func() { firebase.SetOfflineMode(false) })

	report := svc.Doctor(context.Background())
	statuses := make(map[string]string, len(report.Checks))
	for _, check := range report.Checks {
		statuses[check.ID] = check.Status
	}
	if statuses["profile"] != DoctorPass {
		t.Fatalf("profile status = %q, checks = %+v", statuses["profile"], report.Checks)
	}
	if statuses["cache-writable"] != DoctorPass {
		t.Fatalf("cache status = %q, checks = %+v", statuses["cache-writable"], report.Checks)
	}
	if statuses["auth-config"] != DoctorFail {
		t.Fatalf("auth config status = %q, want fail", statuses["auth-config"])
	}
	if statuses["network"] != DoctorWarn {
		t.Fatalf("network status = %q, want warn", statuses["network"])
	}
	projectsCheck := doctorCheckByID(t, report, "projects-config")
	if !strings.Contains(projectsCheck.Detail, "fbrcm projects update") || strings.Contains(projectsCheck.Detail, "projects sync") {
		t.Fatalf("projects remediation = %q", projectsCheck.Detail)
	}
	if !report.Failed() {
		t.Fatal("report with missing auth config should fail")
	}
}

func TestDoctorOAuthTokenRefreshUpdatesCachedExpiryStatus(t *testing.T) {
	diagnostic := firebase.AuthDiagnostic{
		TokenExpired: true, HasRefreshToken: true,
		TokenExpiry: time.Date(2026, 7, 17, 14, 21, 17, 0, time.Local),
	}

	report := DoctorReport{}
	report.addOAuthTokenCheck("main", diagnostic)
	check := doctorCheckByID(t, report, "token:main")
	if check.Status != DoctorWarn || !strings.Contains(check.Detail, "refresh pending") {
		t.Fatalf("pending token check = %+v", check)
	}

	report.updateOAuthTokenRefresh("main", diagnostic, nil)
	check = doctorCheckByID(t, report, "token:main")
	if check.Status != DoctorPass || !strings.Contains(check.Detail, "refresh succeeded") {
		t.Fatalf("refreshed token check = %+v", check)
	}

	report.updateOAuthTokenRefresh("main", diagnostic, errors.New("invalid grant"))
	check = doctorCheckByID(t, report, "token:main")
	if check.Status != DoctorFail || !strings.Contains(check.Detail, "refresh failed: invalid grant") {
		t.Fatalf("failed token check = %+v", check)
	}
}

func TestDoctorOAuthTokenRefreshRemainsWarningOffline(t *testing.T) {
	report := DoctorReport{Offline: true}
	report.addOAuthTokenCheck("main", firebase.AuthDiagnostic{TokenExpired: true, HasRefreshToken: true})
	check := doctorCheckByID(t, report, "token:main")
	if check.Status != DoctorWarn || !strings.Contains(check.Detail, "refresh not tested in offline mode") {
		t.Fatalf("offline token check = %+v", check)
	}
}

func doctorCheckByID(t *testing.T, report DoctorReport, id string) DoctorCheck {
	t.Helper()
	for _, check := range report.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("check %q not found in %+v", id, report.Checks)
	return DoctorCheck{}
}

func TestDoctorReportFailedIgnoresWarnings(t *testing.T) {
	report := DoctorReport{Checks: []DoctorCheck{{Status: DoctorPass}, {Status: DoctorWarn}}}
	if report.Failed() {
		t.Fatal("warnings should not fail doctor")
	}
	report.Checks = append(report.Checks, DoctorCheck{Status: DoctorFail})
	if !report.Failed() {
		t.Fatal("failure should fail doctor")
	}
}

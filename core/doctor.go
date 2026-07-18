package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

const (
	DoctorPass = "pass"
	DoctorWarn = "warn"
	DoctorFail = "fail"
)

var requiredFirebasePermissions = []string{
	"cloudconfig.configs.get",
	"cloudconfig.configs.update",
}

// DoctorCheck is one independently actionable application health check.
type DoctorCheck struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Check  string `json:"check"`
	Detail string `json:"detail"`
}

// DoctorReport contains all local and live diagnostic results.
type DoctorReport struct {
	Profile   string        `json:"profile"`
	ConfigDir string        `json:"config_dir"`
	CacheDir  string        `json:"cache_dir"`
	Offline   bool          `json:"offline"`
	Checks    []DoctorCheck `json:"checks"`
}

// Failed reports whether any diagnostic check failed.
func (r DoctorReport) Failed() bool {
	return slices.ContainsFunc(r.Checks, func(check DoctorCheck) bool { return check.Status == DoctorFail })
}

// Doctor runs non-interactive local and live diagnostics. It never starts an
// OAuth login flow and never writes refreshed credentials.
func (s *Core) Doctor(ctx context.Context) DoctorReport {
	if ctx == nil {
		ctx = context.Background()
	}
	report := DoctorReport{Offline: firebase.IsOffline()}

	if err := config.EnsureActiveProfile(); err != nil {
		report.add("profile", DoctorFail, "Profile", err.Error())
	} else {
		report.add("profile", DoctorPass, "Profile", config.GetActiveProfileName())
	}
	report.Profile = config.GetActiveProfileName()
	report.ConfigDir = config.GetConfigDirPath()
	report.CacheDir = config.GetCacheDirPath()
	report.checkDirectory("config-directory", "Profile config directory", report.ConfigDir, false)
	report.checkDirectory("cache-writable", "Profile cache writable", report.CacheDir, true)

	authFile, authErr := config.LoadAuth()
	if authErr != nil {
		report.add("auth-config", DoctorFail, "Auth config", authErr.Error())
	} else if len(authFile.Auth) == 0 {
		report.add("auth-config", DoctorFail, "Auth config", "no auth identities configured")
	} else {
		report.add("auth-config", DoctorPass, "Auth config", fmt.Sprintf("%d identities; default %s", len(authFile.Auth), authFile.DefaultAuthID))
	}

	localAuthOK := make(map[string]bool)
	authDiagnostics := make(map[string]firebase.AuthDiagnostic)
	if authErr == nil {
		for _, auth := range authFile.Auth {
			diagnostic, err := firebase.InspectAuth(auth)
			checkName := fmt.Sprintf("Credentials (%s)", auth.ID)
			if err != nil {
				report.add("credentials:"+auth.ID, DoctorFail, checkName, err.Error())
				continue
			}
			localAuthOK[auth.ID] = true
			authDiagnostics[auth.ID] = diagnostic
			detail := auth.Type
			if diagnostic.CredentialPath != "" {
				detail += ": " + diagnostic.CredentialPath
			}
			status := DoctorPass
			if diagnostic.CredentialWarning != "" {
				status = DoctorWarn
				detail += "; " + diagnostic.CredentialWarning
			}
			report.add("credentials:"+auth.ID, status, checkName, detail)
			if auth.Type == config.AuthTypeOAuth {
				report.addOAuthTokenCheck(auth.ID, diagnostic)
			}
		}
	}

	projects, projectsErr := config.LoadProjects()
	if projectsErr != nil {
		if errors.Is(projectsErr, os.ErrNotExist) || errors.Is(projectsErr, config.ErrEmptyProjectsFile) {
			report.add("projects-config", DoctorWarn, "Projects config", "not cached; run `fbrcm projects update`")
		} else {
			report.add("projects-config", DoctorFail, "Projects config", projectsErr.Error())
		}
	} else {
		report.add("projects-config", DoctorPass, "Projects config", fmt.Sprintf("%d projects", len(projects)))
	}

	if report.Offline {
		report.add("network", DoctorWarn, "Network", "offline mode enabled; live API and permission checks skipped")
		return report
	}
	report.add("network", DoctorPass, "Network", "online")
	if authErr != nil {
		return report
	}
	if projectsErr != nil || len(projects) == 0 {
		report.add("api:remote-config", DoctorWarn, "Firebase Remote Config API", "skipped: no cached projects")
		report.add("permissions", DoctorWarn, "Firebase permissions", "skipped: no cached projects")
	}

	for _, auth := range authFile.Auth {
		if report.addContextFailure(ctx) {
			return report
		}
		if !localAuthOK[auth.ID] {
			continue
		}
		fb, err := firebase.NewDiagnosticServiceForAuth(ctx, auth)
		if auth.Type == config.AuthTypeOAuth && (err == nil || ctx.Err() == nil) {
			report.updateOAuthTokenRefresh(auth.ID, authDiagnostics[auth.ID], err)
		}
		if err != nil {
			report.add("token-refresh:"+auth.ID, DoctorFail, fmt.Sprintf("Authentication (%s)", auth.ID), err.Error())
			if report.addContextFailure(ctx) {
				return report
			}
			continue
		}
		if _, err := fb.ListProjects(ctx); err != nil {
			report.add("api:cloud-resource-manager:"+auth.ID, DoctorFail, fmt.Sprintf("Cloud Resource Manager API (%s)", auth.ID), err.Error())
			if report.addContextFailure(ctx) {
				return report
			}
		} else {
			report.add("api:cloud-resource-manager:"+auth.ID, DoctorPass, fmt.Sprintf("Cloud Resource Manager API (%s)", auth.ID), "project listing allowed")
		}
		if projectsErr != nil {
			continue
		}
		for _, project := range projects {
			if report.addContextFailure(ctx) {
				return report
			}
			if project.AuthID != auth.ID {
				continue
			}
			report.checkFirebaseProject(ctx, fb, project)
			if report.addContextFailure(ctx) {
				return report
			}
		}
	}
	return report
}

func (r *DoctorReport) add(id, status, check, detail string) {
	r.Checks = append(r.Checks, DoctorCheck{ID: id, Status: status, Check: check, Detail: detail})
}

func (r *DoctorReport) addContextFailure(ctx context.Context) bool {
	if ctx == nil || ctx.Err() == nil {
		return false
	}
	if !slices.ContainsFunc(r.Checks, func(check DoctorCheck) bool { return check.ID == "run-context" }) {
		r.add("run-context", DoctorFail, "Diagnostic run", ctx.Err().Error())
	}
	return true
}

func (r *DoctorReport) checkDirectory(id, check, path string, writable bool) {
	info, err := os.Stat(path)
	if err != nil {
		r.add(id, DoctorFail, check, err.Error())
		return
	}
	if !info.IsDir() {
		r.add(id, DoctorFail, check, "not a directory: "+path)
		return
	}
	if !writable {
		r.add(id, DoctorPass, check, path)
		return
	}
	file, err := os.CreateTemp(path, ".fbrcm-doctor-*")
	if err != nil {
		r.add(id, DoctorFail, check, err.Error())
		return
	}
	name := file.Name()
	closeErr := file.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		r.add(id, DoctorFail, check, closeErr.Error())
		return
	}
	if removeErr != nil {
		r.add(id, DoctorFail, check, removeErr.Error())
		return
	}
	r.add(id, DoctorPass, check, path)
}

func (r *DoctorReport) addOAuthTokenCheck(authID string, diagnostic firebase.AuthDiagnostic) {
	name := fmt.Sprintf("OAuth token (%s)", authID)
	if diagnostic.TokenError != "" {
		r.add("token:"+authID, DoctorFail, name, diagnostic.TokenError+": "+diagnostic.TokenPath)
		return
	}
	if diagnostic.TokenExpired {
		status := DoctorFail
		detail := "cached access token expired"
		if !diagnostic.TokenExpiry.IsZero() {
			detail += " at " + diagnostic.TokenExpiry.Local().Format("2006-01-02 15:04:05 MST")
		}
		if diagnostic.HasRefreshToken {
			status = DoctorWarn
			detail += "; refresh token available"
			if r.Offline {
				detail += "; refresh not tested in offline mode"
			} else {
				detail += "; refresh pending"
			}
		}
		r.add("token:"+authID, status, name, detail)
		return
	}
	detail := "valid"
	if !diagnostic.TokenExpiry.IsZero() {
		detail += " until " + diagnostic.TokenExpiry.Local().Format("2006-01-02 15:04:05 MST")
	}
	r.add("token:"+authID, DoctorPass, name, detail)
}

func (r *DoctorReport) updateOAuthTokenRefresh(authID string, diagnostic firebase.AuthDiagnostic, refreshErr error) {
	if diagnostic.TokenError != "" || !diagnostic.TokenExpired || !diagnostic.HasRefreshToken {
		return
	}
	detail := "cached access token expired"
	if !diagnostic.TokenExpiry.IsZero() {
		detail += " at " + diagnostic.TokenExpiry.Local().Format("2006-01-02 15:04:05 MST")
	}
	status := DoctorPass
	detail += "; refresh succeeded"
	if refreshErr != nil {
		status = DoctorFail
		detail = strings.TrimSuffix(detail, "; refresh succeeded") + "; refresh failed: " + refreshErr.Error()
	}
	for i := range r.Checks {
		if r.Checks[i].ID == "token:"+authID {
			r.Checks[i].Status = status
			r.Checks[i].Detail = detail
			return
		}
	}
}

func (r *DoctorReport) checkFirebaseProject(ctx context.Context, fb *firebase.Service, project Project) {
	checkSuffix := project.ProjectID
	if _, _, err := fb.GetRemoteConfig(ctx, project.ProjectID); err != nil {
		r.add("api:remote-config:"+checkSuffix, DoctorFail, "Firebase Remote Config API ("+checkSuffix+")", err.Error())
	} else {
		r.add("api:remote-config:"+checkSuffix, DoctorPass, "Firebase Remote Config API ("+checkSuffix+")", "read allowed")
	}
	if ctx.Err() != nil {
		return
	}
	granted, err := fb.TestProjectPermissions(ctx, project.ProjectID, requiredFirebasePermissions)
	if err != nil {
		r.add("permissions:"+checkSuffix, DoctorFail, "Firebase permissions ("+checkSuffix+")", err.Error())
		return
	}
	missing := make([]string, 0)
	for _, required := range requiredFirebasePermissions {
		if !slices.Contains(granted, required) {
			missing = append(missing, required)
		}
	}
	if len(missing) > 0 {
		r.add("permissions:"+checkSuffix, DoctorFail, "Firebase permissions ("+checkSuffix+")", "missing: "+strings.Join(missing, ", "))
		return
	}
	r.add("permissions:"+checkSuffix, DoctorPass, "Firebase permissions ("+checkSuffix+")", strings.Join(granted, ", "))
}

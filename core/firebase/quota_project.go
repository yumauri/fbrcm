package firebase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core/config"
	coreenv "github.com/yumauri/fbrcm/core/env"
)

type QuotaProjectSource string

const (
	QuotaProjectSourceEnvironment QuotaProjectSource = "environment"
	QuotaProjectSourceProject     QuotaProjectSource = "project"
	QuotaProjectSourceAuth        QuotaProjectSource = "auth"
	QuotaProjectSourceCredentials QuotaProjectSource = "credentials"
	QuotaProjectSourceTarget      QuotaProjectSource = "target"
	QuotaProjectSourceUnresolved  QuotaProjectSource = "unresolved"
)

// QuotaProjectSelection identifies the effective quota project and its source.
type QuotaProjectSelection struct {
	ProjectID string             `json:"project_id"`
	Source    QuotaProjectSource `json:"source"`
}

// QuotaProjectRequiredError reports that no explicit quota consumer can be
// selected for an authenticated Google API request.
type QuotaProjectRequiredError struct {
	TargetProjectID string
}

// QuotaProjectInvariantError reports a request path that bypassed quota-project
// preparation. It represents an internal programming error.
type QuotaProjectInvariantError struct {
	Method string
	Host   string
}

func (e *QuotaProjectInvariantError) Error() string {
	return fmt.Sprintf("authenticated %s request to %s is missing X-Goog-User-Project", e.Method, e.Host)
}

func (e *QuotaProjectRequiredError) Error() string {
	if e.TargetProjectID == "" {
		return "quota project is required for this targetless Google API request"
	}
	return fmt.Sprintf("quota project is required for Google API request targeting project %s", e.TargetProjectID)
}

// QuotaProjectError identifies malformed quota-project configuration without
// coupling the Firebase layer to CLI error presentation.
type QuotaProjectError struct {
	Source   QuotaProjectSource
	Variable string
	Err      error
}

func (e *QuotaProjectError) Error() string { return e.Err.Error() }
func (e *QuotaProjectError) Unwrap() error { return e.Err }

type quotaProjectPolicy struct {
	environmentQuotaProjectID string
	projectQuotaProjectID     string
	authQuotaProjectID        string
	credentialQuotaProjectID  string
	useTargetProjectQuota     bool
}

func (p quotaProjectPolicy) selectProject(targetProjectID string) (QuotaProjectSelection, error) {
	if p.environmentQuotaProjectID != "" {
		return QuotaProjectSelection{ProjectID: p.environmentQuotaProjectID, Source: QuotaProjectSourceEnvironment}, nil
	}
	if p.projectQuotaProjectID != "" {
		return QuotaProjectSelection{ProjectID: p.projectQuotaProjectID, Source: QuotaProjectSourceProject}, nil
	}
	if p.authQuotaProjectID != "" {
		return QuotaProjectSelection{ProjectID: p.authQuotaProjectID, Source: QuotaProjectSourceAuth}, nil
	}
	if p.credentialQuotaProjectID != "" {
		return QuotaProjectSelection{ProjectID: p.credentialQuotaProjectID, Source: QuotaProjectSourceCredentials}, nil
	}
	if p.useTargetProjectQuota {
		if target := strings.TrimSpace(targetProjectID); target != "" {
			return QuotaProjectSelection{ProjectID: target, Source: QuotaProjectSourceTarget}, nil
		}
	}
	return QuotaProjectSelection{Source: QuotaProjectSourceUnresolved}, &QuotaProjectRequiredError{TargetProjectID: strings.TrimSpace(targetProjectID)}
}

func environmentQuotaProjectID() (string, error) {
	value, ok := coreenv.LookupTrimmed(coreenv.GoogleCloudQuotaProject)
	if !ok {
		return "", nil
	}
	if err := config.ValidateQuotaProjectID(value); err != nil {
		return "", &QuotaProjectError{
			Source:   QuotaProjectSourceEnvironment,
			Variable: coreenv.GoogleCloudQuotaProject,
			Err:      fmt.Errorf("invalid %s: %w", coreenv.GoogleCloudQuotaProject, err),
		}
	}
	return value, nil
}

func credentialQuotaProjectID(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	var payload struct {
		QuotaProjectID string `json:"quota_project_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", &QuotaProjectError{
			Source: QuotaProjectSourceCredentials,
			Err:    fmt.Errorf("parse ADC quota project: %w", err),
		}
	}
	value := strings.TrimSpace(payload.QuotaProjectID)
	if value == "" {
		return "", nil
	}
	if err := config.ValidateQuotaProjectID(value); err != nil {
		return "", &QuotaProjectError{
			Source: QuotaProjectSourceCredentials,
			Err:    fmt.Errorf("invalid ADC quota_project_id: %w", err),
		}
	}
	return value, nil
}

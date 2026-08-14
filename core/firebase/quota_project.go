package firebase

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	coreenv "github.com/yumauri/fbrcm/core/env"
)

type QuotaProjectSource string

const (
	QuotaProjectSourceEnvironment QuotaProjectSource = "environment"
	QuotaProjectSourceCredentials QuotaProjectSource = "credentials"
)

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
	credentialQuotaProjectID  string
	useTargetProjectQuota     bool
}

func (p quotaProjectPolicy) projectID(targetProjectID string) string {
	if p.environmentQuotaProjectID != "" {
		return p.environmentQuotaProjectID
	}
	if p.credentialQuotaProjectID != "" {
		return p.credentialQuotaProjectID
	}
	if p.useTargetProjectQuota {
		return strings.TrimSpace(targetProjectID)
	}
	return ""
}

func environmentQuotaProjectID() (string, error) {
	value, ok := coreenv.LookupTrimmed(coreenv.GoogleCloudQuotaProject)
	if !ok {
		return "", nil
	}
	if err := validateQuotaProjectID(value); err != nil {
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
	if err := validateQuotaProjectID(value); err != nil {
		return "", &QuotaProjectError{
			Source: QuotaProjectSourceCredentials,
			Err:    fmt.Errorf("invalid ADC quota_project_id: %w", err),
		}
	}
	return value, nil
}

func validateQuotaProjectID(value string) error {
	if value == "" {
		return fmt.Errorf("project identifier must not be empty")
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("project identifier must be valid UTF-8")
	}
	if strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return fmt.Errorf("project identifier must not contain whitespace or control characters")
	}
	return nil
}

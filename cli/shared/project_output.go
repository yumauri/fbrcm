package shared

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

// ProjectJSON is the shared machine-readable representation of a project.
type ProjectJSON struct {
	Project                  string                      `json:"project"`
	ProjectID                string                      `json:"project_id"`
	Aliases                  []string                    `json:"aliases"`
	Number                   string                      `json:"number,omitempty"`
	State                    string                      `json:"state,omitempty"`
	ETag                     string                      `json:"etag,omitempty"`
	AuthID                   string                      `json:"auth_id"`
	ConfiguredQuotaProjectID string                      `json:"configured_quota_project_id,omitempty"`
	EffectiveQuotaProjectID  string                      `json:"effective_quota_project_id,omitempty"`
	QuotaProjectSource       firebase.QuotaProjectSource `json:"quota_project_source,omitempty" contract:"enum=environment|project|auth|credentials|target"`
	Disabled                 bool                        `json:"disabled"`
	DiscoveredBy             []string                    `json:"discovered_by,omitempty"`
	Templates                []rctarget.Kind             `json:"templates"`
	PrimaryTemplate          rctarget.Kind               `json:"primary_template"`
	UpdatedAt                string                      `json:"updated_at,omitempty"`
	SyncedAt                 string                      `json:"synced_at,omitempty"`
	URL                      string                      `json:"url,omitempty" contract:"format=uri"`
}

// NewProjectJSON copies a project into its CLI JSON representation.
func NewProjectJSON(project core.Project, withURL bool) ProjectJSON {
	return NewProjectJSONWithAliases(project, nil, withURL)
}

// NewProjectJSONWithAliases includes sorted repository aliases in a project representation.
func NewProjectJSONWithAliases(project core.Project, aliases []string, withURL bool) ProjectJSON {
	_ = project.NormalizeTemplatePreferences()
	out := ProjectJSON{
		Project:                  project.Name,
		ProjectID:                project.ProjectID,
		Aliases:                  append([]string{}, aliases...),
		Number:                   project.ProjectNumber,
		State:                    project.State,
		ETag:                     project.ETag,
		AuthID:                   project.AuthID,
		ConfiguredQuotaProjectID: project.QuotaProjectID,
		Disabled:                 project.Disabled,
		DiscoveredBy:             append([]string(nil), project.DiscoveredBy...),
		Templates:                append([]rctarget.Kind(nil), project.Templates...),
		PrimaryTemplate:          project.PrimaryTemplate,
		UpdatedAt:                project.UpdatedAt,
		SyncedAt:                 project.SyncedAt,
	}
	if withURL {
		out.URL = firebase.RemoteConfigConsoleURL(project.ProjectID)
	}
	return out
}
